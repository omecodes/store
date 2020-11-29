package server

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

func dataRouter() *mux.Router {
	r := mux.NewRouter()

	r.Name("RegisterWorker").Methods(http.MethodPost).Path("/workers").Handler(http.HandlerFunc(registerWorker))
	r.Name("ListWorkers").Methods(http.MethodGet).Path("/workers").Handler(http.HandlerFunc(listWorkers))

	r.Name("SetSettings").Methods(http.MethodPut).Path("/settings").Handler(http.HandlerFunc(setSettings))
	r.Name("GetSettings").Methods(http.MethodGet).Path("/settings").Handler(http.HandlerFunc(getSettings))

	settingsSubRouter := r.PathPrefix("/settings/").Subrouter()
	settingsSubRouter.Name("SetSettings").Methods(http.MethodPost).Handler(http.HandlerFunc(setSettings))
	settingsSubRouter.Name("GetSettings").Methods(http.MethodGet).Handler(http.HandlerFunc(getSettings))

	r.Name("Put").Methods(http.MethodPut).Path("/objects/{id}").Handler(http.HandlerFunc(put))
	r.Name("Patch").Methods(http.MethodPatch).Path("/objects/{id}").Handler(http.HandlerFunc(patch))
	r.Name("Get").Methods(http.MethodGet).Path("/objects/{id}").Handler(http.HandlerFunc(get))
	r.Name("Del").Methods(http.MethodDelete).Path("/objects/{id}").Handler(http.HandlerFunc(del))
	r.Name("GetObjects").Methods(http.MethodGet).Path("/objects").Handler(http.HandlerFunc(list))
	r.Name("Search").Methods(http.MethodPost).Path("/objects").Handler(http.HandlerFunc(search))
	r.PathPrefix("/objects/{id}/").Subrouter().Name("PatchSubDoc").Methods(http.MethodPatch).Handler(http.HandlerFunc(patch))
	r.PathPrefix("/objects/{id}/").Subrouter().Name("Select").Methods(http.MethodGet).Handler(http.HandlerFunc(sel))

	return r
}

func put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var opts oms.PutDataOptions

	object := oms.NewObject()
	object.SetContent(r.Body, r.ContentLength)

	_, err := router.Route().PutObject(ctx, object, nil, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func patch(w http.ResponseWriter, r *http.Request) {
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
	patch.SetContent(r.Body, r.ContentLength)

	err := router.Route().PatchObject(ctx, patch, oms.PatchOptions{})
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id := vars["id"]
	onlyInfo := r.URL.Query().Get("info")

	object, err := router.Route().GetObject(ctx, id, oms.GetDataOptions{Info: onlyInfo == "true"})
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	data, err := ioutil.ReadAll(object.Content())
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

func sel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id := vars["id"]
	filter := strings.Replace(r.RequestURI, fmt.Sprintf("/%s", id), "", 1)

	object, err := router.Route().GetObject(ctx, id, oms.GetDataOptions{Path: filter})
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	data, err := ioutil.ReadAll(object.Content())
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

func del(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	err := router.Route().DeleteObject(ctx, vars["id"])
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func list(w http.ResponseWriter, r *http.Request) {
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
		before = time.Now().Unix()
	}

	opts := oms.ListOptions{
		Path:   r.URL.Query().Get("path"),
		Before: before,
	}

	result, err := router.Route().ListObjects(ctx, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write([]byte(fmt.Sprintf("{\"count\": %d, \"offset\": %d, \"objects\": {", result.Count, result.Offset)))
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

		data, err := ioutil.ReadAll(object.Content())
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

func search(w http.ResponseWriter, r *http.Request) {
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
		before = time.Now().Unix()
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

	result, err := router.Route().SearchObjects(ctx, params, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write([]byte(fmt.Sprintf("{\"count\": %d, \"offset\": %d, \"objects\": {", result.Count, result.Offset)))
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

		data, err := ioutil.ReadAll(object.Content())
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

func registerWorker(w http.ResponseWriter, r *http.Request) {}

func listWorkers(w http.ResponseWriter, r *http.Request) {}

func setSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var o interface{}

	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		log.Error("could not read request body", log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = router.Route().SetSettings(ctx, oms.NewJSON(o), oms.SettingsOptions{Path: strings.TrimPrefix(r.RequestURI, "/.settings")})
	if err != nil {
		log.Error("failed to set settings", log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
	}
}

func getSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	opts := oms.SettingsOptions{Path: strings.TrimPrefix(r.RequestURI, "/.settings")}
	s, err := router.Route().GetSettings(ctx, opts)
	if err != nil {
		log.Error("could not get settings", log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	if opts.Path != "" {
		format := oms.SettingsPathFormats[opts.Path]
		if format != "" {
			mime := oms.SettingsPathValueMimes[opts.Path]
			w.Header().Add("Content-Type", mime)
			_, _ = w.Write([]byte(fmt.Sprintf(format, s.GetObject())))
			return
		}
	}

	data, err := s.Marshal()
	if err != nil {
		log.Error("could not encode settings result", log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(data)
}
