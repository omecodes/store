package oms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/oms"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/router"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
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

	r.Name("Put").Methods(http.MethodPut).Path("/objects").Handler(http.HandlerFunc(s.put))
	r.Name("Patch").Methods(http.MethodPatch).Path("/objects/{id}").Handler(http.HandlerFunc(s.patch))
	r.Name("Get").Methods(http.MethodGet).Path("/objects/{id}").Handler(http.HandlerFunc(s.get))
	r.Name("Del").Methods(http.MethodDelete).Path("/objects/{id}").Handler(http.HandlerFunc(s.del))
	r.Name("GetObjects").Methods(http.MethodGet).Path("/objects").Handler(http.HandlerFunc(s.list))
	r.Name("Search").Methods(http.MethodPost).Path("/objects").Handler(http.HandlerFunc(s.search))

	return r
}

func (s *HTTPUnit) put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var putRequest pb.PutObjectRequest
	err := jsonpb.Unmarshal(r.Body, &putRequest)

	var opts oms.PutDataOptions
	opts.Indexes = putRequest.Indexes

	object := oms.NewObject()
	object.SetContent(bytes.NewBufferString(putRequest.Data))
	object.SetSize(int64(len(putRequest.Data)))

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := route.PutObject(ctx, object, putRequest.AccessSecurityRules, opts)
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

	vars := mux.Vars(r)
	id := vars["id"]
	p := strings.Replace(r.RequestURI, fmt.Sprintf("/%s", id), "", 1)

	patch := oms.NewPatch(id, p)
	patch.SetContent(r.Body)
	patch.SetSize(r.ContentLength)
	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = route.PatchObject(ctx, patch, oms.PatchOptions{})
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func (s *HTTPUnit) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id := vars["id"]
	onlyInfo := r.URL.Query().Get("info")

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	object, err := route.GetObject(ctx, id, oms.GetObjectOptions{Info: onlyInfo == "true"})
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	data, err := ioutil.ReadAll(object.GetContent())
	if err != nil {
		log.Error("Get: failed to encoded object", log.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Error("failed to write response", log.Err(err))
	}
}

func (s *HTTPUnit) del(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = route.DeleteObject(ctx, vars["id"])
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func (s *HTTPUnit) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var before int64
	var err error

	beforeParam := r.URL.Query().Get("before")
	if beforeParam != "" {
		before, err = strconv.ParseInt(beforeParam, 10, 64)
		if err != nil {
			log.Error("could not parse param 'before'")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		before = time.Now().UnixNano() / 1e6
	}

	opts := oms.ListOptions{
		Path:   r.URL.Query().Get("path"),
		Before: before,
	}

	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := route.ListObjects(ctx, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write([]byte(fmt.Sprintf("{\"count\": %d, \"before\": %d, \"objects\": {", result.Count, result.Before)))
	if err != nil {
		log.Error("GetObjects: failed to write response")
		return
	}

	position := 0
	for _, object := range result.Objects {
		var item string
		if position == 0 {
			position++
		} else {
			item = ","
		}

		data, err := ioutil.ReadAll(object.GetContent())
		if err != nil {
			log.Error("GetObjects: failed to encode object", log.Err(err))
			return
		}

		item = item + fmt.Sprintf("\"%s\": %s", object.ID(), string(data))
		_, err = w.Write([]byte(item))
		if err != nil {
			log.Error("GetObjects: failed to write result item", log.Err(err))
			return
		}
	}
	_, err = w.Write([]byte("}}"))
	if err != nil {
		log.Error("GetObjects: failed to write response")
	}
}

func (s *HTTPUnit) search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var before int64
	var err error

	beforeParam := r.URL.Query().Get("before")
	if beforeParam != "" {
		before, err = strconv.ParseInt(beforeParam, 10, 64)
		if err != nil {
			log.Error("could not parse param 'before'")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		before = time.Now().UnixNano() / 1e6
	}

	opts := oms.SearchOptions{
		Path:   r.URL.Query().Get("path"),
		Before: before,
	}

	var params oms.SearchParams
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		log.Error("Search: wrong query", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	route, err := router.NewRoute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := route.SearchObjects(ctx, params, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write([]byte(fmt.Sprintf("{\"count\": %d, \"before\": %d, \"objects\": {", result.Count, result.Before)))
	if err != nil {
		log.Error("Search: failed to write response")
		return
	}
	position := 0
	for _, object := range result.Objects {
		var item string
		if position == 0 {
			position++
		} else {
			item = ","
		}

		data, err := ioutil.ReadAll(object.GetContent())
		if err != nil {
			log.Error("GetObjects: failed to encode object", log.Err(err))
			return
		}

		item = item + fmt.Sprintf("\"%s\": %s", object.ID(), string(data))
		_, err = w.Write([]byte(item))
		if err != nil {
			log.Error("Search: failed to write result item", log.Err(err))
			return
		}
	}
	_, err = w.Write([]byte("}}"))
	if err != nil {
		log.Error("Search: failed to write response")
	}
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
