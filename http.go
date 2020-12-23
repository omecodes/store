package oms

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/router"
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

	r.Name("Put").Methods(http.MethodPut).Path("/objects/{id}").Handler(http.HandlerFunc(s.put))
	r.Name("Patch").Methods(http.MethodPatch).Path("/objects/{id}").Handler(http.HandlerFunc(s.patch))
	r.Name("Get").Methods(http.MethodGet).Path("/objects/{id}").Handler(http.HandlerFunc(s.get))
	r.Name("Del").Methods(http.MethodDelete).Path("/objects/{id}").Handler(http.HandlerFunc(s.del))
	r.Name("GetObjects").Methods(http.MethodGet).Path("/objects").Handler(http.HandlerFunc(s.list))
	r.Name("Search").Methods(http.MethodPost).Path("/objects").Handler(http.HandlerFunc(s.search))
	r.PathPrefix("/objects/{id}/").Subrouter().Name("PatchSubDoc").Methods(http.MethodPatch).Handler(http.HandlerFunc(s.patch))
	r.PathPrefix("/objects/{id}/").Subrouter().Name("Select").Methods(http.MethodGet).Handler(http.HandlerFunc(s.sel))

	return r
}

func (s *HTTPUnit) put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var opts oms.PutDataOptions

	object := oms.NewObject()
	object.SetContent(r.Body)
	object.SetSize(r.ContentLength)

	_, err := router.NewRoute(ctx).PutObject(ctx, object, nil, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
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

	err := router.NewRoute(ctx).PatchObject(ctx, patch, oms.PatchOptions{})
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

	object, err := router.NewRoute(ctx).GetObject(ctx, id, oms.GetObjectOptions{Info: onlyInfo == "true"})
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

func (s *HTTPUnit) sel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id := vars["id"]
	filter := strings.Replace(r.RequestURI, fmt.Sprintf("/%s", id), "", 1)

	route := router.NewRoute(ctx)
	object, err := route.GetObject(ctx, id, oms.GetObjectOptions{Path: filter})
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

	err := router.NewRoute(ctx).DeleteObject(ctx, vars["id"])
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

	route := router.NewRoute(ctx)
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

	result, err := router.NewRoute(ctx).SearchObjects(ctx, params, opts)
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

	var o interface{}

	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		log.Error("could not read request body", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = router.NewRoute(ctx).SetSettings(ctx, "", "", oms.SettingsOptions{})
	if err != nil {
		log.Error("failed to set settings", log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
	}
}

func (s *HTTPUnit) getSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.URL.Query().Get("name")

	settings, err := router.NewRoute(ctx).GetSettings(ctx, name)
	if err != nil {
		log.Error("could not get settings", log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write([]byte(settings))
}
