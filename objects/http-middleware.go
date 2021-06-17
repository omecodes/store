package objects

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type middlewareOptions struct {
	db             DB
	routerProvider RouterProvider
	clientProvider ClientProvider
}

type MiddlewareOption func(*middlewareOptions)

func MiddlewareWithRouterProvider(provider RouterProvider) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.routerProvider = provider
	}
}

func MiddlewareWithDB(db DB) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.db = db
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

			if options.db != nil {
				ctx = context.WithValue(ctx, ctxDB{}, options.db)
			}

			if options.routerProvider != nil {
				ctx = context.WithValue(ctx, ctxRouterProvider{}, options.routerProvider)
			}

			if options.clientProvider != nil {
				ctx = context.WithValue(ctx, ctxClientProvider{}, options.clientProvider)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
