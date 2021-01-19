package oms

import (
	"encoding/json"
	"github.com/omecodes/store/objects"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/accounts"
	"github.com/omecodes/store/auth"
)

const (
	queryBefore        = "before"
	queryAfter         = "after"
	queryOffset        = "offset"
	queryCount         = "count"
	queryAt            = "at"
	queryHeader        = "header"
	pathItemId         = "id"
	pathItemName       = "name"
	pathItemCollection = "collection"
)

func NewHttpUnit() *HTTPUnit {
	return &HTTPUnit{}
}

type HTTPUnit struct{}

func (s *HTTPUnit) MuxRouter() *mux.Router {
	r := mux.NewRouter()

	r.Name("GetAccount").Methods(http.MethodGet).Path("/accounts/{name}").Handler(http.HandlerFunc(s.getAccount))
	r.Name("FindAccount").Methods(http.MethodPost).Path("/accounts").Handler(http.HandlerFunc(s.findAccount))
	r.Name("CreateAccount").Methods(http.MethodPut).Path("/accounts").Handler(http.HandlerFunc(s.createAccount))
	r.Name("CreateAccess").Methods(http.MethodPut).Path("/auth/access").Handler(http.HandlerFunc(s.createAccess))
	r.Name("SaveAuthProvider≈").Methods(http.MethodPut).Path("/auth/providers").Handler(http.HandlerFunc(s.saveProvider))
	r.Name("GetAuthProvider").Methods(http.MethodGet).Path("/auth/providers/{name}").Handler(http.HandlerFunc(s.getProvider))
	r.Name("DeleteAuthProvider").Methods(http.MethodDelete).Path("/auth/providers/{name}").Handler(http.HandlerFunc(s.deleteProvider))
	r.Name("ListProviders").Methods(http.MethodGet).Path("/auth/providers").Handler(http.HandlerFunc(s.listProviders))

	r.PathPrefix("/objects").Handler(http.StripPrefix("/objects", objects.NewHTTPRouter()))

	return r
}

func (s *HTTPUnit) saveProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.Get(ctx)

	if userInfo == nil || userInfo.Uid != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var provider *auth.Provider
	err := json.NewDecoder(r.Body).Decode(&provider)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if provider.Config == nil || provider.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	providers := auth.GetProviders(ctx)
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

func (s *HTTPUnit) getProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.Get(ctx)

	if userInfo == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	providers := auth.GetProviders(ctx)
	if providers == nil {
		logs.Error("missing providers manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	provider, err := providers.Get(vars[pathItemId])
	if err != nil {
		logs.Error("failed to get provider", logs.Details("id", vars[pathItemId]), logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	if userInfo.Uid != "admin" {
		provider.Config = nil
	}

	err = json.NewEncoder(w).Encode(provider)
	if err != nil {
		logs.Error("failed to send provider as response", logs.Err(err))
	}
}

func (s *HTTPUnit) deleteProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.Get(ctx)

	if userInfo == nil || userInfo.Uid != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	providers := auth.GetProviders(ctx)
	if providers == nil {
		logs.Error("missing providers manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	name := vars[pathItemId]
	err := providers.Delete(name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (s *HTTPUnit) listProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.Get(ctx)

	if userInfo == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	providers := auth.GetProviders(ctx)
	if providers == nil {
		logs.Error("missing providers manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	providerList, err := providers.GetAll(userInfo.Uid != "admin")
	if err != nil {
		logs.Error("failed to get provider", logs.Details("id", vars[pathItemId]), logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	err = json.NewEncoder(w).Encode(providerList)
	if err != nil {
		logs.Error("failed to send provider as response", logs.Err(err))
	}
}

func (s *HTTPUnit) createAccess(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.Get(ctx)

	if userInfo == nil || userInfo.Uid != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var access *auth.APIAccess
	err := json.NewDecoder(r.Body).Decode(&access)
	if err != nil {
		logs.Error("failed to decode request body", logs.Err(err))
		return
	}

	manager := auth.GetCredentialsManager(ctx)
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

func (s *HTTPUnit) listAccesses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.Get(ctx)

	if userInfo == nil || userInfo.Uid != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	manager := auth.GetCredentialsManager(ctx)
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

func (s *HTTPUnit) deleteAccess(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.Get(ctx)

	if userInfo == nil || userInfo.Uid != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	name := vars[pathItemId]

	manager := auth.GetCredentialsManager(ctx)
	if manager == nil {
		logs.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	access, err := manager.GetAccess(name)
	if err != nil {
		logs.Error("could not get access", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(access)
	if err != nil {
		logs.Error("failed to send provider as response", logs.Err(err))
	}
}

func (s *HTTPUnit) getAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.Get(ctx)

	if userInfo == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	name := vars[pathItemName]

	if name == "admin" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	manager := accounts.GetManager(ctx)
	account, err := manager.Get(ctx, name)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(account)
}

func (s *HTTPUnit) findAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jwt := auth.JWT(ctx)

	if jwt == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	manager := accounts.GetManager(ctx)
	account, err := manager.Find(ctx, jwt.Claims.Iss, jwt.Claims.Sub)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(account)
}

func (s *HTTPUnit) createAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jwt := auth.JWT(ctx)

	if jwt == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var account *accounts.Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	account.Source = &accounts.Source{
		Provider: jwt.Claims.Iss,
		Name:     jwt.Claims.Sub,
		Email:    jwt.Claims.Profile.Email,
	}

	manager := accounts.GetManager(ctx)
	err = manager.Create(ctx, account)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}
