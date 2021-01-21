package files

import (
	"context"
	"net/http"
)

type Middleware func(h http.Handler) http.Handler

func WithSourceManagerMiddleware(manager SourceManager) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxSourceManager{}, manager)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WithFSProviderMiddleware(provider FSProvider) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxFsProvider{}, provider)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
