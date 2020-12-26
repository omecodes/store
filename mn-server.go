package oms

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/gorilla/mux"
	"github.com/omecodes/omestore/auth"
	"github.com/sethvargo/go-password/password"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/netx"
	"github.com/omecodes/common/utils/log"
	errors2 "github.com/omecodes/libome/errors"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/router"
	"github.com/omecodes/omestore/services/acl"
)

// Config contains info to configure an instance of Server
type MNConfig struct {
	Domains    []string
	WorkingDir string
	AutoCert   bool
	JwtSecret  string
	DSN        string
}

// NewMNServer is a server constructor
func NewMNServer(config MNConfig) *MNServer {
	s := new(MNServer)
	s.config = &config
	return s
}

// Server embeds an Ome data store
// it also exposes an API server
type MNServer struct {
	initialized   bool
	options       []netx.ListenOption
	config        *MNConfig
	adminPassword string
	key           []byte
	celPolicyEnv  *cel.Env
	celSearchEnv  *cel.Env

	objects     oms.Objects
	workers     *bome.JSONMap
	settings    oms.SettingsManager
	accessStore acl.Store
	listener    net.Listener
	Errors      chan error
	server      *http.Server
}

func (s *MNServer) init() error {
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

	s.accessStore, err = acl.NewSQLStore(db, bome.MySQL, "accesses")
	if err != nil {
		return err
	}

	s.settings, err = oms.NewSQLSettings(db, bome.MySQL, "settings")
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
		return err
	}

	adminPwdFilename := filepath.Join(s.config.WorkingDir, "admin-pwd")
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

	err = s.settings.Set(oms.SettingsDataMaxSizePath, oms.DefaultSettings[oms.SettingsDataMaxSizePath])
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	err = s.settings.Set(oms.SettingsCreateDataSecurityRule, oms.DefaultSettings[oms.SettingsCreateDataSecurityRule])
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func (s *MNServer) getStoredKey(name string, size int) ([]byte, error) {
	cookiesKeyFilename := filepath.Join(s.config.WorkingDir, name+".key")
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

func (s *MNServer) GetRouter(ctx context.Context) router.Router {
	return router.DefaultRouter()
}

// Start starts API server
func (s *MNServer) Start() error {
	err := s.init()
	if err != nil {
		return err
	}

	if s.config.AutoCert {
		return s.startAutoCertAPIServer()
	}

	return s.startDefaultAPIServer()
}

func (s *MNServer) startDefaultAPIServer() error {
	var err error
	s.listener, err = net.Listen("tcp", ":")
	if err != nil {
		return err
	}

	address := s.listener.Addr().String()
	log.Info("starting HTTP server", log.Field("address", address))

	middlewareList := []mux.MiddlewareFunc{
		auth.DetectBasicMiddleware(auth.CredentialsMangerFunc(func(user string) (string, error) {
			if user == "admin" {
				return s.adminPassword, nil
			}
			return "", errors2.New(errors2.CodeForbidden, "authentication failed")
		})),
		auth.DetectOauth2Middleware(s.config.JwtSecret),
		s.enrichContext,
		httpx.Logger("omestore").Handle,
	}
	var handler http.Handler
	handler = NewHttpUnit().MuxRouter()

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

func (s *MNServer) startAutoCertAPIServer() error {
	cfg := simplecert.Default
	cfg.Domains = s.config.Domains
	cfg.CacheDir = filepath.Join(s.config.WorkingDir, "lets-encrypt")
	cfg.SSLEmail = "omecodes@gmail.com"
	cfg.DNSProvider = "cloudflare"

	certReloadAgent, err := simplecert.Init(cfg, nil)
	if err != nil {
		return err
	}

	log.Info("starting HTTP Listener on Port 80")
	go func() {
		if err := http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpx.Redirect(w, &httpx.RedirectURL{
				URL:         fmt.Sprintf("https://%s:443%s", s.config.Domains[0], r.URL.Path),
				Code:        http.StatusPermanentRedirect,
				ContentType: "text/html",
			})
		})); err != nil {
			log.Error("listen to port 80 failed", log.Err(err))
		}
	}()

	tlsConf := tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)
	tlsConf.GetCertificate = certReloadAgent.GetCertificateFunc()

	middlewareList := []mux.MiddlewareFunc{
		auth.DetectBasicMiddleware(auth.CredentialsMangerFunc(func(user string) (string, error) {
			if user == "admin" {
				return s.adminPassword, nil
			}
			return "", errors2.New(errors2.CodeForbidden, "authentication failed")
		})),
		auth.DetectOauth2Middleware(s.config.JwtSecret),
		s.enrichContext,
		httpx.Logger("omestore").Handle,
	}
	var handler http.Handler
	handler = NewHttpUnit().MuxRouter()

	for _, m := range middlewareList {
		handler = m.Middleware(handler)
	}

	// init server
	srv := &http.Server{
		Addr:      ":443",
		TLSConfig: tlsConf,
		Handler:   handler,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			s.Errors <- err
		}
	}()
	return nil
}

func (s *MNServer) enrichContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = router.WithAccessStore(s.accessStore)(ctx)
		ctx = router.WithCelPolicyEnv(s.celPolicyEnv)(ctx)
		ctx = router.WithCelSearchEnv(s.celSearchEnv)(ctx)
		ctx = router.WithSettings(s.settings)(ctx)
		ctx = router.WithObjectsStore(s.objects)(ctx)
		ctx = router.WithWorkers(s.workers)(ctx)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// Stop stops API server
func (s *MNServer) Stop() {
	_ = s.listener.Close()
}