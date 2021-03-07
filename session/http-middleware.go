package session

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"net/http"
)

func WithHTTPSessionMiddleware(store *sessions.CookieStore) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxCookieStore{}, store)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
