package oms

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/accounts"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/router"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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

	r.Name("SetSettings").Methods(http.MethodPut).Path("/settings").Handler(http.HandlerFunc(s.setSettings))
	r.Name("GetSettings").Methods(http.MethodGet).Path("/settings").Handler(http.HandlerFunc(s.getSettings))

	r.Name("GetAccount").Methods(http.MethodGet).Path("/accounts/{name}").Handler(http.HandlerFunc(s.getAccount))
	r.Name("FindAccount").Methods(http.MethodPost).Path("/accounts").Handler(http.HandlerFunc(s.findAccount))
	r.Name("CreateAccount").Methods(http.MethodPut).Path("/accounts").Handler(http.HandlerFunc(s.createAccount))

	r.Name("SaveAuthProvider≈").Methods(http.MethodPut).Path("/auth/providers").Handler(http.HandlerFunc(s.saveProvider))
	r.Name("GetAuthProvider").Methods(http.MethodGet).Path("/auth/providers/{name}").Handler(http.HandlerFunc(s.getProvider))
	r.Name("DeleteAuthProvider").Methods(http.MethodDelete).Path("/auth/providers/{name}").Handler(http.HandlerFunc(s.deleteProvider))
	r.Name("ListProviders").Methods(http.MethodGet).Path("/auth/providers").Handler(http.HandlerFunc(s.listProviders))

	r.Name("CreateAccess").Methods(http.MethodPut).Path("/auth/access").Handler(http.HandlerFunc(s.createAccess))

	r.Name("CreateCollection").Methods(http.MethodPut).Path("/collections").Handler(http.HandlerFunc(s.createCollection))
	r.Name("ListCollections").Methods(http.MethodGet).Path("/collections").Handler(http.HandlerFunc(s.listCollections))
	r.Name("DeleteCollection").Methods(http.MethodGet).Path("/collections/{id}").Handler(http.HandlerFunc(s.deleteCollection))
	r.Name("GetCollection").Methods(http.MethodGet).Path("/collections/{id}").Handler(http.HandlerFunc(s.listProviders))

	r.Name("PutObject").Methods(http.MethodPut).Path("/objects/{collection}").Handler(http.HandlerFunc(s.put))
	r.Name("PatchObject").Methods(http.MethodPatch).Path("/objects/{collection}/{id}").Handler(http.HandlerFunc(s.patch))
	r.Name("MoveObject").Methods(http.MethodPost).Path("/objects/{collection}/{id}").Handler(http.HandlerFunc(s.patch))
	r.Name("GetObject").Methods(http.MethodGet).Path("/objects/{collection}/{id}").Handler(http.HandlerFunc(s.get))
	r.Name("DeleteObject").Methods(http.MethodDelete).Path("/objects/{collection}/{id}").Handler(http.HandlerFunc(s.del))
	r.Name("GetObjects").Methods(http.MethodGet).Path("/objects/{collection}").Handler(http.HandlerFunc(s.list))
	r.Name("SearchObjects").Methods(http.MethodPost).Path("/objects/{collection}").Handler(http.HandlerFunc(s.search))

	return r
}

func (s *HTTPUnit) put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var putRequest pb.PutObjectRequest
	err := jsonpb.Unmarshal(r.Body, &putRequest)
	if err != nil {
		log.Error("failed to decode request body", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if putRequest.Object.Header == nil {
		putRequest.Object.Header = &pb.Header{}
	}
	putRequest.Object.Header.Size = int64(len(putRequest.Object.Data))

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := route.PutObject(ctx, collection, putRequest.Object, putRequest.AccessSecurityRules, putRequest.Indexes, pb.PutOptions{})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write([]byte(fmt.Sprintf("{\"id\": \"%s\"}", id)))
}

func (s *HTTPUnit) patch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var patch pb.Patch
	err := jsonpb.Unmarshal(r.Body, &patch)
	if err != nil {
		log.Error("failed to decode request body", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]
	patch.ObjectId = vars[pathItemId]

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = route.PatchObject(ctx, collection, &patch, pb.PatchOptions{})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func (s *HTTPUnit) move(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var request pb.MoveObjectRequest
	err := jsonpb.Unmarshal(r.Body, &request)
	if err != nil {
		log.Error("failed to decode request body", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]
	objectId := vars[pathItemId]

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = route.MoveObject(ctx, collection, objectId, request.TargetCollection, request.AccessSecurityRules, pb.MoveOptions{})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func (s *HTTPUnit) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]
	id := vars[pathItemId]

	header := r.URL.Query().Get(queryHeader)
	at := r.URL.Query().Get(queryAt)

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	object, err := route.GetObject(ctx, collection, id, pb.GetOptions{
		At:   at,
		Info: header == "true",
	})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write([]byte(object.Data))
	if err != nil {
		log.Error("failed to write response", log.Err(err))
	}
}

func (s *HTTPUnit) del(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]
	id := vars[pathItemId]

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = route.DeleteObject(ctx, collection, id)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func (s *HTTPUnit) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		err  error
		opts pb.ListOptions
	)

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]

	opts.Offset, err = Int64QueryParam(r, queryOffset)
	if err != nil {
		log.Error("could not parse param 'before'")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	opts.At = r.URL.Query().Get(queryAt)

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cursor, err := route.ListObjects(ctx, collection, opts)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	defer func() {
		if cErr := cursor.Close(); cErr != nil {
			log.Error("cursor closed with an error", log.Err(cErr))
		}
	}()
	w.Header().Add("Content-Type", "application/json")

	_, err = w.Write([]byte("{"))
	position := 0
	for {
		object, err2 := cursor.Browse()
		if err2 != nil {
			if err2 == io.EOF {
				break
			}
			w.WriteHeader(errors.HTTPStatus(err2))
			return
		}

		var item string
		if position == 0 {
			position++
		} else {
			item = ","
		}

		item = item + fmt.Sprintf("\"%s\": %s", object.Header.Id, object.Data)
		_, err = w.Write([]byte(item))
		if err != nil {
			log.Error("GetObjects: failed to write result item", log.Err(err))
			return
		}
	}
	_, err = w.Write([]byte("}"))
}

