package oms

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/router"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	queryBefore        = "before"
	queryAfter         = "after"
	queryCount         = "count"
	queryAt            = "at"
	queryHeader        = "header"
	pathItemId         = "id"
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

	settingsSubRouter := r.PathPrefix("/settings/").Subrouter()
	settingsSubRouter.Name("SetSettings").Methods(http.MethodPost).Handler(http.HandlerFunc(s.setSettings))
	settingsSubRouter.Name("GetSettings").Methods(http.MethodGet).Handler(http.HandlerFunc(s.getSettings))

	r.Name("CreateCollection").Methods(http.MethodPut).Path("/collections").Handler(http.HandlerFunc(s.createCollection))
	r.Name("ListCollections").Methods(http.MethodGet).Path("/collections").Handler(http.HandlerFunc(s.listCollections))
	r.Name("DeleteCollection").Methods(http.MethodGet).Path("/collections/{id}").Handler(http.HandlerFunc(s.deleteCollection))
	r.Name("GetCollection").Methods(http.MethodGet).Path("/collections/{id}").Handler(http.HandlerFunc(s.getCollection))

	r.Name("Put").Methods(http.MethodPut).Path("/objects/{collection}").Handler(http.HandlerFunc(s.put))
	r.Name("Patch").Methods(http.MethodPatch).Path("/objects/{collection}/{id}").Handler(http.HandlerFunc(s.patch))
	r.Name("Get").Methods(http.MethodGet).Path("/objects/{collection}/{id}").Handler(http.HandlerFunc(s.get))
	r.Name("Del").Methods(http.MethodDelete).Path("/objects/{collection}/{id}").Handler(http.HandlerFunc(s.del))
	r.Name("GetObjects").Methods(http.MethodGet).Path("/objects/{collection}").Handler(http.HandlerFunc(s.list))
	r.Name("Search").Methods(http.MethodPost).Path("/objects/{collection}").Handler(http.HandlerFunc(s.search))

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
		w.WriteHeader(errors.HttpStatus(err))
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
		w.WriteHeader(errors.HttpStatus(err))
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
		w.WriteHeader(errors.HttpStatus(err))
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
		w.WriteHeader(errors.HttpStatus(err))
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

	opts.DateOptions.Before, err = Int64QueryParam(r, queryBefore)
	if err != nil {
		log.Error("could not parse param 'before'")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	opts.DateOptions.After, err = Int64QueryParam(r, queryAfter)
	if err != nil {
		log.Error("could not parse param 'after'")
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
		w.WriteHeader(errors.HttpStatus(err))
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
			w.WriteHeader(errors.HttpStatus(err2))
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

	var query pb.BooleanExp
	err := jsonpb.Unmarshal(r.Body, &query)
	if err != nil {
		log.Error("could not parse param 'before'")
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
		w.WriteHeader(errors.HttpStatus(err))
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
			w.WriteHeader(errors.HttpStatus(err2))
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
		w.WriteHeader(errors.HttpStatus(err))
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
		w.WriteHeader(errors.HttpStatus(err))
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
		w.WriteHeader(errors.HttpStatus(err))
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
		w.WriteHeader(errors.HttpStatus(err))
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
		w.WriteHeader(errors.HttpStatus(err))
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
		w.WriteHeader(errors.HttpStatus(err))
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
