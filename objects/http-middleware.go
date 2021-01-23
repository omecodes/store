package objects

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type middlewareOptions struct {
	gRPCRouterProvider GRPCRouterProvider
	acl                ACLManager
	db                 DB
	settings           SettingsManager
	routerProvider     RouterProvider
}

type MiddlewareOption func(*middlewareOptions)

func MiddlewareWithSettings(manager SettingsManager) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.settings = manager
	}
}

func MiddlewareWithRouterProvider(provider RouterProvider) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.routerProvider = provider
	}
}

func MiddlewareWithACLManager(manager ACLManager) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.acl = manager
	}
}

func MiddlewareWithDB(db DB) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.db = db
	}
}

func WithGRPCRouterProvider(provider GRPCRouterProvider) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.gRPCRouterProvider = provider
	}
}

func Middleware(opt ...MiddlewareOption) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var options middlewareOptions
			for _, o := range opt {
				o(&options)
			}

			ctx := r.Context()
			if options.acl != nil {
				ctx = context.WithValue(ctx, ctxACLStore{}, options.acl)
			}

			if options.db != nil {
				ctx = context.WithValue(ctx, ctxDB{}, options.db)
			}

			if options.routerProvider != nil {
				ctx = context.WithValue(ctx, ctxRouterProvider{}, options.routerProvider)
			}

			if options.settings != nil {
				ctx = context.WithValue(ctx, ctxSettings{}, options.settings)
			}

			if options.gRPCRouterProvider != nil {
				ctx = context.WithValue(ctx, ctxGrpcRouterClientProvider{}, options.gRPCRouterProvider)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
