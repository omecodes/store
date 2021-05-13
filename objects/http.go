package objects

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/common"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"
	"net/http"
	"strings"
)

const (
	queryOffset = "offset"
	queryAt     = "at"
	queryHeader = "header"
)

func MuxRouter(middleware ...mux.MiddlewareFunc) http.Handler {
	r := mux.NewRouter()

	r.Name("CreateCollection").Methods(http.MethodPut).Path(common.ApiCreateCollectionRoute).Handler(http.HandlerFunc(HTTPHandleCreateCollection))
	r.Name("ListCollections").Methods(http.MethodGet).Path(common.ApiListCollectionRoute).Handler(http.HandlerFunc(HTTPHandleListCollections))
	r.Name("DeleteCollection").Methods(http.MethodGet).Path(common.ApiDeleteCollectionRoute).Handler(http.HandlerFunc(HTTPHandleDeleteCollection))
	r.Name("GetCollection").Methods(http.MethodGet).Path(common.ApiGetCollectionRoute).Handler(http.HandlerFunc(HTTPHandleGetCollection))

	r.Name("PutObject").Methods(http.MethodPut).Path(common.ApiPutObjectRoute).Handler(http.HandlerFunc(HTTPHandlePutObject))
	r.Name("PatchObject").Methods(http.MethodPatch).Path(common.ApiPatchObjectRoute).Handler(http.HandlerFunc(HTTPHandlePatchObject))
	r.Name("MoveObject").Methods(http.MethodPost).Path(common.ApiMoveObjectRoute).Handler(http.HandlerFunc(HTTPHandleMoveObject))
	r.Name("GetObject").Methods(http.MethodGet).Path(common.ApiGetObjectRoute).Handler(http.HandlerFunc(HTTPHandleGetObject))
	r.Name("DeleteObject").Methods(http.MethodDelete).Path(common.ApiDeleteObjectRoute).Handler(http.HandlerFunc(HTTPHandleDeleteObject))
	r.Name("ListObjects").Methods(http.MethodGet).Path(common.ApiListObjectsRoute).Handler(http.HandlerFunc(HTTPHandleListObjects))
	r.Name("SearchObjects").Methods(http.MethodPost).Path(common.ApiSearchObjectsRoute).Handler(http.HandlerFunc(HTTPHandleSearchObjects))

	var h http.Handler
	h = r
	for _, m := range middleware {
		h = m(h)
	}
	return h
}

func HTTPHandlePutObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	collection := vars[common.ApiRouteVarCollectionName]

	/*contentType := r.Header.Get(common.HttpHeaderContentType)
	if contentType != common.ContentTypeJSON {
		w.WriteHeader(http.StatusBadRequest)
		return
	}*/

	var putRequest *pb.PutObjectRequest
	err := json.NewDecoder(r.Body).Decode(&putRequest)
	if err != nil {
		logs.Error("failed to decode request body", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if putRequest.Object.Header == nil {
		putRequest.Object.Header = &pb.Header{}
	}
	putRequest.Object.Header.Size = int64(len(putRequest.Object.Data))

	id, err := PutObject(ctx, collection, putRequest.Object, putRequest.ActionAuthorizedUsers, putRequest.Indexes, PutOptions{})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	w.Header().Add(common.HttpHeaderContentType, common.ContentTypeJSON)
	_, _ = w.Write([]byte(fmt.Sprintf("{\"id\": \"%s\"}", id)))
}

func HTTPHandlePatchObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.Header.Get(common.HttpHeaderContentType)
	if contentType != common.ContentTypeJSON {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var patch pb.Patch
	err := jsonpb.Unmarshal(r.Body, &patch)
	if err != nil {
		logs.Error("failed to decode request body", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	collection := vars[common.ApiRouteVarCollectionName]
	patch.ObjectId = vars[common.ApiRouteVarIdName]

	err = PatchObject(ctx, collection, &patch, PatchOptions{})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func HTTPHandleMoveObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.Header.Get(common.HttpHeaderContentType)
	if contentType != common.ContentTypeJSON {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var request pb.MoveObjectRequest
	err := jsonpb.Unmarshal(r.Body, &request)
	if err != nil {
		logs.Error("failed to decode request body", logs.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	collection := vars[common.ApiRouteVarCollectionName]
	objectId := vars[common.ApiRouteVarIdName]

	err = MoveObject(ctx, collection, objectId, request.TargetCollection, request.AccessSecurityRules, MoveOptions{})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func HTTPHandleGetObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	collection := vars[common.ApiRouteVarCollectionName]
	id := vars[common.ApiRouteVarIdName]

	header := r.URL.Query().Get(queryHeader)
	at := r.URL.Query().Get(queryAt)

	object, err := GetObject(ctx, collection, id, GetOptions{
		At:   at,
		Info: header == "true",
	})
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	w.Header().Add(common.HttpHeaderContentType, common.ContentTypeJSON)
	_, err = w.Write([]byte(object.Data))
	if err != nil {
		logs.Error("failed to write response", logs.Err(err))
	}
}

func HTTPHandleDeleteObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	collection := vars[common.ApiRouteVarCollectionName]
	id := vars[common.ApiRouteVarIdName]

	err := DeleteObject(ctx, collection, id)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func HTTPHandleListObjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		err  error
		opts ListOptions
	)

	vars := mux.Vars(r)
	collection := vars[common.ApiRouteVarCollectionName]

	opts.Offset, err = common.Int64QueryParam(r, queryOffset)
	if err != nil {
		logs.Error("could not parse param 'before'")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	opts.At = r.URL.Query().Get(queryAt)

	cursor, err := ListObjects(ctx, collection, opts)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	defer func() {
		if cErr := cursor.Close(); cErr != nil {
			logs.Error("cursor closed with an error", logs.Err(cErr))
		}
	}()

	accept := r.Header.Get(common.HttpHeaderAccept)
	acceptsJsonStream := strings.Contains(accept, common.ContentTypeJSONStream)
	if acceptsJsonStream {
		w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSONStream)
	} else {
		w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSON)
	}

	encoder := json.NewEncoder(w)

	if !acceptsJsonStream {
		_, err = w.Write([]byte("["))
	}

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
		if !acceptsJsonStream {
			if position == 0 {
				position++
			} else {
				item = ","
			}
			_, err = w.Write([]byte(item))
			if err != nil {
				logs.Error("GetObjects: failed to write result objects", logs.Err(err))
				return
			}
		}

		err = encoder.Encode(object)
		if err != nil {
			logs.Error("GetObjects: failed to write result objects", logs.Err(err))
			return
		}
	}

	if !acceptsJsonStream {
		_, err = w.Write([]byte("]"))
	}
}

func HTTPHandleSearchObjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var query pb.SearchQuery
	err := jsonpb.Unmarshal(r.Body, &query)
	if err != nil {
		logs.Error("could not parse search query")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	collection := vars[common.ApiRouteVarCollectionName]

	cursor, err := SearchObjects(ctx, collection, &query)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
	defer func() {
		if cErr := cursor.Close(); cErr != nil {
			logs.Error("cursor closed with an error", logs.Err(cErr))
		}
	}()

	accept := r.Header.Get(common.HttpHeaderAccept)
	acceptsJsonStream := strings.Contains(accept, common.ContentTypeJSONStream)
	if acceptsJsonStream {
		w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSONStream)
	} else {
		w.Header().Set(common.HttpHeaderContentType, common.ContentTypeJSON)
	}

	encoder := json.NewEncoder(w)

	if !acceptsJsonStream {
		_, err = w.Write([]byte("["))
	}

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
		if !acceptsJsonStream {
			if position == 0 {
				position++
			} else {
				item = ","
			}
			_, err = w.Write([]byte(item))
			if err != nil {
				logs.Error("GetObjects: failed to write result objects", logs.Err(err))
				return
			}
		}

		err = encoder.Encode(object)
		if err != nil {
			logs.Error("GetObjects: failed to write result objects", logs.Err(err))
			return
		}
	}

	if !acceptsJsonStream {
		_, err = w.Write([]byte("]"))
	}
}

func HTTPHandleCreateCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var collection *pb.Collection
	err := json.NewDecoder(r.Body).Decode(&collection)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = CreateCollection(ctx, collection)
	if err != nil {
		logs.Error("could not create collection", logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}

func HTTPHandleListCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	collections, err := ListCollections(ctx)
	if err != nil {
		logs.Error("could not load collections", logs.Err(err))
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	w.Header().Add(common.HttpHeaderContentType, common.ContentTypeJSON)
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

func HTTPHandleGetCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars[common.ApiRouteVarIdName]

	collection, err := GetCollection(ctx, id)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	w.Header().Add(common.HttpHeaderContentType, common.ContentTypeJSON)

	data, err := json.Marshal(collection)
	if err != nil {
		logs.Error("could not get collection", logs.Details("col-id", collection), logs.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(data)
}

func HTTPHandleDeleteCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars[common.ApiRouteVarIdName]

	err := DeleteCollection(ctx, id)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}
}
