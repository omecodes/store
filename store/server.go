package store

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/omecodes/common/dao/mapping"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/netx"
	"github.com/omecodes/common/utils/log"
	ome "github.com/omecodes/libome"
	authpb "github.com/omecodes/libome/proto/auth"
	"github.com/omecodes/omestore/ent"
	"github.com/omecodes/omestore/ent/user"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/omestore/store/internals"
	"github.com/omecodes/omestore/store/store"
	"github.com/sethvargo/go-password/password"
	"github.com/shibukawa/configdir"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

var debug = os.Getenv("OMESTORE_DEBUG")

// Config contains info to run configure an instance of Server
type Config struct {
	OmeHost         string
	OmeCertFilename string
	Dir             string
	DSN             string
	Domain          string
	Port            int
	Ome             *ome.Info
}

// NewServer is a server constructor
func NewServer(config Config, opts ...netx.ListenOption) *Server {
	s := new(Server)
	s.options = opts
	s.config = &config
	return s
}

// Server is a omestore API server
type Server struct {
	initialized bool

	options []netx.ListenOption
	config  *Config
	OmeInfo *ome.Info

	adminPassword   string
	dataStore       pb.Store
	dataAccessRules mapping.DoubleMap
	appData         internals.Store
	key             []byte
	celEnv          *cel.Env
	entDB           *ent.Client

	dListener net.Listener
}

func (s *Server) init() error {
	if s.initialized {
		return nil
	}
	s.initialized = true

	var err error
	s.OmeInfo, err = ome.GetInfo(s.config.OmeHost, s.config.OmeCertFilename)
	if err != nil {
		return err
	}

	if s.config.Dir == "" {
		dirs := configdir.New("Ome", "omestore")
		globalFolder := dirs.QueryFolders(configdir.Global)[0]
		s.config.Dir = globalFolder.Path
	} else {
		s.config.Dir, err = filepath.Abs(s.config.Dir)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(s.config.Dir, os.ModePerm)
	if err != nil {
		return err
	}

	log.File = filepath.Join(s.config.Dir, "omestore.log")

	db, err := sql.Open("mysql", s.config.DSN)
	if err != nil {
		return err
	}

	s.entDB, err = ent.Open("mysql", s.config.DSN)
	if err != nil {
		return err
	}

	err = s.entDB.Schema.Create(context.Background())
	if err != nil {
		return err
	}

	s.dataStore, err = store.MySQL(db)
	if err != nil {
		return err
	}

	s.appData, err = internals.NewStore(db)

	err = os.MkdirAll(s.config.Dir, os.ModePerm)
	if err != nil {
		return err
	}

	s.celEnv, err = cel.NewEnv(
		cel.Types(&pb.AuthCEL{}),
		cel.Types(&pb.DataCEL{}),
		cel.Types(&pb.PermCEL{}),
		cel.Types(&pb.GraftCEL{}),
		cel.Declarations(
			decls.NewVar("at", decls.Int),
			decls.NewVar("auth", decls.NewObjectType("pb.AuthCEL")),
			decls.NewVar("data", decls.NewObjectType("pb.DataCEL")),
			decls.NewVar("graft", decls.NewObjectType("pb.GraftCEL")),
			decls.NewFunction("acl",
				decls.NewOverload(
					"acl",
					[]*expr.Type{decls.String, decls.String}, decls.NewObjectType("pb.PermCEL"),
				),
			),
		),
	)
	if err != nil {
		return err
	}

	adminPwdFilename := filepath.Join(s.config.Dir, "admin-pwd")
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

	settings, err := s.appData.Get(internals.Settings, "")
	if err != nil && !errors.IsNotFound(err) {
		log.Error("could not get store settings")
		return err
	}

	if settings == "" {
		return s.appData.Set(internals.Settings, defaultSettings)
	}

	return nil
}

func (s *Server) getStoredKey(name string, size int) ([]byte, error) {
	cookiesKeyFilename := filepath.Join(s.config.Dir, name+".key")
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

// Start starts omestore API server
func (s *Server) Start() error {
	err := s.init()
	if err != nil {
		return err
	}
	return s.startDataServer()
}

func (s *Server) startDataServer() error {
	var err error
	address := fmt.Sprintf("%s:%d", s.config.Domain, s.config.Port)
	s.dListener, err = netx.Listen(address, s.options...)
	if err != nil {
		return err
	}

	fmt.Println("Running server at", address)
	var handler http.Handler

	router := dataRouter()
	handler = s.updateContext(router)
	handler = s.detectAuthentication(handler)
	handler = httpx.Logger("omestore").Handle(handler)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.Domain, s.config.Port),
		Handler: handler,
	}
	go srv.Serve(s.dListener)
	return nil
}

func (s *Server) startControlServer() error {
	return nil
}

func (s *Server) updateContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = s.contextWithDataDB()(ctx)
		ctx = s.contextWithDB()(ctx)
		ctx = s.contextWithDataDir()(ctx)
		ctx = s.contextWithAdminPassword()(ctx)
		ctx = s.contextWithAppData()(ctx)
		ctx = s.contextWithCelEnv()(ctx)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) detectAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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
					matchingUser, err := s.entDB.User.Query().Where(user.ID(authUser)).First(ctx)
					if err != nil {
						if ent.IsNotFound(err) {
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
					if hex.EncodeToString(hashed) != matchingUser.Password {
						w.WriteHeader(http.StatusForbidden)
						return
					}
				}

				ctx = ome.ContextWithCredentials(ctx, &ome.Credentials{
					Username: authUser,
					Password: pass,
				})
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) detectOAuth2Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header
		authorization := h.Get("Authorization")
		if authorization != "" {
			token := strings.TrimPrefix(authorization, "Bearer ")
			jwt, err := authpb.ParseJWT(token)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			state, err := jwt.Verify(s.OmeInfo.Oauth2.SignatureKey)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if state != authpb.JWTState_VALID {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			ctx := contextWithAuthCEL(r.Context(), pb.AuthCEL{
				Uid:       jwt.Claims.Sub,
				Email:     jwt.Claims.Email,
				Validated: jwt.Claims.EmailVerified,
				Scope:     strings.Split(jwt.Claims.Scope, ""),
			})
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// Stop stops omestore API server
func (s *Server) Stop() {
	_ = s.dListener.Close()
}