func (s *HTTPUnit) search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var query pb.SearchQuery
	err := jsonpb.Unmarshal(r.Body, &query)
	if err != nil {
		log.Error("could not parse search query")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cursor, err := route.SearchObjects(ctx, collection, &query)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
	defer func() {
		if cErr := cursor.Close(); cErr != nil {
			log.Error("cursor closed with an error", log.Err(cErr))
		}
	}()
	w.Header().Add("Content-Type", "application/json")

	_, err = w.Write([]byte("{"))
	position := 0
	for {
		object, err2 := cursor.Browse()
		if err2 != nil {
			if err2 == io.EOF {
				break
			}
			w.WriteHeader(errors.HTTPStatus(err2))
			_, err = w.Write([]byte("}"))
			return
		}

		var item string
		if position == 0 {
			position++
		} else {
			item = ","
		}

		item = item + fmt.Sprintf("\"%s\": %s", object.Header.Id, object.Data)
		_, err = w.Write([]byte(item))
		if err != nil {
			_, err = w.Write([]byte("}"))
			log.Error("GetObjects: failed to write result item", log.Err(err))
			return
		}
	}
	_, err = w.Write([]byte("}"))
}

func (s *HTTPUnit) setSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ai := auth.Get(ctx)
	if ai == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if ai.Uid != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	name := r.URL.Query().Get("name")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("could not read request body", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	settings := router.Settings(ctx)
	if settings == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = settings.Set(name, string(data))
	if err != nil {
		log.Error("failed to set settings", log.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
	}
}

func (s *HTTPUnit) getSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.URL.Query().Get("name")

	ai := auth.Get(ctx)
	if ai == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if ai.Uid != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	settings := router.Settings(ctx)
	if settings == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	value, err := settings.Get(name)
	if err != nil {
		log.Error("failed to set settings", log.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
	}

	w.Header().Add("Content-Type", "text/plain")
	_, _ = w.Write([]byte(value))
}

func (s *HTTPUnit) createCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var collection pb.Collection
	err := json.NewDecoder(r.Body).Decode(&collection)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = route.CreateCollection(ctx, &collection)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func (s *HTTPUnit) listCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	collections, err := route.ListCollections(ctx)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if collections == nil {
		_, _ = w.Write([]byte("[]"))
		return
	}

	data, err := json.Marshal(collections)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(data)
}

func (s *HTTPUnit) getCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	id := vars[pathItemId]

	collection, err := route.GetCollection(ctx, id)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")

	data, err := json.Marshal(collection)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(data)
}

func (s *HTTPUnit) deleteCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	id := vars[pathItemId]

	err = route.DeleteCollection(ctx, id)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
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
		log.Error("missing providers manager in context")
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
		log.Error("missing providers manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	provider, err := providers.Get(vars[pathItemId])
	if err != nil {
		log.Error("failed to get provider", log.Field("id", vars[pathItemId]), log.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	if userInfo.Uid != "admin" {
		provider.Config = nil
	}

	err = json.NewEncoder(w).Encode(provider)
	if err != nil {
		log.Error("failed to send provider as response", log.Err(err))
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
		log.Error("missing providers manager in context")
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
		log.Error("missing providers manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	providerList, err := providers.GetAll(userInfo.Uid != "admin")
	if err != nil {
		log.Error("failed to get provider", log.Field("id", vars[pathItemId]), log.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	err = json.NewEncoder(w).Encode(providerList)
	if err != nil {
		log.Error("failed to send provider as response", log.Err(err))
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
		log.Error("failed to decode request body", log.Err(err))
		return
	}

	manager := auth.GetCredentialsManager(ctx)
	if manager == nil {
		log.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = manager.SaveAccess(access)
	if err != nil {
		log.Error("failed to save access", log.Err(err))
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
		log.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	accesses, err := manager.GetAllAccesses()
	if err != nil {
		log.Error("failed to get access", log.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	err = json.NewEncoder(w).Encode(accesses)
	if err != nil {
		log.Error("failed to send provider as response", log.Err(err))
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
		log.Error("missing credentials manager in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	access, err := manager.GetAccess(name)
	if err != nil {
		log.Error("could not get access", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(access)
	if err != nil {
		log.Error("failed to send provider as response", log.Err(err))
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

func Int64QueryParam(r *http.Request, name string) (int64, error) {
	beforeParam := r.URL.Query().Get(name)
	if beforeParam != "" {
		return strconv.ParseInt(beforeParam, 10, 64)
	} else {
		return 0, nil
	}
}
