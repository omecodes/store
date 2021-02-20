package store

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"github.com/omecodes/libome/logs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"

	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/netx"
	"github.com/omecodes/store/accounts"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/objects"
	"github.com/omecodes/store/webapp"
)

// Config contains info to configure an instance of Server
type Config struct {
	Dev          bool
	TLS          bool
	AutoCert     bool
	CertFilename string
	KeyFilename  string

	Domains    []string
	FSRootDir  string
	WorkingDir string
	WebDir     string
	AdminInfo  string
	DSN        string
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

	s.db = common.GetDB("mysql", s.config.DSN)

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
		s.sourceManager, err = files.NewSourceSQLManager(s.db, bome.MySQL, "store_")
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
				URI:         fmt.Sprintf("files://%s", files.NormalizePath(s.config.FSRootDir)),
				ExpireTime:  -1,
			}
			_, err = s.sourceManager.Save(ctx, source)
			if err != nil {
				return err
			}
		}
	}
	webapp.Dir = s.config.WebDir
	return nil
}

func (s *Server) GetRouter(_ context.Context) objects.Router {
	return objects.DefaultRouter()
}

// Start starts API server
func (s *Server) Start() error {
	err := s.init()
	if err != nil {
		return err
	}

	if s.config.Dev {
		return s.startDevServer()
	}

	if s.config.TLS {
		if s.config.AutoCert {
			return s.startAutoCertAPIServer()
		}
		return s.startSecureAPIServer()
	}

	return s.startNonSecureAPIServer()
}

func (s *Server) startDevServer() error {
	var err error
	s.listener, err = net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	address := s.listener.Addr().String()
	logs.Info("starting HTTP server", logs.Details("address", address))

	r := s.httpRouter()

	go func() {
		s.server = &http.Server{
			Addr:    address,
			Handler: r,
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
	r := s.httpRouter()
	srv := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			s.Errors <- err
		}
	}()

	logs.Info("starting HTTP Listener on Port 80")
	go func() {
		h := certManager.HTTPHandler(nil)
		if err := http.ListenAndServe(":80", h); err != nil {
			logs.Error("listen to port 80 failed", logs.Err(err))
		}
	}()
	return nil
}

func (s *Server) startSecureAPIServer() error {
	r := s.httpRouter()
	srv := &http.Server{
		Addr:    ":443",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServeTLS(s.config.CertFilename, s.config.KeyFilename); err != nil {
			s.Errors <- err
		}
	}()

	logs.Info("starting HTTP Listener on Port 80")
	go func() {
		if err := http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			redirectURL := fmt.Sprintf("https://%s", s.config.Domains[0])
			b := strings.Builder{}

			b.WriteString(fmt.Sprintf("<head>\n"))
			b.WriteString(fmt.Sprintf("\t<meta http-equiv=\"refresh\" content=\"0; URL=%s\" />\n", redirectURL))
			b.WriteString(fmt.Sprintf("</head>"))
			contentBytes := []byte(b.String())

			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("Location", redirectURL)
			w.WriteHeader(http.StatusPermanentRedirect)
			_, _ = w.Write(contentBytes)
		})); err != nil {
			logs.Error("listen to port 80 failed", logs.Err(err))
		}
	}()
	return nil
}

func (s *Server) startNonSecureAPIServer() error {
	var err error
	s.listener, err = net.Listen("tcp", "0.0.0.0:80")
	if err != nil {
		return err
	}

	address := s.listener.Addr().String()
	logs.Info("starting HTTP server", logs.Details("address", address))

	r := s.httpRouter()

	go func() {
		s.server = &http.Server{
			Addr:    address,
			Handler: r,
		}
		if err := s.server.Serve(s.listener); err != nil {
			s.Errors <- err
		}
	}()
	return nil
}

func (s *Server) httpRouter() *mux.Router {
	r := mux.NewRouter()
	filesRouter := files.MuxRouter(
		files.Middleware(files.MiddlewareWithSourceManager(s.sourceManager)),
		accounts.Middleware(accounts.MiddlewareWithAccountManager(s.accountsManager)),
		auth.Middleware(
			auth.MiddlewareWithCredentials(s.credentialsManager),
			auth.MiddlewareWithProviderManager(s.authenticationProviders),
		),
		httpx.Logger("files").Handle,
	)

	r.PathPrefix("/api/files/").Subrouter().Name("ServeFiles").
		Handler(http.StripPrefix("/api/files", filesRouter))

	objectsRouter := objects.MuxRouter(
		objects.Middleware(
			objects.MiddlewareWithACLManager(s.accessStore),
			objects.MiddlewareWithRouterProvider(s),
			objects.MiddlewareWithDB(s.objects),
			objects.MiddlewareWithSettings(s.settings),
		),
		accounts.Middleware(
			accounts.MiddlewareWithAccountManager(s.accountsManager),
		),
		auth.Middleware(
			auth.MiddlewareWithCredentials(s.credentialsManager),
			auth.MiddlewareWithProviderManager(s.authenticationProviders),
		),
		httpx.Logger("objects").Handle,
	)
	r.PathPrefix("/api/objects/").Subrouter().Name("ServeObjects").
		Handler(http.StripPrefix("/api/objects", objectsRouter))

	authRouter := auth.MuxRouter(
		auth.Middleware(
			auth.MiddlewareWithCredentials(s.credentialsManager),
			auth.MiddlewareWithProviderManager(s.authenticationProviders),
		),
		httpx.Logger("auth").Handle,
	)
	r.PathPrefix("/api/auth/").Subrouter().Name("ManageAuthentication").
		Handler(http.StripPrefix("/api/auth", authRouter))

	accountsRouter := accounts.MuxRouter(
		accounts.Middleware(
			accounts.MiddlewareWithAccountManager(s.accountsManager),
		),
		httpx.Logger("accounts").Handle,
	)
	r.PathPrefix("/api/accounts/").Subrouter().Name("ManageAccounts").
		Handler(http.StripPrefix("/api/accounts", accountsRouter))

	staticFilesRouter := http.HandlerFunc(webapp.ServeApps)
	r.NotFoundHandler = staticFilesRouter

	return r
}

// Stop stops API server
func (s *Server) Stop() {
	if s.listener != nil {
		_ = s.listener.Close()
	}
	_ = s.db.Close()
}
