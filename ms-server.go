package oms

import (
	"context"
	"database/sql"
	"encoding/base64"
	"github.com/gorilla/mux"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/discover"
	ome "github.com/omecodes/libome"
	net2 "github.com/omecodes/libome/net"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/omestore/router"
	"github.com/omecodes/service"
	"net"
	"net/http"
	"strings"
)

type MsConfig struct {
	Box             *service.Box
	Address         string
	JWTSecret       string
	DBUri           string
	TlsCertFilename string
	TlsKeyFilename  string
}

type MSServer struct {
	config         *MsConfig
	settings       *bome.Map
	registry       discover.Server
	ca             *service.Box
	provider       router.Provider
	listener       net.Listener
	adminPassword  string
	workerPassword string
	Errors         chan error
	loadBalancer   *router.BaseHandler
}

func (s *MSServer) init() error {
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

func (s *MSServer) initLoadBalancer() error {
	return nil
}

func (s *MSServer) startRegistry() error {
	return nil
}

func (s *MSServer) startCA() error {
	return nil
}

func (s *MSServer) startAPIServer() error {
	err := s.init()
	if err != nil {
		return err
	}

	var opts []net2.ListenOption
	if s.config.TlsCertFilename != "" {
		opts = append(opts, net2.WithTLSParams(s.config.TlsCertFilename, s.config.TlsKeyFilename))
	}

	s.listener, err = net2.Listen(s.config.Address, opts...)
	if err != nil {
		return err
	}

	address := s.listener.Addr().String()
	log.Info("starting HTTP server", log.Field("address", address))

	middlewareList := []mux.MiddlewareFunc{
		s.enrichHTTPContextHandler,
		s.detectAuthentication,
		s.detectOAuth2Authorization,
		httpx.Logger("oms").Handle,
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

func (s *MSServer) GetRouter(ctx context.Context) router.Router {
	customRouter := router.NewCustomRouter(nil)
	return customRouter
}

func (s *MSServer) enrichHTTPContextHandler(http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
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
	return nil
}

func (s *MSServer) Stop() error {
	return nil
}
