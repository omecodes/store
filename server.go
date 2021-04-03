package store

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/crypt"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/session"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"

	"github.com/omecodes/bome"
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
	settings                common.SettingsManager
	accountsManager         accounts.Manager
	authenticationProviders auth.ProviderManager
	credentialsManager      auth.CredentialsManager
	accessStore             objects.ACLManager
	sourceManager           files.SourceManager
	cookieStore             *sessions.CookieStore

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

	if s.config.AdminInfo == "" {
		adminAuthContent, err := ioutil.ReadFile("./admin-auth")
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		if adminAuthContent == nil {
			phrase, err := crypt.GenerateVerificationCode(8)
			if err != nil {
				return err
			}

			_, info, err := crypt.Generate(phrase, 16)
			if err != nil {
				return err
			}

			data, err := json.Marshal(info)
			if err != nil {
				return err
			}

			s.config.AdminInfo = base64.RawStdEncoding.EncodeToString(data)
			err = ioutil.WriteFile("./admin-auth", []byte(phrase+":"+s.config.AdminInfo), os.ModePerm)
			if err != nil {
				return err
			}

		} else {
			parts := strings.Split(string(adminAuthContent), ":")
			s.config.AdminInfo = parts[1]
		}
	}

	s.accessStore, err = objects.NewSQLACLStore(s.db, bome.MySQL, "store_acl")
	if err != nil {
		return err
	}

	s.settings, err = common.NewSQLSettings(s.db, bome.MySQL, "store_settings")
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

	_, err = s.settings.Get(common.SettingsDataMaxSizePath)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		err = s.settings.Set(common.SettingsDataMaxSizePath, common.DefaultSettings[common.SettingsDataMaxSizePath])
		if err != nil && !errors.IsConflict(err) {
			return err
		}
	}

	_, err = s.settings.Get(common.SettingsDataMaxSizePath)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		err = s.settings.Set(common.SettingsCreateDataSecurityRule, common.DefaultSettings[common.SettingsCreateDataSecurityRule])
		if err != nil && !errors.IsConflict(err) {
			return err
		}
	}

	_, err = s.settings.Get(common.SettingsObjectListMaxCount)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		err = s.settings.Set(common.SettingsObjectListMaxCount, common.DefaultSettings[common.SettingsObjectListMaxCount])
		if err != nil && !errors.IsConflict(err) {
			return err
		}
	}

	cookiesKey, err := s.generateKey(64)
	if err != nil {
		return err
	}
	s.cookieStore = sessions.NewCookieStore(cookiesKey[:31], cookiesKey[32:])

	// Files initialization
	if s.config.FSRootDir != "" {
		s.sourceManager, err = files.NewSourceSQLManager(s.db, bome.MySQL, "store")
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
				Id:          "main",
				Label:       "Default file source",
				Description: "",
				Type:        files.SourceType_Default,
				Uri:         fmt.Sprintf("files://%s", files.NormalizePath(s.config.FSRootDir)),
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

func (s *Server) generateKey(size int) ([]byte, error) {
	key := make([]byte, size)
	_, err := rand.Read(key)
	if err != nil {
		log.Error("could not generate secret key", log.Err(err))
		return nil, err
	}
	return key, nil
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
			logs.Info("starting api server over tls with auto-cert")
			return s.startAutoCertAPIServer()
		}
		logs.Info("starting api server over TLS")
		return s.startSecureAPIServer()
	}

	logs.Info("starting api server")
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
	s.listener, err = net.Listen("tcp", ":80")
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

func (s *Server) httpRouter() http.Handler {
	r := mux.NewRouter()
	filesRouter := files.MuxRouter(
		files.Middleware(files.MiddlewareWithSourceManager(s.sourceManager)),
		accounts.Middleware(accounts.MiddlewareWithAccountManager(s.accountsManager)),
		auth.Middleware(
			auth.MiddlewareWithCredentials(s.credentialsManager),
			auth.MiddlewareWithProviderManager(s.authenticationProviders),
		),
	)

	r.PathPrefix("/api/files/").Subrouter().Name("ServeFiles").
		Handler(http.StripPrefix("/api/files", filesRouter))

	objectsRouter := objects.MuxRouter(
		objects.Middleware(
			objects.MiddlewareWithACLManager(s.accessStore),
			//objects.MiddlewareWithRouterProvider(s),
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
	)
	r.PathPrefix("/api/objects/").Subrouter().Name("ServeObjects").
		Handler(http.StripPrefix("/api/objects", objectsRouter))

	authRouter := auth.MuxRouter(
		auth.Middleware(
			auth.MiddlewareWithCredentials(s.credentialsManager),
			auth.MiddlewareWithProviderManager(s.authenticationProviders),
		),
	)
	r.PathPrefix("/api/auth/").Subrouter().Name("ManageAuthentication").
		Handler(http.StripPrefix("/api/auth", authRouter))

	r.Handle("/login", auth.UserSessionHandler(
		auth.Middleware(
			auth.MiddlewareWithCredentials(s.credentialsManager),
			auth.MiddlewareWithProviderManager(s.authenticationProviders),
		),
	)).Methods(http.MethodPost)

	accountsRouter := accounts.MuxRouter(
		accounts.Middleware(
			accounts.MiddlewareWithAccountManager(s.accountsManager),
		),
	)

	r.PathPrefix("/api/accounts/").Subrouter().Name("ManageAccounts").
		Handler(http.StripPrefix("/api/accounts", accountsRouter))

	staticFilesRouter := http.HandlerFunc(webapp.ServeApps)
	r.NotFoundHandler = staticFilesRouter

	return common.MiddlewareLogger(session.WithHTTPSessionMiddleware(s.cookieStore)(r))
}

// Stop stops API server
func (s *Server) Stop() {
	if s.listener != nil {
		_ = s.listener.Close()
	}
	_ = s.db.Close()
}
