package oms

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"

	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/discover"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/libome/ports"
	"github.com/omecodes/service"
	sca "github.com/omecodes/services-ca"
	"github.com/omecodes/store/accounts"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/objects"
)

type MsConfig struct {
	Name         string
	BindIP       string
	Domains      []string
	RegistryPort int
	CAPort       int
	APIPort      int
	DBUri        string
	AdminInfo    string
	WorkingDir   string
	Dev          bool
}

func NewMSServer(cfg MsConfig) *MSServer {
	return &MSServer{config: cfg}
}

type MSServer struct {
	config         MsConfig
	listener       net.Listener
	adminPassword  string
	workerPassword string
	Errors         chan error
	loadBalancer   *objects.BaseHandler
	registry       ome.Registry
	caServer       *sca.Server
	autoCertDir    string
	db             *sql.DB

	authenticationProviders auth.ProviderManager
	credentialsManager      auth.CredentialsManager
	settings                objects.SettingsManager
	accountsManager         accounts.Manager
}

func (s *MSServer) init() error {

	if !s.config.Dev {
		s.autoCertDir = filepath.Join(s.config.WorkingDir, "autocert")
		err := os.MkdirAll(s.autoCertDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	s.db = GetDB("mysql", s.config.DBUri)

	var err error
	s.settings, err = objects.NewSQLSettings(s.db, bome.MySQL, "store_settings")
	if err != nil {
		return err
	}

	s.accountsManager, err = accounts.NewSQLManager(s.db, bome.MySQL, "store")
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
		Domain:     s.config.Domains[0],
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
	if !s.config.Dev {
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
		auth.DetectBasicMiddleware,
		auth.DetectOauth2Middleware,
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
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.config.Domains...),
	}
	certManager.Cache = autocert.DirCache(s.autoCertDir)

	middlewareList := []mux.MiddlewareFunc{
		auth.DetectBasicMiddleware,
		auth.DetectOauth2Middleware,
		httpx.Logger("OMS").Handle,
		s.httpEnrichContext,
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

func (s *MSServer) httpEnrichContext(next http.Handler) http.Handler {
	box := service.CreateBox(
		service.Registry(s.registry),
		service.CertFile("ca/ca.crt"),
		service.KeyFIle("ca/ca.key"),
		service.CACertFile("ca/ca.crt"),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = accounts.ContextWithManager(ctx, s.accountsManager)
		ctx = auth.ContextWithCredentialsManager(ctx, s.credentialsManager)
		ctx = auth.ContextWithProviders(ctx, s.authenticationProviders)
		ctx = service.ContextWithBox(ctx, box)
		ctx = objects.WithRouterProvider(ctx, objects.ObjectsRouterProvideFunc(s.GetRouter))
		ctx = objects.WithSettings(s.settings)(ctx)
		ctx = objects.WithRouterGrpcClientProvider(ctx, &objects.LoadBalancer{})

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *MSServer) GetRouter(ctx context.Context) objects.ObjectsRouter {
	return objects.NewCustomObjectsRouter(
		objects.NewGRPCObjectsClientHandler(objects.ServiceTypeHandler),
		objects.WithDefaultObjectsParamsHandler(),
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

	_ = s.db.Close()
	return s.caServer.Stop()
}
