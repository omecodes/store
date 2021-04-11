package common

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
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

func MiddlewareWithSettings(manager SettingsManager) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updatedContext := ContextWithSettings(r.Context(), manager)
			next.ServeHTTP(w, r.WithContext(updatedContext))
		})
	}
}

func ErrorFromHttpResponse(rsp *http.Response) error {
	var err error

	if rsp.StatusCode != 200 {

		switch rsp.StatusCode {

		case http.StatusBadRequest:
			err = errors.BadRequest(rsp.Status)

		case http.StatusInternalServerError:
			err = errors.Internal(rsp.Status)

		case http.StatusBadGateway:
			err = errors.ServiceUnavailable(rsp.Status)

		case http.StatusForbidden:
			err = errors.Forbidden(rsp.Status)

		case http.StatusUnauthorized:
			err = errors.Unauthorized(rsp.Status)

		case http.StatusNotFound:
			err = errors.NotFound(rsp.Status)

		default:
			err = errors.New(rsp.Status)
		}

		if rsp.ContentLength > 0 {
			body, _ := ioutil.ReadAll(rsp.Body)
			err.(*errors.Error).AddDetails("content", string(body))
		}
	}
	return err
}
