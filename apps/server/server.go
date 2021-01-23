package server

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/webapp"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"

	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/netx"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/accounts"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/objects"
)

// Config contains info to configure an instance of Server
type Config struct {
	Dev            bool
	Domains        []string
	FSRootDir      string
	WorkingDir     string
	WebAppsDir     string
	StaticFilesDir string
	AdminInfo      string
	DSN            string
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
	initialized bool
	options     []netx.ListenOption
	config      *Config
	autoCertDir string

	objects                 objects.DB
	settings                objects.SettingsManager
	accountsManager         accounts.Manager
	authenticationProviders auth.ProviderManager
	credentialsManager      auth.CredentialsManager
	accessStore             objects.ACLManager
	sourceManager           files.SourceManager

	listener net.Listener
	Errors   chan error
	server   *http.Server
	db       *sql.DB
}

func (s *Server) init() error {
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

	s.db = GetDB("mysql", s.config.DSN)

	var err error
	s.accessStore, err = objects.NewSQLACLStore(s.db, bome.MySQL, "store_acl")
	if err != nil {
		return err
	}

	s.settings, err = objects.NewSQLSettings(s.db, bome.MySQL, "store_settings")
	if err != nil {
		return err
	}

	s.accountsManager, err = accounts.NewSQLManager(s.db, bome.MySQL, "store")
	if err != nil {
		return err
	}

	s.objects, err = objects.NewSqlDB(s.db, bome.MySQL, "store")
	if err != nil {
		return err
	}

	s.credentialsManager, err = auth.NewCredentialsSQLManager(s.db, bome.MySQL, "store", s.config.AdminInfo)
	if err != nil {
		return err
	}

	s.authenticationProviders, err = auth.NewProviderSQLManager(s.db, bome.MySQL, "store_auth_providers")
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

	// Files initialization
	if s.config.FSRootDir != "" {
		s.sourceManager, err = files.NewSourceSQLManager(s.db, bome.MySQL, "store_files_sources")
		if err != nil {
			return err
		}

		ctx := context.Background()
		source, err := s.sourceManager.Get(ctx, "main")
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if source == nil {
			source = &files.Source{
				ID:          "main",
				Label:       "Default file source",
				Description: "",
				Type:        files.TypeDisk,
				URI:         fmt.Sprintf("files://%s", s.config.FSRootDir),
				ExpireTime:  -1,
			}
			_, err = s.sourceManager.Save(ctx, source)
			if err != nil {
				return err
			}
		}
	}

	if s.config.WebAppsDir != "" {
		webapp.Dir = s.config.WebAppsDir
		go webapp.WatchDir()
	}

	return nil
}

func (s *Server) GetRouter(ctx context.Context) objects.Router {
	return objects.DefaultRouter()
}

// Start starts API server
func (s *Server) Start() error {
	err := s.init()
	if err != nil {
		return err
	}

	if !s.config.Dev {
		return s.startAutoCertAPIServer()
	}

	return s.startDefaultAPIServer()
}

func (s *Server) startDefaultAPIServer() error {
	var err error
	s.listener, err = net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	address := s.listener.Addr().String()
	log.Info("starting HTTP server", log.Field("address", address))

	middlewareList := []mux.MiddlewareFunc{
		objects.Middleware(
			objects.MiddlewareWithACLManager(s.accessStore),
			objects.MiddlewareWithRouterProvider(s),
			objects.MIddlewareWithDB(s.objects),
			objects.MiddlewareWithSettings(s.settings),
		),
		files.Middleware(
			files.MiddlewareWithSourceManager(s.sourceManager),
		),
		accounts.Middleware(
			accounts.MiddlewareWithAccountManager(s.accountsManager),
		),
		auth.Middleware(
			auth.MiddlewareWithCredentials(s.credentialsManager),
			auth.MiddlewareWithProviderManager(s.authenticationProviders),
		),
		httpx.Logger("store").Handle,
	}

	var handler http.Handler
	handler = httpRouter(
		WithObjects(),
		WithStaticFiles(s.config.StaticFilesDir),
		WithFiles(s.config.FSRootDir != ""),
		WithWebApp(s.config.WebAppsDir != ""),
		WithAccounts(true),
		WithAuth(true),
	)

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

func (s *Server) startAutoCertAPIServer() error {

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.config.Domains...),
	}
	certManager.Cache = autocert.DirCache(s.autoCertDir)

	middlewareList := []mux.MiddlewareFunc{
		objects.Middleware(
			objects.MiddlewareWithACLManager(s.accessStore),
			objects.MiddlewareWithRouterProvider(s),
			objects.MIddlewareWithDB(s.objects),
			objects.MiddlewareWithSettings(s.settings),
		),
		files.Middleware(
			files.MiddlewareWithSourceManager(s.sourceManager),
		),
		accounts.Middleware(
			accounts.MiddlewareWithAccountManager(s.accountsManager),
		),
		auth.Middleware(
			auth.MiddlewareWithCredentials(s.credentialsManager),
			auth.MiddlewareWithProviderManager(s.authenticationProviders),
		),
		httpx.Logger("store").Handle,
	}

	var handler http.Handler
	handler = httpRouter(
		WithObjects(),
		WithStaticFiles(s.config.StaticFilesDir),
		WithFiles(s.config.FSRootDir != ""),
		WithWebApp(s.config.WebAppsDir != ""),
		WithAccounts(true),
		WithAuth(true),
	)

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

// Stop stops API server
func (s *Server) Stop() {
	webapp.StopWatch()

	_ = s.listener.Close()
	_ = s.db.Close()
}
