package store

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/omestore/ent"
	"github.com/omecodes/omestore/pb"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func dataRouter() *mux.Router {
	r := mux.NewRouter()

	r.Name("RegisterUser").Methods(http.MethodPost).Path("/.users").Handler(http.HandlerFunc(registerUser))
	r.Name("AddUser").Methods(http.MethodPut).Path("/.users").Handler(http.HandlerFunc(createUser))
	r.Name("ListUsers").Methods(http.MethodGet).Path("/.users").Handler(http.HandlerFunc(listUsers))
	r.Name("GetUser").Methods(http.MethodGet).Path("/.users/{name}").Handler(http.HandlerFunc(getUser))
	r.Name("GetCollections").Methods(http.MethodGet).Path("/.collections").Handler(http.HandlerFunc(getCollections))

	r.Name("PutGraft").Methods(http.MethodPut).Path("/.grafts/{collection}/{data}").Handler(http.HandlerFunc(putGraft))
	r.Name("GetGraft").Methods(http.MethodGet).Path("/.grafts/{collection}/{data}/{id}").Handler(http.HandlerFunc(getGraft))
	r.Name("GetAllGrafts").Methods(http.MethodGet).Path("/.grafts/{collection}/{data}").Handler(http.HandlerFunc(getAllGrafts))
	r.Name("DelGraft").Methods(http.MethodDelete).Path("/.grafts/{collection}/{data}/{id}").Handler(http.HandlerFunc(delGraft))

	r.Name("SetSettings").Methods(http.MethodPut).Path("/.settings").Handler(http.HandlerFunc(setSettings))
	r.Name("GetSettings").Methods(http.MethodGet).Path("/.settings").Handler(http.HandlerFunc(getSettings))
	settingsSubRouter := r.PathPrefix("/.settings/").Subrouter()
	settingsSubRouter.Name("SetSettings").Methods(http.MethodPost).Handler(http.HandlerFunc(setSettings))
	settingsSubRouter.Name("GetSettings").Methods(http.MethodGet).Handler(http.HandlerFunc(getSettings))

	dr := mux.NewRouter()
	dr.Name("Put").Methods(http.MethodPut).Path("/{collection}/{id}").Handler(http.HandlerFunc(put))
	dr.Name("Patch").Methods(http.MethodPatch).Path("/{collection}/{id}").Handler(http.HandlerFunc(patch))
	dr.Name("Get").Methods(http.MethodGet).Path("/{collection}/{id}").Handler(http.HandlerFunc(get))
	dr.Name("Del").Methods(http.MethodDelete).Path("/{collection}/{id}").Handler(http.HandlerFunc(del))
	dr.Name("List").Methods(http.MethodGet).Path("/{collection}").Handler(http.HandlerFunc(list))
	dr.PathPrefix("/{collection}/{id}/").Subrouter().Name("PatchSubDoc").Methods(http.MethodPatch).Handler(http.HandlerFunc(patch))
	dr.PathPrefix("/{collection}/{id}/").Subrouter().Name("Select").Methods(http.MethodGet).Handler(http.HandlerFunc(sel))

	r.NotFoundHandler = dr
	return r
}

