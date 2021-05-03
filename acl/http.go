package acl

import (
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/common"
	pb "github.com/omecodes/store/gen/go/proto"
	"net/http"
)

func HTTPRoutesHandler(middleware ...mux.MiddlewareFunc) http.Handler {
	r := mux.NewRouter()

	r.Name("SaveACLNamespace").Methods(http.MethodPut, http.MethodGet).Handler(http.StripPrefix(common.ApiCreateACLNamespaceConfigRoute, http.HandlerFunc(HTTPSaveNamespaceConfig)))
	r.Name("GetACLNamespace").Methods(http.MethodGet).Handler(http.StripPrefix(common.ApiGetACLNamespaceConfigRoute, http.HandlerFunc(HTTPGetNamespaceConfig)))
	r.Name("DeleteACLNamespace").Methods(http.MethodDelete).Handler(http.StripPrefix(common.ApiDeleteACLNamespaceConfigRoute, http.HandlerFunc(HTTPDeleteNamespaceConfig)))

	r.Name("SaveACLRelationTuple").Methods(http.MethodPut).Path(common.ApiSaveACLRelationTupleRoute).Handler(http.HandlerFunc(HTTPSaveACL))
	r.Name("DeleteACLRelationTuple").Methods(http.MethodDelete).Path(common.ApiDeleteACLRelationTupleRoute).Handler(http.HandlerFunc(HTTPDeleteACL))
	r.Name("CheckACLRelationTuple").Methods(http.MethodPost).Path(common.ApiCheckACLRelationTupleRoute).Handler(http.HandlerFunc(HTTPCheckACL))
	r.Name("GetACLRelationTupleSubjectsSet").Methods(http.MethodPost).Path(common.ApiGetACLRelationTupleSubjectsNamesRoute).Handler(http.HandlerFunc(HTTPGetSubjectsNames))
	r.Name("GetACLRelationTupleObjectsSet").Methods(http.MethodPost).Path(common.ApiGetACLRelationTupleObjectsNamesRoute).Handler(http.HandlerFunc(HTTPGetObjectsNames))

	var handler http.Handler
	handler = r

	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func HTTPSaveNamespaceConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSON)

	nc := new(pb.NamespaceConfig)
	err := jsonpb.Unmarshal(r.Body, nc)
	if err != nil {
		logs.Error("could not decode request body", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(err)
		return
	}

	ctx := r.Context()

	handler := GetHandler(ctx)
	err = handler.SaveNamespaceConfig(ctx, nc, SaveNamespaceConfigOptions{})

	if err != nil {
		logs.Error("failed to save namespace config", logs.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		_ = json.NewEncoder(w).Encode(err)
		return
	}
}

func HTTPGetNamespaceConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSON)

	id := mux.Vars(r)[common.ApiRouteVarIdName]
	ctx := r.Context()

	handler := GetHandler(ctx)
	nc, err := handler.GetNamespaceConfig(ctx, id, GetNamespaceOptions{})
	if err != nil {
		logs.Error("failed to get namespace config", logs.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		_ = json.NewEncoder(w).Encode(err)
		return
	}

	_ = json.NewEncoder(w).Encode(nc)
}

func HTTPDeleteNamespaceConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSON)

	id := mux.Vars(r)[common.ApiRouteVarIdName]
	ctx := r.Context()

	err := GetHandler(ctx).DeleteNamespaceConfig(ctx, id, DeleteNamespaceOptions{})
	if err != nil {
		logs.Error("failed to delete namespace config", logs.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		_ = json.NewEncoder(w).Encode(err)
		return
	}
}

func HTTPSaveACL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSON)

	a := new(pb.ACL)
	err := jsonpb.Unmarshal(r.Body, a)
	if err != nil {
		logs.Error("could not decode request body", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(err)
		return
	}

	ctx := r.Context()

	handler := GetHandler(ctx)
	err = handler.SaveACL(ctx, a, SaveACLOptions{})

	if err != nil {
		logs.Error("failed to save acl", logs.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		_ = json.NewEncoder(w).Encode(err)
		return
	}
}

func HTTPDeleteACL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSON)

	a := new(pb.ACL)
	err := jsonpb.Unmarshal(r.Body, a)
	if err != nil {
		logs.Error("could not decode request body", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(err)
		return
	}

	ctx := r.Context()

	err = GetHandler(ctx).DeleteACL(ctx, a, DeleteACLOptions{})
	if err != nil {
		logs.Error("failed to delete namespace config", logs.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		_ = json.NewEncoder(w).Encode(err)
		return
	}
}

func HTTPCheckACL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSON)

	set := new(pb.SubjectSet)
	err := jsonpb.Unmarshal(r.Body, set)
	if err != nil {
		logs.Error("could not decode request body", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(err)
		return
	}

	ctx := r.Context()

	valid, err := GetHandler(ctx).CheckACL(ctx, "", set, CheckACLOptions{})
	if err != nil {
		logs.Error("failed to delete namespace config", logs.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		_ = json.NewEncoder(w).Encode(err)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"valid": valid,
	})
}

func HTTPGetSubjectsNames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSON)

	set := new(pb.SubjectSet)
	err := jsonpb.Unmarshal(r.Body, set)
	if err != nil {
		logs.Error("could not decode request body", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(err)
		return
	}

	ctx := r.Context()

	names, err := GetHandler(ctx).GetSubjectNames(ctx, set, GetSubjectsNamesOptions{})
	if err != nil {
		logs.Error("failed to get subjects names", logs.Details("set", set), logs.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		_ = json.NewEncoder(w).Encode(err)
		return
	}

	_ = json.NewEncoder(w).Encode(names)
}

func HTTPGetObjectsNames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSON)

	set := new(pb.ObjectSet)
	err := jsonpb.Unmarshal(r.Body, set)
	if err != nil {
		logs.Error("could not decode request body", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(err)
		return
	}

	ctx := r.Context()

	names, err := GetHandler(ctx).GetObjectNames(ctx, set, GetObjectsSetOptions{})
	if err != nil {
		logs.Error("failed to get objects names", logs.Details("set", set), logs.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		_ = json.NewEncoder(w).Encode(err)
		return
	}

	_ = json.NewEncoder(w).Encode(names)
}
