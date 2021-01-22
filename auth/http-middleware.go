package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
)

type middlewareOptions struct {
	credentials CredentialsManager
	providers   ProviderManager
}

type MiddlewareOption func(*middlewareOptions)

func MiddlewareWithProviderManager(manager ProviderManager) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.providers = manager
	}
}

func MiddlewareWithCredentials(manager CredentialsManager) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.credentials = manager
	}
}

func Middleware(opts ...MiddlewareOption) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		next = detectBasic(next)
		next = detectOauth2(next)
		next = detectProxyBasic(next)

		var options middlewareOptions
		for _, opt := range opts {
			opt(&options)
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updatedContext := r.Context()
			if options.credentials != nil {
				updatedContext = context.WithValue(updatedContext, ctxCredentialsManager{}, options.credentials)
			}

			if options.providers != nil {
				updatedContext = context.WithValue(r.Context(), ctxAuthenticationProviders{}, options.providers)
			}

			next.ServeHTTP(w, r.WithContext(updatedContext))
		})
	}
}

func detectBasic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		if authorization != "" {
			authorizationParts := strings.SplitN(authorization, " ", 2)
			authType := strings.ToLower(authorizationParts[0])
			if len(authorizationParts) > 1 {
				authorization = authorizationParts[1]
			}

			if authType == "basic" {
				if authorization == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				ctx := r.Context()
				ctx, err := updateContextWithBasic(ctx, authorization)
				if err != nil {
					if err2, ok := err.(*errors.Error); ok {
						w.WriteHeader(err2.Code)
						return
					}
					w.WriteHeader(http.StatusForbidden)
					return
				}
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func detectProxyBasic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Proxy-Authorization")

		if authorization != "" {
			authorizationParts := strings.SplitN(authorization, " ", 2)
			authType := strings.ToLower(authorizationParts[0])
			if len(authorizationParts) > 1 {
				authorization = authorizationParts[1]
			}

			if authType == "basic" {
				if authorization == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				ctx := r.Context()
				ctx, err := updateContextWithProxyBasic(ctx, authorization)
				if err != nil {
					if err2, ok := err.(*errors.Error); ok {
						w.WriteHeader(err2.Code)
						return
					}
					w.WriteHeader(http.StatusForbidden)
					return
				}
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func detectOauth2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization != "" {
			authorizationParts := strings.SplitN(authorization, " ", 2)
			authType := strings.ToLower(authorizationParts[0])
			if len(authorizationParts) > 1 {
				authorization = authorizationParts[1]
			}

			if authType == "bearer" {
				if authorization == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				ctx, err := updateContextWithOauth2(r.Context(), authorization)
				if err != nil {
					if err2, ok := err.(*errors.Error); ok {
						w.WriteHeader(err2.Code)
						return
					}
					w.WriteHeader(http.StatusForbidden)
					return
				}
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}
