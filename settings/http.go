package settings

import (
	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common"
	"io/ioutil"
	"net/http"
)

func MiddlewareWithManager(manager Manager) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updatedContext := ContextWithManager(r.Context(), manager)
			next.ServeHTTP(w, r.WithContext(updatedContext))
		})
	}
}

func MuxRouter(middleware ...mux.MiddlewareFunc) http.Handler {
	r := mux.NewRouter()
	r.Name("SetSettings").Methods(http.MethodPut).Path(common.ApiSetSettingsRoute).Handler(http.HandlerFunc(HTTPHandleSetSettings))
	r.Name("GetSettings").Methods(http.MethodGet).Path(common.ApiGetSettingsRoute).Handler(http.HandlerFunc(HTTPHandleGetSettings))
	return r
}

func HTTPHandleSetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := auth.Get(ctx)
	if user == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if user.Name != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	name := r.URL.Query().Get(common.ApiParamName)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logs.Error("could not read request body", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	settingsManager := GetManager(ctx)
	if settingsManager == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = settingsManager.Set(name, string(data))
	if err != nil {
		logs.Error("failed to set settings", logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
	}
}

func HTTPHandleGetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.URL.Query().Get(common.ApiParamName)

	user := auth.Get(ctx)
	if user == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if user.Name != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	settingsManager := GetManager(ctx)
	if settingsManager == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	value, err := settingsManager.Get(name)
	if err != nil {
		logs.Error("failed to set settings", logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
	}

	w.Header().Add(common.HttpHeaderContentType, "text/plain")
	_, _ = w.Write([]byte(value))
}
