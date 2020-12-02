package server

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"github.com/omecodes/common/env/app"
	"github.com/omecodes/omestore/pb"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/gorilla/mux"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/netx"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/libome"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/router"
	"github.com/sethvargo/go-password/password"
)

var debug = os.Getenv("OMS_DEBUG")

// Config contains info to configure an instance of Server
type Config struct {
	App         *app.App
	TLS         *tls.Config
	BindAddress string
	DSN         string
}

// New is a server constructor
func New(config Config) *Server {
	s := new(Server)
	s.config = &config
	return s
}

// Server embeds an Ome data store
// it also exposes an API server
type Server struct {
	initialized   bool
	options       []netx.ListenOption
	config        *Config
	adminPassword string
	key           []byte
	celPolicyEnv  *cel.Env
	celSearchEnv  *cel.Env

	objects     oms.Objects
	workers     *bome.JSONMap
	settings    *bome.Map
	accessStore oms.AccessStore
	listener    net.Listener
	Errors      chan error
	server      *http.Server
}

func (s *Server) init() error {
	if s.initialized {
		return nil
	}
	s.initialized = true

	db, err := sql.Open("mysql", s.config.DSN)
	if err != nil {
		return err
	}

	s.workers, err = bome.NewJSONMap(db, bome.MySQL, "users")
	if err != nil {
		return err
	}

	s.accessStore, err = oms.NewSQLAccessStore(db, bome.MySQL, "accesses")
	if err != nil {
		return err
	}

	s.settings, err = bome.NewMap(db, bome.MySQL, "settings")
	if err != nil {
		return err
	}

	s.objects, err = oms.NewSQLObjects(db, bome.MySQL)
	if err != nil {
		return err
	}

	s.celPolicyEnv, err = cel.NewEnv(
		cel.Declarations(
			decls.NewVar("auth", decls.NewMapType(decls.String, decls.Dyn)),
			decls.NewVar("data", decls.NewMapType(decls.String, decls.Dyn)),
		),
	)
	if err != nil {
		return err
	}

	s.celSearchEnv, err = cel.NewEnv(
		cel.Declarations(decls.NewVar("o", decls.NewMapType(decls.String, decls.Dyn))))
	if err != nil {
		return nil
	}

	adminPwdFilename := filepath.Join(s.config.App.DataDir(), "admin-pwd")
	passwordBytes, err := ioutil.ReadFile(adminPwdFilename)
	if err != nil {
		genPassword, err := password.Generate(16, 5, 11, false, false)
		passwordBytes = []byte(genPassword)
		if err != nil {
			return err
		}
		s.adminPassword = base64.RawStdEncoding.EncodeToString(passwordBytes)
		err = ioutil.WriteFile(adminPwdFilename, []byte(s.adminPassword), os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		s.adminPassword = string(passwordBytes)
	}

	s.key, err = s.getStoredKey("token-key", 64)
	if err != nil {
		return err
	}

	err = s.settings.Save(&bome.MapEntry{
		Key:   oms.SettingsDataMaxSizePath,
		Value: oms.DefaultSettings[oms.SettingsDataMaxSizePath],
	})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	err = s.settings.Save(&bome.MapEntry{
		Key:   oms.SettingsCreateDataSecurityRule,
		Value: oms.DefaultSettings[oms.SettingsCreateDataSecurityRule],
	})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func (s *Server) getStoredKey(name string, size int) ([]byte, error) {
	cookiesKeyFilename := filepath.Join(s.config.App.DataDir(), name+".key")
	key, err := ioutil.ReadFile(cookiesKeyFilename)
	if err != nil {
		key = make([]byte, size)
		_, err = rand.Read(key)
		if err != nil {
			log.Error("could not generate secret key", log.Err(err), log.Field("name", name))
			return nil, err
		}
		err = ioutil.WriteFile(cookiesKeyFilename, key, os.ModePerm)
		if err != nil {
			log.Error("could not save secret key", log.Err(err), log.Field("name", name))
			return nil, err
		}
	}
	return key, nil
}

// Start starts API server
func (s *Server) Start() error {
	err := s.init()
	if err != nil {
		return err
	}

	if s.config.TLS != nil {
		s.listener, err = tls.Listen("tcp", s.config.BindAddress, s.config.TLS)
	} else {
		s.listener, err = net.Listen("tcp", s.config.BindAddress)
	}
	if err != nil {
		return err
	}

	address := s.listener.Addr().String()
	log.Info("starting HTTP server", log.Field("address", address))

	middlewareList := []mux.MiddlewareFunc{
		s.enrichContext,
		s.detectAuthentication,
		s.detectOAuth2Authorization,
		httpx.Logger("omestore").Handle,
	}
	var handler http.Handler
	handler = dataRouter()
	for _, m := range middlewareList {
		handler = m.Middleware(handler)
	}

	go func() {
		s.server = &http.Server{
			Addr:    address,
			Handler: handler,
		}
		if err := s.server.Serve(s.listener); err != nil {
			s.Errors <- err
		}
	}()

	return nil
}

func (s *Server) enrichContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = router.WithAccessStore(s.accessStore)(ctx)
		ctx = router.WithCelPolicyEnv(s.celPolicyEnv)(ctx)
		ctx = router.WithCelSearchEnv(s.celSearchEnv)(ctx)
		ctx = router.WithSettings(nil)(ctx)
		ctx = router.WithObjectsStore(nil)(ctx)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) detectAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ctx := r.Context()
		h := r.Header
		authorization := h.Get("Authorization")

		if authorization != "" {
			splits := strings.SplitN(authorization, " ", 2)
			if strings.ToLower(splits[0]) == "basic" {
				bytes, err := base64.StdEncoding.DecodeString(splits[1])
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				parts := strings.Split(string(bytes), ":")
				if len(parts) != 2 {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				authUser := parts[0]
				var pass string
				if len(parts) > 1 {
					pass = parts[1]
				}

				if authUser == "admin" {
					if pass != s.adminPassword {
						w.WriteHeader(http.StatusForbidden)
						return
					}
				} else {
					secret, err := s.workers.ExtractAt(authUser, "$.secret")
					if err != nil {
						if bome.IsNotFound(err) {
							w.WriteHeader(http.StatusForbidden)
							return
						}
						log.Error("could not get auth user info", log.Err(err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					sh := sha512.New()
					_, err = sh.Write([]byte(pass))
					if err != nil {
						log.Error("could not hash password", log.Err(err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					hashed := sh.Sum(nil)
					if hex.EncodeToString(hashed) != secret {
						w.WriteHeader(http.StatusForbidden)
						return
					}
				}

				ctx := router.WithUserInfo(r.Context(), &pb.Auth{
					Uid:       authUser,
					Validated: true,
					Worker:    "admin" != authUser,
				})
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) detectOAuth2Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("authorization")
		if authorization != "" && strings.HasPrefix(authorization, "Bearer ") {
			authorization = strings.TrimPrefix(authorization, "Bearer ")
			jwt, err := ome.ParseJWT(authorization)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			state, err := jwt.Verify("")
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if state != ome.JWTState_Valid {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			ctx := router.WithUserInfo(r.Context(), &pb.Auth{
				Uid:       jwt.Claims.Sub,
				Email:     jwt.Claims.Profile.Email,
				Worker:    false,
				Validated: jwt.Claims.Profile.Verified,
				Scope:     strings.Split(jwt.Claims.Scope, ""),
			})

			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// Stop stops API server
func (s *Server) Stop() {
	_ = s.listener.Close()
}
