package common

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func Int64QueryParam(r *http.Request, name string) (int64, error) {
	beforeParam := r.URL.Query().Get(name)
	if beforeParam != "" {
		return strconv.ParseInt(beforeParam, 10, 64)
	} else {
		return 0, nil
	}
}

func MiddlewareWithSettings(manager SettingsManager) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updatedContext := ContextWithSettings(r.Context(), manager)
			next.ServeHTTP(w, r.WithContext(updatedContext))
		})
	}
}
