package accounts

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common"
	"net/http"
	"strings"
)

func MuxRouter(middleware ...mux.MiddlewareFunc) http.Handler {
	r := mux.NewRouter()
	r.Name("GetAccount").Methods(http.MethodGet).Path(common.ApiGetAccountRoute).Handler(http.HandlerFunc(GetAccount))
	r.Name("FindAccount").Methods(http.MethodPost).Path(common.ApiFindAccountRoute).Handler(http.HandlerFunc(FindAccount))
	r.Name("CreateAccount").Methods(http.MethodPut).Path(common.ApiCreateAccountRoute).Handler(http.HandlerFunc(CreateAccount))
	var handler http.Handler
	handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func FindAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jwt := auth.JWT(ctx)

	if jwt == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	manager := GetManager(ctx)
	account, err := manager.Find(ctx, jwt.Claims.Iss, jwt.Claims.Sub)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add(common.HttpHeaderContentType, common.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(account)
}

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jwt := auth.JWT(ctx)

	if jwt == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	contentType := r.Header.Get(common.HttpHeaderContentType)
	if !strings.HasPrefix(contentType, common.ContentTypeJSON) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var account *Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	account.Source = &Source{
		Provider: jwt.Claims.Iss,
		Name:     jwt.Claims.Sub,
		Email:    jwt.Claims.Profile.Email,
	}

	manager := GetManager(ctx)
	err = manager.Create(ctx, account)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}

func GetAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.Get(ctx)

	if userInfo == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	name := vars[common.ApiRouteVarIdName]

	if name == "admin" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	manager := GetManager(ctx)
	account, err := manager.Get(ctx, name)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add(common.HttpHeaderContentType, common.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(account)
}
