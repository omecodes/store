package auth

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/session"
)

const pathItemKey = "key"
const pathItemName = "name"
const formValueUsername = "username"
const formValuePassword = "password"
const queryContinueURL = "continue"

func MuxRouter(middleware ...mux.MiddlewareFunc) http.Handler {
	r := mux.NewRouter()
	r.Name("SaveAuthProvider").Methods(http.MethodPut).Path("/providers").Handler(http.HandlerFunc(SaveProvider))
	r.Name("GetAuthProvider").Methods(http.MethodGet).Path("/providers/{name}").Handler(http.HandlerFunc(GetProvider))
	r.Name("DeleteAuthProvider").Methods(http.MethodDelete).Path("/providers/{name}").Handler(http.HandlerFunc(DeleteProvider))
	r.Name("ListProviders").Methods(http.MethodGet).Path("/providers").Handler(http.HandlerFunc(ListProviders))
	r.Name("CreateAccess").Methods(http.MethodPut).Path("/access").Handler(http.HandlerFunc(CreateAccess))
	r.Name("ListAccesses").Methods(http.MethodGet).Path("/accesses").Handler(http.HandlerFunc(ListAccesses))
	r.Name("DeleteAccess").Methods(http.MethodDelete).Path("/providers/{key}").Handler(http.HandlerFunc(DeleteAccess))
	r.Name("InitWebClientSession").Methods(http.MethodPost).Path("/sessions/client-app").Handler(http.HandlerFunc(InitClientAppSession))
	r.Name("CreateUser").Methods(http.MethodPut).Path("/users").Handler(http.HandlerFunc(SaveUser))
	r.Name("SearchUser").Methods(http.MethodGet).Path("/users").Handler(http.HandlerFunc(SearchUsers))

	var handler http.Handler
	handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func UserSessionHandler(middleware ...mux.MiddlewareFunc) http.Handler {
	var handler http.Handler
	handler = http.HandlerFunc(CreateUserWebSession)
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func redirectToLocation(w http.ResponseWriter, status int, location string, params url.Values) {
	if params != nil {
		location += "&" + params.Encode()
	}

	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("<head>\n"))
	b.WriteString(fmt.Sprintf("\t<meta http-equiv=\"refresh\" content=\"0; URL=%s\" />\n", location))
	b.WriteString(fmt.Sprintf("</head>"))
	contentBytes := []byte(b.String())

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Location", location)
	w.WriteHeader(status)
	_, _ = w.Write(contentBytes)
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

	var access *ClientApp
	err := json.NewDecoder(r.Body).Decode(&access)
	if err != nil {
		logs.Error("failed to decode request body", logs.Err(err))
		return
	}

	manager := GetCredentialsManager(ctx)
	if manager == nil {
		logs.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = manager.SaveClientApp(access)
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

	accesses, err := manager.GetAllClientApps()
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

	err := manager.DeleteClientApp(name)
	if err != nil {
		logs.Error("could not get access", logs.Err(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func SaveUser(w http.ResponseWriter, r *http.Request) {
	var user *UserCredentials

	ctx := r.Context()
	requester := Get(ctx)

	if requester == nil || requester.Name != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		logs.Error("could not decode request data")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" {
		logs.Error("username or password is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	manager := GetCredentialsManager(ctx)
	if manager == nil {
		logs.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hashedBytes := sha512.Sum512([]byte(user.Password))
	user.Password = hex.EncodeToString(hashedBytes[:])

	err = manager.SaveUserCredentials(user)
	if err != nil {
		logs.Error("could not save user", logs.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func SearchUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requester := Get(ctx)
	if requester == nil || requester.Name == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	manager := GetCredentialsManager(ctx)
	if manager == nil {
		logs.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func CreateUserWebSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	clientApp := App(ctx)
	if clientApp == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if clientApp.Type != ClientType_web {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	manager := GetCredentialsManager(ctx)
	if manager == nil {
		logs.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	continueURL := r.URL.Query().Get(queryContinueURL)
	if continueURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := r.ParseForm()
	if err != nil {
		logs.Error("could not parse form", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	username := r.Form.Get(formValueUsername)
	password := r.Form.Get(formValuePassword)

	logs.Info("login", logs.Details("username", username), logs.Details("password", password))

	loadedPassword, err := manager.GetUserPassword(username)
	if err != nil {
		if errors.IsNotFound(err) {
			redirectToLocation(w, http.StatusForbidden, continueURL, nil)
		} else {
			redirectToLocation(w, errors.HTTPStatus(err), continueURL, nil)
		}
		return
	}

	hashedBytes := sha512.Sum512([]byte(password))
	password = hex.EncodeToString(hashedBytes[:])

	if loadedPassword != password {
		redirectToLocation(w, http.StatusForbidden, continueURL, nil)
		return
	}

	userSession, err := session.GetWebSession(session.UserSession, r)
	if err != nil {
		logs.Error("could not initialize client app session", logs.Err(err))
		redirectToLocation(w, errors.HTTPStatus(err), continueURL, nil)
		return
	}

	userSession.Put(session.KeyUsername, username)
	err = userSession.Save(w)
	if err != nil {
		logs.Error("could not save client app session", logs.Err(err))
		redirectToLocation(w, errors.HTTPStatus(err), continueURL, nil)
		return
	}

	redirectToLocation(w, http.StatusOK, continueURL, nil)
}

func InitClientAppSession(w http.ResponseWriter, r *http.Request) {
	var requestData *InitClientAppSessionRequest
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		logs.Error("could not decode request data", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if requestData.ClientApp == nil {
		logs.Error("missing client app data in request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	manager := GetCredentialsManager(r.Context())
	if manager == nil {
		logs.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	clientApp, err := manager.GetClientApp(requestData.ClientApp.Key)
	if err != nil {
		logs.Error("could not load access", logs.Details("", requestData.ClientApp.Key), logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	if clientApp.Secret != requestData.ClientApp.Secret {
		logs.Error("client app authentication failed", logs.Err(err))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	caSession, err := session.GetWebSession(session.ClientAppSession, r)
	if err != nil {
		logs.Error("could not initialize client app session", logs.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if clientApp.Info != nil {
		encodedInfo, err := json.Marshal(clientApp.Info)
		if err != nil {
			logs.Error("could not encode client app info", logs.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		caSession.Put(session.KeyAccessInfo, string(encodedInfo))
	}

	caSession.Put(session.KeyAccessType, clientApp.Type.String())
	caSession.Put(session.KeyAccessKey, clientApp.Key)

	err = caSession.Save(w)
	if err != nil {
		logs.Error("could not save client app session", logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}
