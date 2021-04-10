package service

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/omecodes/bome"
	"github.com/omecodes/discover"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/libome/crypt"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/libome/ports"
	sca "github.com/omecodes/services-ca"
	"github.com/omecodes/store/accounts"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/objects"
	"github.com/omecodes/store/session"
	"github.com/omecodes/store/webapp"
	"golang.org/x/crypto/acme/autocert"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

type FrontConfig struct {
	Dev             bool
	Name            string
	AdminAuth       string
	Domains         []string
	IP              string
	CAPort          int
	RegistryPort    int
	Port            int
	WorkingDir      string
	Database        string
	TLS             bool
	TLSAuto         bool
	TLSCertFilename string
	TLSKeyFilename  string
}

func NewFront(config FrontConfig) *Front {
	return &Front{config: &config}
}

type Front struct {
	config *FrontConfig

	cookieStore *sessions.CookieStore
	settings    common.SettingsManager
	accounts    accounts.Manager
	idProviders auth.ProviderManager
	credentials auth.CredentialsManager

	registry ome.Registry
	caServer *sca.Server

	db *sql.DB

	listener  net.Listener
	apiServer *http.Server
	Errors    chan error
}

func (f *Front) init() error {
	f.db = common.GetDB(bome.MySQL, f.config.Database)

	var err error

	if f.config.AdminAuth == "" {
		adminAuthContent, err := ioutil.ReadFile(common.AdminAuthFile)
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

			f.config.AdminAuth = base64.RawStdEncoding.EncodeToString(data)
			err = ioutil.WriteFile(common.AdminAuthFile, []byte(phrase+":"+f.config.AdminAuth), os.ModePerm)
			if err != nil {
				return err
			}

		} else {
			parts := strings.Split(string(adminAuthContent), ":")
			f.config.AdminAuth = parts[1]
		}
	}

	f.settings, err = common.NewSQLSettings(f.db, bome.MySQL, "store_settings")
	if err != nil {
		return err
	}

	f.accounts, err = accounts.NewSQLManager(f.db, bome.MySQL, "store")
	if err != nil {
		return err
	}

	f.credentials, err = auth.NewCredentialsSQLManager(f.db, bome.MySQL, "store", f.config.AdminAuth)
	if err != nil {
		return err
	}

	f.idProviders, err = auth.NewProviderSQLManager(f.db, bome.MySQL, "store_auth_providers")
	if err != nil {
		return err
	}

	cookiesKey, err := common.LoadOrGenerateKey(common.CookiesKeyFilename, 64)
	if err != nil {
		return err
	}

	f.cookieStore = sessions.NewCookieStore(cookiesKey[:31], cookiesKey[32:])

	f.Errors = make(chan error, 4)
	return nil
}

func (f *Front) provideFilesRouter(ctx context.Context) files.Router {
	return files.NewCustomRouter(
		files.NewHandlerServiceClient(common.ServiceTypeSecurityAccess),
		files.WithDefaultParamsHandler(),
	)
}

func (f *Front) provideObjectsRouter(ctx context.Context) objects.Router {
	return objects.NewCustomRouter(
		objects.NewGRPCObjectsClientHandler(common.ServiceTypeSecurityAccess),
		objects.WithDefaultParamsHandler(),
	)
}

