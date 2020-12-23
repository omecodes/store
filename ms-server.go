package oms

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
	"github.com/omecodes/libome/ports"
	context2 "github.com/omecodes/omestore/context"
	"github.com/omecodes/omestore/services/units"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/utils/log"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/omestore/clients"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/omestore/router"
	"github.com/omecodes/service"
	sca "github.com/omecodes/services-ca"
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
	Production   bool
}

func NewMSServer(cfg MsConfig) *MSServer {
	return &MSServer{config: cfg}
}

type MSServer struct {
	config         MsConfig
	settings       *bome.Map
	listener       net.Listener
	adminPassword  string
	workerPassword string
	Errors         chan error
	loadBalancer   *router.BaseHandler
	box            *service.Box
	registry       ome.Registry
	caServer       *sca.Server
}

func (s *MSServer) init() error {
	var err error

	s.box = service.CreateBox(
		service.Name(s.config.Name),
		service.Dir(s.config.WorkingDir),
		service.RegAddr(fmt.Sprintf("%s:%d", s.config.BindIP, s.config.RegistryPort)),
		service.Domain(s.config.Domain),
		service.Ip(s.config.BindIP),
	)

	db, err := sql.Open("mysql", s.config.DBUri)
	if err != nil {
		return err
	}

	s.settings, err = bome.NewMap(db, bome.MySQL, "settings")
	if err != nil {
		return err
	}

	return nil
}

func (s *MSServer) startRegistry() error {
	err := s.box.Options.StartRegistryServer()
	s.registry = s.box.Registry()
	return err
}

func (s *MSServer) getServiceSecret(name string) (string, error) {
	return "", nil
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

func (s *MSServer) startAPIServer() error {
	if s.config.Production {
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
		s.detectAuthentication,
		s.detectOAuth2Authorization,
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

	// redirect HTTP to HTTPS
	// CAUTION: This has to be done AFTER simplecert setup
	// Otherwise Port 80 will be blocked and cert registration fails!
	log.Info("starting HTTP Listener on Port 80")
	go func() {
		if err := http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpx.Redirect(w, &httpx.RedirectURL{
				URL:         fmt.Sprintf("https://%s:443%s", s.config.Domain, r.URL.Path),
				Code:        http.StatusPermanentRedirect,
				ContentType: "text/html",
			})
		})); err != nil {
			log.Error("listen to port 80 failed", log.Err(err))
		}
	}()

	// init strict tlsConfig with certReloadAgent
	// you could also use a default &tls.Config{}, but be warned this is highly insecure
	tlsConf := tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

	// now set GetCertificate to the reload agent GetCertificateFunc to enable hot reload
	tlsConf.GetCertificate = certReloadAgent.GetCertificateFunc()

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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context2.WithRegistry(ctx, s.registry)
		ctx = router.WithRouterProvider(ctx, router.ProviderFunc(s.GetRouter))
		ctx = router.WithSettings(s.settings)(ctx)
		ctx = service.ContextWithBox(ctx, s.box)
		ctx = clients.WithUnitClientProvider(ctx, &clients.LoadBalancer{})

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *MSServer) GetRouter(ctx context.Context) router.Router {
	return router.NewCustomRouter(
		units.NewGRPCClientHandler(),
		router.WithDefaultParamsHandler(),
	)
}

func (s *MSServer) detectAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
					if pass != s.workerPassword {
						w.WriteHeader(http.StatusForbidden)
						return
					}
				}

				ctx := router.WithUserInfo(r.Context(), &pb.Auth{
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

func (s *MSServer) detectOAuth2Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("authorization")
		if authorization != "" && strings.HasPrefix(authorization, "Bearer ") {
			authorization = strings.TrimPrefix(authorization, "Bearer ")
			jwt, err := ome.ParseJWT(authorization)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			signature, err := jwt.SecretBasedSignature(s.config.JWTSecret)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if signature != jwt.Signature {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			ctx := router.WithUserInfo(r.Context(), &pb.Auth{
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
	s.box.Stop()
	return s.box.StopRegistry()
}
