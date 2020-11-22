package oms

import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/gorilla/mux"
	"github.com/sethvargo/go-password/password"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	"github.com/omecodes/bome"
	"github.com/omecodes/common/dao/mapping"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/netx"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/libome/v2"
	"github.com/omecodes/service/v2"
)

var debug = os.Getenv("OMESTORE_DEBUG")

// Config contains info to configure an instance of Server
type Config struct {
	Box *service.Box
	DSN string
}

// NewServer is a server constructor
func NewServer(config Config) *Server {
	s := new(Server)
	s.config = &config
	return s
}

// Server embeds an Ome data store
// it also exposes an API server
type Server struct {
	initialized bool

	options []netx.ListenOption
	config  *Config

	adminPassword string

	dataAccessRules mapping.DoubleMap
	key             []byte
	celEnv          *cel.Env
	dataStore       Store

	settings    bome.JSONMap
	workers     bome.JSONMap
	permissions bome.JSONDoubleMap

	dListener net.Listener
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

	s.permissions, err = bome.NewJSONDoubleMap(db, bome.MySQL, "permissions")
	if err != nil {
		return err
	}

	s.settings, err = bome.NewJSONMap(db, bome.MySQL, "settings")
	if err != nil {
		return err
	}

	s.dataStore, err = NewStore(db)
	if err != nil {
		return err
	}

	s.celEnv, err = cel.NewEnv(
		cel.Types(&Auth{}),
		cel.Types(&Object{}),
		cel.Types(&Perm{}),
		cel.Types(&Graft{}),
		cel.Declarations(
			decls.NewVar("at", decls.Int),
			decls.NewVar("auth", decls.NewObjectType("Auth")),
			decls.NewVar("data", decls.NewObjectType("Object")),
			decls.NewVar("graft", decls.NewObjectType("Graft")),
			decls.NewFunction("acl",
				decls.NewOverload(
					"acl",
					[]*expr.Type{decls.String, decls.String}, decls.NewObjectType("Perm"),
				),
			),
		),
	)
	if err != nil {
		return err
	}

	adminPwdFilename := filepath.Join(s.config.Box.Dir(), "admin-pwd")
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

	settings, err := s.settings.Get(Settings)
	if err != nil && !errors.IsNotFound(err) {
		log.Error("could not get db settings")
		return err
	}

	if settings == "" {
		defaultSettings := &bome.MapEntry{
			Key:   Settings,
			Value: defaultSettings,
		}
		return s.settings.Save(defaultSettings)
	}

	return nil
}

func (s *Server) getStoredKey(name string, size int) ([]byte, error) {
	cookiesKeyFilename := filepath.Join(s.config.Box.Dir(), name+".key")
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

	return s.config.Box.StartGateway(&service.GatewayParams{
		ForceRegister: false,
		MiddlewareList: []mux.MiddlewareFunc{
			s.enrichContext,
			s.detectAuthentication,
			httpx.Logger("omestore").Handle,
		},
		Port: 80,
		ProvideRouter: func() *mux.Router {
			return dataRouter()
		},
		Node: &ome.Node{
			Protocol: ome.Protocol_Http,
			Security: ome.Security_Insecure,
		},
	})
}

func (s *Server) enrichContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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

				ctx := ContextWithUserInfo(r.Context(), &Auth{
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

			ctx := ContextWithUserInfo(r.Context(), &Auth{
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
	_ = s.dListener.Close()
}
