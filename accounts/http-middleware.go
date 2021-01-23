package accounts

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type middlewareOptions struct {
	manager Manager
}

type MiddlewareOption func(*middlewareOptions)

func MiddlewareWithAccountManager(manager Manager) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.manager = manager
	}
}

func Middleware(opts ...MiddlewareOption) mux.MiddlewareFunc {
	var options middlewareOptions
	for _, opt := range opts {
		opt(&options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updatedContext := context.WithValue(r.Context(), ctxManager{}, options.manager)
			r = r.WithContext(updatedContext)
			next.ServeHTTP(w, r)
		})
	}
}
