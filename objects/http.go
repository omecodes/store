package objects

import (
	"encoding/json"
	"fmt"
	se "github.com/omecodes/store/search-engine"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/auth"
)

const (
	queryOffset        = "offset"
	queryAt            = "at"
	queryHeader        = "header"
	pathItemId         = "id"
	pathItemCollection = "collection"
)

func NewHTTPRouter() *mux.Router {
	r := mux.NewRouter()

	r.Name("SetSettings").Methods(http.MethodPut).Path("/settings").Handler(http.HandlerFunc(setSettings))
	r.Name("GetSettings").Methods(http.MethodGet).Path("/settings").Handler(http.HandlerFunc(getSettings))

	r.Name("CreateCollection").Methods(http.MethodPut).Path("/collections").Handler(http.HandlerFunc(createCollection))
	r.Name("ListCollections").Methods(http.MethodGet).Path("/collections").Handler(http.HandlerFunc(listCollections))
	r.Name("DeleteCollection").Methods(http.MethodGet).Path("/collections/{id}").Handler(http.HandlerFunc(deleteCollection))
	r.Name("GetCollection").Methods(http.MethodGet).Path("/collections/{id}").Handler(http.HandlerFunc(getCollection))

	r.Name("PutObject").Methods(http.MethodPut).Path("/data/{collection}").Handler(http.HandlerFunc(putObject))
	r.Name("PatchObject").Methods(http.MethodPatch).Path("/data/{collection}/{id}").Handler(http.HandlerFunc(patchObject))
	r.Name("MoveObject").Methods(http.MethodPost).Path("/data/{collection}/{id}").Handler(http.HandlerFunc(moveObject))
	r.Name("GetObject").Methods(http.MethodGet).Path("/data/{collection}/{id}").Handler(http.HandlerFunc(getObject))
	r.Name("DeleteObject").Methods(http.MethodDelete).Path("/data/{collection}/{id}").Handler(http.HandlerFunc(deleteObject))
	r.Name("GetObjects").Methods(http.MethodGet).Path("/data/{collection}").Handler(http.HandlerFunc(listObjects))
	r.Name("SearchObjects").Methods(http.MethodPost).Path("/data/{collection}").Handler(http.HandlerFunc(searchObjects))

	return r
}

func putObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var putRequest PutObjectRequest
	err := jsonpb.Unmarshal(r.Body, &putRequest)
	if err != nil {
		log.Error("failed to decode request body", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if putRequest.Object.Header == nil {
		putRequest.Object.Header = &Header{}
	}
	putRequest.Object.Header.Size = int64(len(putRequest.Object.Data))

	route, err := NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := route.PutObject(ctx, collection, putRequest.Object, putRequest.AccessSecurityRules, putRequest.Indexes, PutOptions{})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write([]byte(fmt.Sprintf("{\"id\": \"%s\"}", id)))
}

func patchObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var patch Patch
	err := jsonpb.Unmarshal(r.Body, &patch)
	if err != nil {
		log.Error("failed to decode request body", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]
	patch.ObjectId = vars[pathItemId]

	route, err := NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = route.PatchObject(ctx, collection, &patch, PatchOptions{})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func moveObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var request MoveObjectRequest
	err := jsonpb.Unmarshal(r.Body, &request)
	if err != nil {
		log.Error("failed to decode request body", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]
	objectId := vars[pathItemId]

	route, err := NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = route.MoveObject(ctx, collection, objectId, request.TargetCollection, request.AccessSecurityRules, MoveOptions{})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func getObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]
	id := vars[pathItemId]

	header := r.URL.Query().Get(queryHeader)
	at := r.URL.Query().Get(queryAt)

	route, err := NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	object, err := route.GetObject(ctx, collection, id, GetOptions{
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

func deleteObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]
	id := vars[pathItemId]

	route, err := NewRoute(ctx)
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

func listObjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		err  error
		opts ListOptions
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

	route, err := NewRoute(ctx)
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

func searchObjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var query se.SearchQuery
	err := jsonpb.Unmarshal(r.Body, &query)
	if err != nil {
		log.Error("could not parse search query")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	collection := vars[pathItemCollection]

	route, err := NewRoute(ctx)
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

func setSettings(w http.ResponseWriter, r *http.Request) {
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

	settings := Settings(ctx)
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

func getSettings(w http.ResponseWriter, r *http.Request) {
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

	settings := Settings(ctx)
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

func createCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var collection Collection
	err := json.NewDecoder(r.Body).Decode(&collection)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	route, err := NewRoute(ctx)
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

func listCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	route, err := NewRoute(ctx)
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

func getCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	route, err := NewRoute(ctx)
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

func deleteCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	route, err := NewRoute(ctx)
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

func Int64QueryParam(r *http.Request, name string) (int64, error) {
	beforeParam := r.URL.Query().Get(name)
	if beforeParam != "" {
		return strconv.ParseInt(beforeParam, 10, 64)
	} else {
		return 0, nil
	}
}