func (f *Front) httpRouter() http.Handler {
	middleware := []mux.MiddlewareFunc{
		auth.Middleware(
			auth.MiddlewareWithCredentials(f.credentials),
			auth.MiddlewareWithProviderManager(f.idProviders),
		),
		accounts.Middleware(accounts.MiddlewareWithAccountManager(f.accounts)),
		session.WithHTTPSessionMiddleware(f.cookieStore),
		common.MiddlewareWithSettings(f.settings),
		common.MiddlewareLogger,
	}
	r := mux.NewRouter()

	r.PathPrefix("/api/files/").Subrouter().Name("ServeFiles").Handler(http.StripPrefix("/api/files", f.filesHandler()))
	r.PathPrefix("/api/objects/").Subrouter().Name("ServeObjects").Handler(http.StripPrefix("/api/objects", f.objectsHandler()))
	r.PathPrefix("/api/auth/").Subrouter().Name("ManageAuthentication").Handler(http.StripPrefix("/api/auth", auth.MuxRouter()))
	r.Handle("/login", auth.UserSessionHandler()).Methods(http.MethodPost)
	r.PathPrefix("/api/accounts/").Subrouter().Name("ManageAccounts").Handler(http.StripPrefix("/api/accounts", accounts.MuxRouter()))

	staticFilesRouter := http.HandlerFunc(webapp.ServeApps)
	r.NotFoundHandler = staticFilesRouter

	var handler http.Handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func (f *Front) filesHandler() http.Handler {
	return files.MuxRouter(
		files.Middleware(
			files.MiddlewareWithRouterProvider(files.RouterProvideFunc(f.provideFilesRouter)),
		),
	)
}

func (f *Front) objectsHandler() http.Handler {
	return objects.MuxRouter(
		objects.Middleware(
			objects.MiddlewareWithRouterProvider(objects.RouterProvideFunc(f.provideObjectsRouter)),
		),
	)
}

func (f *Front) getServiceSecret(name string) (string, error) {
	return "ome", nil
}

func (f *Front) startRegistryServer() (err error) {
	f.registry, err = discover.Serve(&discover.ServerConfig{
		Name:                 f.config.Name,
		StoreDir:             f.config.WorkingDir,
		BindAddress:          fmt.Sprintf("%s:%d", f.config.IP, f.config.RegistryPort),
		CertFilename:         "ca/ca.crt",
		KeyFilename:          "ca/ca.key",
		ClientCACertFilename: "ca/ca.crt",
	})
	return
}

func (f *Front) startCAServer() error {
	port := f.config.CAPort
	if port == 0 {
		port = ports.CA
	}

	cfg := &sca.ServerConfig{
		Manager: sca.CredentialsManagerFunc(f.getServiceSecret),
		Domain:  f.config.Domains[0],
		Port:    port,
		BindIP:  f.config.IP,
	}

	f.caServer = sca.NewServer(cfg)
	return f.caServer.Start()
}

func (f *Front) startDev() error {
	err := f.init()
	if err != nil {
		return err
	}

	f.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", f.config.IP, f.config.Port))
	if err != nil {
		return err
	}

	address := f.listener.Addr().String()
	logs.Info("starting HTTP server", logs.Details("address", address))

	go func() {
		srv := &http.Server{
			Addr:    address,
			Handler: f.httpRouter(),
		}
		if err := srv.Serve(f.listener); err != nil {
			f.Errors <- err
		}
	}()

	return nil
}

func (f *Front) startProductionWithTLSAutoCert() error {
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(f.config.Domains...),
	}
	certManager.Cache = autocert.DirCache(f.config.WorkingDir)

	f.apiServer = &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		Handler: f.httpRouter(),
	}

	go func() {
		if err := f.apiServer.ListenAndServe(); err != nil {
			f.Errors <- err
		}
	}()

	logs.Info("starting front HTTP server on 80")
	go func() {
		h := certManager.HTTPHandler(nil)
		if err := http.ListenAndServe(":80", h); err != nil {
			logs.Error("listen to port 80 failed", logs.Err(err))
		}
	}()
	return nil
}

func (f *Front) startProductionWithTLS() error {
	r := f.httpRouter()
	srv := &http.Server{
		Addr:    ":443",
		Handler: r,
	}
	go func() {
		logs.Info("starting front Listener on Port 443")
		if err := srv.ListenAndServeTLS(f.config.TLSCertFilename, f.config.TLSKeyFilename); err != nil {
			f.Errors <- err
		}
	}()

	go func() {
		logs.Info("starting front Listener on Port 80")
		if err := http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			redirectURL := fmt.Sprintf("https://%s", f.config.Domains[0])
			b := strings.Builder{}
			b.WriteString(fmt.Sprintf("<head>\n"))
			b.WriteString(fmt.Sprintf("\t<meta http-equiv=\"refresh\" content=\"0; URL=%s\" />\n", redirectURL))
			b.WriteString(fmt.Sprintf("</head>"))
			contentBytes := []byte(b.String())
			w.Header().Set(common.HttpHeaderContentType, r.Header.Get("Accept"))
			w.Header().Set("Location", redirectURL)
			w.WriteHeader(http.StatusPermanentRedirect)
			_, _ = w.Write(contentBytes)
		})); err != nil {
			logs.Error("could not run front HTTP server on port 80", logs.Err(err))
			f.Errors <- err
		}
	}()
	return nil
}

func (f *Front) startProductionWithoutTLS() error {
	logs.Info("starting HTTP Listener on Port 80")
	go func() {
		if err := http.ListenAndServe(":80", f.httpRouter()); err != nil {
			logs.Error("listen to port 80 failed", logs.Err(err))
			f.Errors <- err
		}
	}()
	return nil
}

func (f *Front) Start() error {
	err := f.init()
	if err != nil {
		return err
	} else {
		if f.config.Dev {
			err = f.startDev()
			if err != nil {
				return err
			}
		} else {
			if f.config.TLS {
				if f.config.TLSAuto {
					err = f.startProductionWithTLSAutoCert()
				} else {
					err = f.startProductionWithTLS()
				}
			} else {
				err = f.startProductionWithoutTLS()
			}

			if err != nil {
				return err
			}
		}
	}

	err = f.startCAServer()
	if err != nil {
		return err
	}

	return f.startRegistryServer()
}

func (f *Front) Stop() error {
	return nil
}
