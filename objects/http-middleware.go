package objects

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type middlewareOptions struct {
	db                DB
	routerProvider    RouterProvider
	acl               ACLManager
	clientProvider    ClientProvider
	aclClientProvider ACLClientProvider
}

type MiddlewareOption func(*middlewareOptions)

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

func MiddlewareWithACLClientProvider(provider ACLClientProvider) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.aclClientProvider = provider
	}
}

func WithClientProvider(provider ClientProvider) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.clientProvider = provider
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

			if options.clientProvider != nil {
				ctx = context.WithValue(ctx, ctxClientProvider{}, options.clientProvider)
			}

			if options.aclClientProvider != nil {
				ctx = context.WithValue(ctx, ctxACLClientProvider{}, options.aclClientProvider)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
