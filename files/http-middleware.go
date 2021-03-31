package files

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type middlewareRouteOptions struct {
	sourceManager  SourceManager
	fsProvider     FSProvider
	routerProvider RouterProvider
}

type MiddlewareOption func(options *middlewareRouteOptions)

func Middleware(opts ...MiddlewareOption) mux.MiddlewareFunc {
	var options middlewareRouteOptions
	for _, opt := range opts {
		opt(&options)
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if options.fsProvider != nil {
				ctx = context.WithValue(ctx, ctxFsProvider{}, options.fsProvider)
			}

			if options.sourceManager != nil {
				ctx = context.WithValue(ctx, ctxSourceManager{}, options.sourceManager)
			}

			if options.routerProvider != nil {
				ctx = context.WithValue(ctx, ctxRouterProvider{}, options.routerProvider)
			}

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}

}

func MiddlewareWithSourceManager(manager SourceManager) MiddlewareOption {
	return func(options *middlewareRouteOptions) {
		options.sourceManager = manager
	}
}

func MiddlewareWithFsProvider(provider FSProvider) MiddlewareOption {
	return func(options *middlewareRouteOptions) {
		options.fsProvider = provider
	}
}

func MiddlewareWithRouterProvider(provider RouterProvider) MiddlewareOption {
	return func(options *middlewareRouteOptions) {
		options.routerProvider = provider
	}
}