// Data handler
func put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var opts pb.PutDataOptions

	vars := mux.Vars(r)
	data := &pb.Data{
		ID:         vars["id"],
		Collection: vars["collection"],
		Size:       r.ContentLength,
	}

	contentBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data.Content = string(contentBytes)

	h := getRoute()
	err = h.PutData(ctx, data, opts)
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
	collection := vars["collection"]
	id := vars["id"]
	p := strings.Replace(r.RequestURI, fmt.Sprintf("/%s/%s", collection, id), "", 1)

	h := getRoute()
	err := h.PatchData(ctx, vars["collection"], vars["id"], r.Body, r.ContentLength, pb.PatchOptions{Path: p})
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	collection := vars["collection"]
	id := vars["id"]
	onlyInfo := r.URL.Query().Get("info")

	var (
		data *pb.Data
		err  error
	)

	h := getRoute()
	if onlyInfo == "true" {
		info, err := h.Info(ctx, collection, id)
		if err != nil {
			w.WriteHeader(errors.HttpStatus(err))
			return
		}

		infoBytes, err := json.Marshal(info)
		if err != nil {
			//log.Error("could not json encode response", //log.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(infoBytes)
	} else {
		data, err = h.GetData(ctx, collection, id, pb.GetDataOptions{})
		if err != nil {
			w.WriteHeader(errors.HttpStatus(err))
			return
		}
		infoBytes, err := json.Marshal(data)
		if err != nil {
			//log.Error("could not json encode response", //log.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(infoBytes)
	}
}

func sel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	collection := vars["collection"]
	id := vars["id"]
	filter := strings.Replace(r.RequestURI, fmt.Sprintf("/%s/%s", collection, id), "", 1)

	h := getRoute()
	data, err := h.GetData(ctx, collection, id, pb.GetDataOptions{Path: filter})
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	infoBytes, err := json.Marshal(data)
	if err != nil {
		//log.Error("could not json encode response", //log.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(infoBytes)
}

func del(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	h := getRoute()
	err := h.Delete(ctx, vars["collection"], vars["id"])
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	var before int64
	var err error

	beforeParam := r.URL.Query().Get("before")
	if beforeParam != "" {
		before, err = strconv.ParseInt(beforeParam, 10, 64)
		if err != nil {
			//log.Error("could not parse param 'before'")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		before = time.Now().Unix()
	}

	route := getRoute()
	result, err := route.List(ctx, vars["collection"], pb.ListOptions{
		Path:   r.URL.Query().Get("path"),
		Before: before,
	})
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	if result == nil {
		_, _ = w.Write([]byte("{}"))
		return
	}

	defer result.Cursor.Close()
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write([]byte("{"))
	if err != nil {
		//log.Error("failed to write response")
		return
	}

	position := 0
	for {
		ok, err := result.Cursor.Walk()
		if err != nil {
			//log.Error("could not get next item from cursor", //log.Err(err))
			break
		}

		if !ok {
			break
		}

		data := result.Cursor.Get()

		var item string
		if position == 0 {
			position++
		} else {
			item = ","
		}

		item = item + fmt.Sprintf("\"%s\":{\"created_by\": \"%s\", \"created_at\": %d, \"data\": %s}", data.ID, data.CreatedBy, data.CreatedAt, data.Content)
		_, err = w.Write([]byte(item))
		if err != nil {
			//log.Error("failed to write result item", //log.Err(err))
			return
		}

	}

	_, err = w.Write([]byte("}"))
	if err != nil {
		//log.Error("failed to write response")
	}
}

func putGraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	graft := &pb.Graft{
		ID:         vars["id"],
		DataID:     vars["data"],
		Collection: vars["collection"],
		Size:       r.ContentLength,
	}

	contentBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	graft.Content = string(contentBytes)

	route := getRoute()
	id, err := route.SaveGraft(ctx, graft)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write([]byte(fmt.Sprintf("{\"id\": \"%s\"}", id)))
}

func getGraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	route := getRoute()
	graft, err := route.GetGraft(ctx, vars["collection"], vars["data"], vars["id"])
	if err != nil {
		//log.Error("could not get data graft", log.Field("collection", vars["collection"]), log.Field("data", vars["data"]), //log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	graft.Collection = vars["collection"]

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("{\"%s\":{\"created_by\": \"%s\", \"created_at\": %d, \"data\": %s}}", graft.ID, graft.CreatedBy, graft.CreatedAt, graft.Content)))
}

func getAllGrafts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	route := getRoute()
	result, err := route.ListGrafts(ctx, vars["collection"], vars["data"], pb.ListOptions{})
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	if result == nil {
		_, _ = w.Write([]byte("{}"))
		return
	}

	defer result.Cursor.Close()
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write([]byte("{"))
	if err != nil {
		//log.Error("failed to write response")
		return
	}

	position := 0
	for {
		ok, err := result.Cursor.Walk()
		if err != nil {
			//log.Error("could not get next item from cursor", //log.Err(err))
			break
		}

		if !ok {
			break
		}

		graft := result.Cursor.Get()

		var item string
		if position == 0 {
			position++
		} else {
			item = ","
		}

		item = item + fmt.Sprintf("\"%s\":{\"created_by\": \"%s\", \"created_at\": %d, \"data\": %s}", graft.ID, graft.CreatedBy, graft.CreatedAt, graft.Content)
		_, err = w.Write([]byte(item))
		if err != nil {
			//log.Error("failed to write result item", //log.Err(err))
			return
		}
	}

	_, err = w.Write([]byte("}"))
	if err != nil {
		//log.Error("failed to write response")
	}
}

func delGraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	h := getRoute()
	err := h.DeleteGraft(ctx, vars["collection"], vars["data"], vars["id"])
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

// app settings handler
func createUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var user ent.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		//log.Error("could not decode request body", //log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	route := getRoute()

	existing, err := route.UserInfo(ctx, user.ID, pb.UserOptions{})
	if err != nil && !errors.IsNotFound(err) && !ent.IsNotFound(err) {
		//log.Error("could not verify existing user with the same name", log.Field("name", user.ID), //log.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if existing != nil {
		//log.Error("attempt to create a user with already taken name")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = route.CreateUser(ctx, &user)
	if err != nil {
		//log.Error("could not create user", //log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	route := getRoute()
	info, err := route.UserInfo(ctx, vars["name"], pb.UserOptions{})
	if err != nil {
		//log.Error("could not get user info", //log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	data, err := json.Marshal(info)
	if err != nil {
		//log.Error("could not encode response object", //log.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(data)
}

func getCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	route := getRoute()
	collections, err := route.GetCollections(ctx)
	if err != nil {
		//log.Error("user registration failed", //log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	if collections == nil {
		_, _ = w.Write([]byte("[]"))
		return
	}

	encoded, err := json.Marshal(collections)
	if err != nil {
		//log.Error("could not encode response data", //log.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Content-Size", fmt.Sprintf("%d", len(encoded)))
	_, _ = w.Write(encoded)
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user ent.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		//log.Error("could not decode request body", //log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	route := getRoute()

	existing, err := route.UserInfo(ctx, user.ID, pb.UserOptions{})
	if err != nil && !errors.IsNotFound(err) && !ent.IsNotFound(err) {
		//log.Error("could not verify existing user with the same name", log.Field("name", user.ID), //log.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if existing != nil {
		//log.Error("attempt to create a user with already taken name")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.Validated = false
	err = route.RegisterUser(ctx, &user, pb.UserOptions{})
	if err != nil {
		//log.Error("user registration failed", //log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query()

	route := getRoute()
	users, err := route.ListUsers(ctx, pb.UserOptions{
		WithGroups: "true" == query.Get("with_group"),
	})
	if err != nil {
		//log.Error("user registration failed", //log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	if users == nil {
		_, _ = w.Write([]byte("{}"))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write([]byte("{"))
	if err != nil {
		//log.Error("failed to write response")
		return
	}

	position := 0
	for _, user := range users {

		var item string
		if position == 0 {
			position++
		} else {
			item = ","
		}

		encoded, err := json.Marshal(user)
		if err != nil {
			//log.Error("could not encode response lst item", log.Field(user.ID, user), //log.Err(err))
			return
		}

		item = item + fmt.Sprintf("\"%s\":%s", user.ID, string(encoded))
		_, err = w.Write([]byte(item))
		if err != nil {
			//log.Error("failed to write result item", //log.Err(err))
			return
		}
	}

	_, err = w.Write([]byte("}"))
	if err != nil {
		//log.Error("failed to write response")
	}
}

func setSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var o interface{}

	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		//log.Error("could not read request body", //log.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	route := getRoute()
	err = route.SetSettings(ctx, &JSON{Object: o}, pb.SettingsOptions{Path: strings.TrimPrefix(r.RequestURI, "/.settings")})
	if err != nil {
		//log.Error("failed to set settings", //log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
	}
}

func getSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	route := getRoute()
	opts := pb.SettingsOptions{Path: strings.TrimPrefix(r.RequestURI, "/.settings")}
	s, err := route.GetSettings(ctx, opts)
	if err != nil {
		//log.Error("could not get settings", //log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	if opts.Path != "" {
		format := settingsPathFormats[opts.Path]
		if format != "" {
			mime := settingsPathValueMimes[opts.Path]
			w.Header().Add("Content-Type", mime)
			_, _ = w.Write([]byte(fmt.Sprintf(format, s.Object)))
			return
		}
	}

	data, err := json.Marshal(s.Object)
	if err != nil {
		//log.Error("could not encode settings result", //log.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(data)
}
