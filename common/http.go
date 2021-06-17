package common

import (
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

func AllowCORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HttpHeaderAccessControlAllowOrigin, "*")
		next.ServeHTTP(w, r)
	})
}
