package accounts

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/auth"
	"net/http"
	"strings"
)

const pathItemName = "name"

func GetMiddleware(manager Manager) mux.MiddlewareFunc {
	return middleware(manager)
}

func NewHttpHandler() (handler http.Handler) {
	r := mux.NewRouter()
	r.Name("GetAccount").Methods(http.MethodGet).Path("/accounts/{name}").Handler(http.HandlerFunc(getAccount))
	r.Name("FindAccount").Methods(http.MethodPost).Path("/accounts").Handler(http.HandlerFunc(findAccount))
	r.Name("CreateAccount").Methods(http.MethodPut).Path("/accounts").Handler(http.HandlerFunc(createAccount))
	handler = r
	return
}

func findAccount(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(account)
}

func createAccount(w http.ResponseWriter, r *http.Request) {
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

func getAccount(w http.ResponseWriter, r *http.Request) {
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

	manager := GetManager(ctx)
	account, err := manager.Get(ctx, name)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(account)
}

func middleware(manager Manager) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updatedContext := context.WithValue(r.Context(), ctxManager{}, manager)
			r = r.WithContext(updatedContext)
			next.ServeHTTP(w, r)
		})
	}
}
