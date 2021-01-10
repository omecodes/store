package oms

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"github.com/google/cel-go/cel"
	"github.com/gorilla/mux"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/cenv"
	"github.com/omecodes/store/objects"
	"golang.org/x/crypto/acme/autocert"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/omecodes/bome"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/netx"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/router"
)

// Config contains info to configure an instance of Server
type MNConfig struct {
	Dev        bool
	Domains    []string
	WorkingDir string
	AdminInfo  string
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
	initialized  bool
	options      []netx.ListenOption
	config       *MNConfig
	key          []byte
	celPolicyEnv *cel.Env
	celSearchEnv *cel.Env
	autoCertDir  string

	objects                 objects.Objects
	settings                objects.SettingsManager
	authenticationProviders auth.ProviderManager
	credentialsManager      auth.CredentialsManager
	accessStore             acl.Store

	listener net.Listener
	Errors   chan error
	server   *http.Server
}

func (s *MNServer) init() error {
	if s.initialized {
		return nil
	}
	s.initialized = true

	if !s.config.Dev {
		s.autoCertDir = filepath.Join(s.config.WorkingDir, "autocert")
		err := os.MkdirAll(s.autoCertDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	db, err := sql.Open("mysql", s.config.DSN)
	if err != nil {
		return err
	}

	s.accessStore, err = acl.NewSQLStore(db, bome.MySQL, "objects_acl")
	if err != nil {
		return err
	}

	s.settings, err = objects.NewSQLSettings(db, bome.MySQL, "objects_settings")
	if err != nil {
		return err
	}

	s.objects, err = objects.NewSQLStore(db, bome.MySQL, "objects")
	if err != nil {
		return err
	}

	s.credentialsManager, err = auth.NewCredentialsSQLManager(db, bome.MySQL, "agents", s.config.AdminInfo)
	if err != nil {
		return err
	}

	s.authenticationProviders, err = auth.NewProviderSQLManager(db, bome.MySQL, "auth_providers")
	if err != nil {
		return err
	}

	s.celPolicyEnv, err = cenv.ACLEnv()
	if err != nil {
		return err
	}

	s.celSearchEnv, err = cenv.SearchEnv()
	if err != nil {
		return err
	}

	s.key, err = s.getStoredKey("token-key", 64)
	if err != nil {
		return err
	}

	_, err = s.settings.Get(objects.SettingsDataMaxSizePath)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		err = s.settings.Set(objects.SettingsDataMaxSizePath, objects.DefaultSettings[objects.SettingsDataMaxSizePath])
		if err != nil && !bome.IsPrimaryKeyConstraintError(err) {
			return err
		}
	}

	_, err = s.settings.Get(objects.SettingsDataMaxSizePath)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		err = s.settings.Set(objects.SettingsCreateDataSecurityRule, objects.DefaultSettings[objects.SettingsCreateDataSecurityRule])
		if err != nil && !bome.IsPrimaryKeyConstraintError(err) {
			return err
		}
	}

	_, err = s.settings.Get(objects.SettingsObjectListMaxCount)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		err = s.settings.Set(objects.SettingsObjectListMaxCount, objects.DefaultSettings[objects.SettingsObjectListMaxCount])
		if err != nil && !bome.IsPrimaryKeyConstraintError(err) {
			return err
		}
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

	if !s.config.Dev {
		return s.startAutoCertAPIServer()
	}

	return s.startDefaultAPIServer()
}

func (s *MNServer) startDefaultAPIServer() error {
	var err error
	s.listener, err = net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	address := s.listener.Addr().String()
	log.Info("starting HTTP server", log.Field("address", address))

	middlewareList := []mux.MiddlewareFunc{
		auth.DetectBasicMiddleware,
		auth.DetectOauth2Middleware,
		s.enrichContext,
		httpx.Logger("store").Handle,
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

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.config.Domains...),
	}
	certManager.Cache = autocert.DirCache(s.autoCertDir)

	middlewareList := []mux.MiddlewareFunc{
		auth.DetectBasicMiddleware,
		auth.DetectOauth2Middleware,
		s.enrichContext,
		httpx.Logger("store").Handle,
	}
	var handler http.Handler
	handler = NewHttpUnit().MuxRouter()

	for _, m := range middlewareList {
		handler = m.Middleware(handler)
	}
	// create the server itself
	srv := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		Handler: handler,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			s.Errors <- err
		}
	}()

	log.Info("starting HTTP Listener on Port 80")
	go func() {
		h := certManager.HTTPHandler(nil)
		if err := http.ListenAndServe(":80", h); err != nil {
			log.Error("listen to port 80 failed", log.Err(err))
		}
	}()

	return nil
}

func (s *MNServer) enrichContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = auth.ContextWithCredentialsManager(ctx, s.credentialsManager)
		ctx = auth.ContextWithProviders(ctx, s.authenticationProviders)
		ctx = acl.ContextWithStore(ctx, s.accessStore)
		ctx = objects.ContextWithStore(ctx, s.objects)
		ctx = router.WithCelPolicyEnv(s.celPolicyEnv)(ctx)
		ctx = router.WithCelSearchEnv(s.celSearchEnv)(ctx)
		ctx = router.WithSettings(s.settings)(ctx)
		ctx = router.WithRouterProvider(ctx, s)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// Stop stops API server
func (s *MNServer) Stop() {
	_ = s.listener.Close()
}
