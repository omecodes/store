package auth

import (
	"encoding/json"
	"github.com/omecodes/store/session"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
)

const pathItemKey = "key"
const pathItemName = "name"

func MuxRouter(middleware ...mux.MiddlewareFunc) http.Handler {
	r := mux.NewRouter()
	r.Name("SaveAuthProvider").Methods(http.MethodPut).Path("/providers").Handler(http.HandlerFunc(SaveProvider))
	r.Name("GetAuthProvider").Methods(http.MethodGet).Path("/providers/{name}").Handler(http.HandlerFunc(GetProvider))
	r.Name("DeleteAuthProvider").Methods(http.MethodDelete).Path("/providers/{name}").Handler(http.HandlerFunc(DeleteProvider))
	r.Name("ListProviders").Methods(http.MethodGet).Path("/providers").Handler(http.HandlerFunc(ListProviders))
	r.Name("CreateAccess").Methods(http.MethodPut).Path("/access").Handler(http.HandlerFunc(CreateAccess))
	r.Name("ListAccesses").Methods(http.MethodGet).Path("/accesses").Handler(http.HandlerFunc(ListAccesses))
	r.Name("DeleteAccess").Methods(http.MethodDelete).Path("/providers/{key}").Handler(http.HandlerFunc(DeleteAccess))
	r.Name("InitSession").Methods(http.MethodPost).Path("/sessions").Handler(http.HandlerFunc(InitClientAppSession))

	var handler http.Handler
	handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func SaveProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := Get(ctx)

	if user == nil || user.Name != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var provider *Provider
	err := json.NewDecoder(r.Body).Decode(&provider)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if provider.Config == nil || provider.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	providers := GetProviders(ctx)
	if providers == nil {
		logs.Error("missing providers manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = providers.Save(provider)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func GetProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := Get(ctx)

	if user == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	providers := GetProviders(ctx)
	if providers == nil {
		logs.Error("missing providers manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	provider, err := providers.Get(vars[pathItemName])
	if err != nil {
		logs.Error("failed to get provider", logs.Details("id", vars[pathItemName]), logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	if user.Name != "admin" {
		provider.Config = nil
	}

	err = json.NewEncoder(w).Encode(provider)
	if err != nil {
		logs.Error("failed to send provider as response", logs.Err(err))
	}
}

func DeleteProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := Get(ctx)

	if user == nil || user.Name != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	providers := GetProviders(ctx)
	if providers == nil {
		logs.Error("missing providers manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	name := vars[pathItemName]
	err := providers.Delete(name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func ListProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := Get(ctx)

	if user == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	providers := GetProviders(ctx)
	if providers == nil {
		logs.Error("missing providers manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	providerList, err := providers.GetAll(user.Name != "admin")
	if err != nil {
		logs.Error("failed to get provider", logs.Details("id", vars[pathItemName]), logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	err = json.NewEncoder(w).Encode(providerList)
	if err != nil {
		logs.Error("failed to send provider as response", logs.Err(err))
	}
}

func CreateAccess(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := Get(ctx)

	if user == nil || user.Name != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var access *APIAccess
	err := json.NewDecoder(r.Body).Decode(&access)
	if err != nil {
		logs.Error("failed to decode request body", logs.Err(err))
		return
	}

	logs.Info("API save access", logs.Details("access", access))
	manager := GetCredentialsManager(ctx)
	if manager == nil {
		logs.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = manager.SaveAccess(access)
	if err != nil {
		logs.Error("failed to save access", logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func ListAccesses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := Get(ctx)

	if user == nil || user.Name != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	manager := GetCredentialsManager(ctx)
	if manager == nil {
		logs.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	accesses, err := manager.GetAllAccesses()
	if err != nil {
		logs.Error("failed to get access", logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	err = json.NewEncoder(w).Encode(accesses)
	if err != nil {
		logs.Error("failed to send provider as response", logs.Err(err))
	}
}

func DeleteAccess(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := Get(ctx)

	if user == nil || user.Name != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	name := vars[pathItemKey]

	manager := GetCredentialsManager(ctx)
	if manager == nil {
		logs.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err := manager.DeleteAccess(name)
	if err != nil {
		logs.Error("could not get access", logs.Err(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func InitClientAppSession(w http.ResponseWriter, r *http.Request) {
	var requestData *InitClientAppSessionRequest
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		logs.Error("could not decode request data", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	manager := GetCredentialsManager(r.Context())
	if manager == nil {
		logs.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	access, err := manager.GetAccess(requestData.Access.Key)
	if err != nil {
		logs.Error("could not load access", logs.Details("", requestData.Access.Key), logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	caSession, err := session.GetWebSession(session.ClientAppSession, r)
	if err != nil {
		logs.Error("could not initialize client app session", logs.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	caSession.Put(session.KeyAccessType, access.Type)
	//caSession.Put(session.KeyAccessInfo, access.Info)

	err = caSession.Save(w)
	if err != nil {
		logs.Error("could not save client app session", logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}
