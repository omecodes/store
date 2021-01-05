package oms

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
	"github.com/gorilla/mux"
	"github.com/omecodes/discover"
	errors2 "github.com/omecodes/libome/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/service"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/objects"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/utils/log"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/libome/ports"
	sca "github.com/omecodes/services-ca"
	"github.com/omecodes/store/clients"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/router"
)

type MsConfig struct {
	Name         string
	BindIP       string
	Domain       string
	RegistryPort int
	CAPort       int
	APIPort      int
	DBUri        string
	JWTSecret    string
	WorkingDir   string
	Development  bool
}

func NewMSServer(cfg MsConfig) *MSServer {
	return &MSServer{config: cfg}
}

type MSServer struct {
	config         MsConfig
	settings       objects.SettingsManager
	listener       net.Listener
	adminPassword  string
	workerPassword string
	Errors         chan error
	loadBalancer   *router.BaseHandler
	registry       ome.Registry
	caServer       *sca.Server
}

func (s *MSServer) init() error {
	var err error

	db, err := sql.Open("mysql", s.config.DBUri)
	if err != nil {
		return err
	}

	s.settings, err = objects.NewSQLSettings(db, bome.MySQL, "objects_settings")
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

func (s *MSServer) getServiceSecret(name string) (string, error) {
	return "ome", nil
}

func (s *MSServer) startCA() error {
	workingDir := filepath.Join(s.config.WorkingDir, "ca")
	err := os.MkdirAll(workingDir, os.ModePerm)
	if err != nil {
		return err
	}

	port := s.config.CAPort
	if port == 0 {
		port = ports.CA
	}

	cfg := &sca.ServerConfig{
		Manager:    sca.CredentialsManagerFunc(s.getServiceSecret),
		Domain:     s.config.Domain,
		Port:       port,
		BindIP:     s.config.BindIP,
		WorkingDir: workingDir,
	}
	s.caServer = sca.NewServer(cfg)
	return s.caServer.Start()
}

func (s *MSServer) startRegistry() (err error) {
	s.registry, err = discover.Serve(&discover.ServerConfig{
		Name:                 s.config.Name,
		StoreDir:             s.config.WorkingDir,
		BindAddress:          fmt.Sprintf("%s:%d", s.config.BindIP, s.config.RegistryPort),
		CertFilename:         "ca/ca.crt",
		KeyFilename:          "ca/ca.key",
		ClientCACertFilename: "ca/ca.crt",
	})
	return
}

func (s *MSServer) startAPIServer() error {
	if !s.config.Development {
		return s.startProductionAPIServer()
	}

	err := s.init()
	if err != nil {
		return err
	}

	s.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.BindIP, s.config.APIPort))
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
		auth.DetectOauth2Middleware(s.config.JWTSecret),
		httpx.Logger("OMS").Handle,
		s.httpEnrichContext,
	}
	var handler http.Handler
	handler = NewHttpUnit().MuxRouter()

	for _, m := range middlewareList {
		handler = m.Middleware(handler)
	}

	go func() {
		srv := &http.Server{
			Addr:    address,
			Handler: handler,
		}
		if err := srv.Serve(s.listener); err != nil {
			s.Errors <- err
		}
	}()

	return nil
}

func (s *MSServer) startProductionAPIServer() error {
	cfg := simplecert.Default
	cfg.Domains = []string{s.config.Domain}
	cfg.CacheDir = filepath.Join(s.config.WorkingDir, "lets-encrypt")
	cfg.SSLEmail = "omecodes@gmail.com"
	cfg.DNSProvider = "cloudflare"

	certReloadAgent, err := simplecert.Init(cfg, nil)
	if err != nil {
		log.Fatal("simplecert init failed: ", log.Err(err))
	}

	log.Info("starting HTTP Listener on Port 80")
	go func() {
		if err := http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpx.Redirect(w, &httpx.RedirectURL{
				URL:         fmt.Sprintf("https://%s:443%s", s.config.Domain, r.URL.Path),
				Code:        http.StatusPermanentRedirect,
				ContentType: "text/html",
			})
		})); err != nil {
			logs.Error("Plain http listen caused error", logs.Err(err))
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
		auth.DetectOauth2Middleware(s.config.JWTSecret),
		httpx.Logger("OMS").Handle,
		s.httpEnrichContext,
	}
	var handler http.Handler
	handler = NewHttpUnit().MuxRouter()

	for _, m := range middlewareList {
		handler = m.Middleware(handler)
	}

	// init server
	srv := &http.Server{
		Addr:      fmt.Sprintf("%s:443", s.config.BindIP),
		TLSConfig: tlsConf,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("listen to port 443 failed", log.Err(err))
		}
	}()
	return nil
}

func (s *MSServer) httpEnrichContext(next http.Handler) http.Handler {
	box := service.CreateBox(
		service.Registry(s.registry),
		service.CertFile("ca/ca.crt"),
		service.KeyFIle("ca/ca.key"),
		service.CACertFile("ca/ca.crt"),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = service.ContextWithBox(ctx, box)
		ctx = router.WithRouterProvider(ctx, router.ProviderFunc(s.GetRouter))
		ctx = router.WithSettings(s.settings)(ctx)
		ctx = clients.WithRouterGrpcClientProvider(ctx, &clients.LoadBalancer{})

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *MSServer) GetRouter(ctx context.Context) router.Router {
	return router.NewCustomRouter(
		router.NewGRPCClientHandler(common.ServiceTypeHandler),
		router.WithDefaultParamsHandler(),
	)
}

func (s *MSServer) Start() error {
	err := s.init()
	if err != nil {
		return err
	}

	err = s.startCA()
	if err != nil {
		log.Error("MS Sore • could not start CA", log.Err(err))
		return errors.Internal
	}

	err = s.startRegistry()
	if err != nil {
		log.Error("MS Sore • could not start CA", log.Err(err))
		return errors.Internal
	}

	return s.startAPIServer()
}

func (s *MSServer) Stop() error {
	if closer, ok := s.registry.(interface {
		Close() error
	}); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return s.caServer.Stop()
}
