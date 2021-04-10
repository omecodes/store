package files

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/common"
	"io"
	"net/http"
	"strings"
)

func MuxRouter(middleware ...mux.MiddlewareFunc) http.Handler {
	r := mux.NewRouter()

	treeRoute := r.PathPrefix("/tree/").Subrouter()
	treeRoute.Name(common.ApiFileTreeRoutePrefix).Methods(http.MethodPut).Handler(http.StripPrefix(common.ApiFileTreeRoutePrefix, http.HandlerFunc(HTTPHandleCreateFile)))
	treeRoute.Name("ListDir").Methods(http.MethodPost).Handler(http.StripPrefix(common.ApiFileTreeRoutePrefix, http.HandlerFunc(HTTPHandleListDir)))
	treeRoute.Name("GetFileInfo").Methods(http.MethodGet).Handler(http.StripPrefix(common.ApiFileTreeRoutePrefix, http.HandlerFunc(HTTPHandleGetFileInfo)))
	treeRoute.Name("DeleteFile").Methods(http.MethodDelete).Handler(http.StripPrefix(common.ApiFileTreeRoutePrefix, http.HandlerFunc(HTTPHandleDeleteFile)))
	treeRoute.Name("PatchTree").Methods(http.MethodPatch).Handler(http.StripPrefix(common.ApiFileTreeRoutePrefix, http.HandlerFunc(HTTPHandlePatchFileTree)))

	attrRoute := r.PathPrefix(common.ApiFileAttributesRoutePrefix).Subrouter()
	attrRoute.Name("GetFileAttributes").Methods(http.MethodGet).Handler(http.StripPrefix(common.ApiFileAttributesRoutePrefix, http.HandlerFunc(HTTPHandleGetFileAttributes)))
	attrRoute.Name("SetFileAttributes").Methods(http.MethodPost).Handler(http.StripPrefix(common.ApiFileTreeRoutePrefix, http.HandlerFunc(HTTPHandleSetFileAttributes)))

	dataRoute := r.PathPrefix(common.ApiFileDataRoutePrefix).Subrouter()
	dataRoute.Name("Download").Methods(http.MethodGet).Handler(http.StripPrefix(common.ApiFileDataRoutePrefix, http.HandlerFunc(HTTPHandleDownloadFile)))
	dataRoute.Name("Upload").Methods(http.MethodPut, http.MethodPost).Handler(http.StripPrefix(common.ApiFileDataRoutePrefix, http.HandlerFunc(HTTPHandleUploadFile)))

	r.Name("CreateSource").Path(common.ApiCreateFileSource).Methods(http.MethodPut).HandlerFunc(HTTPHandleCreateSource)
	r.Name("ListSources").Path(common.ApiListFileSources).Methods(http.MethodGet).HandlerFunc(HTTPHandleListSources)
	r.Name("GetSource").Path(common.ApiGetFileSource).Methods(http.MethodGet).HandlerFunc(HTTPHandleGetSource)
	r.Name("DeleteSource").Path(common.ApiDeleteFileSource).Methods(http.MethodDelete).HandlerFunc(HTTPHandleDeleteSource)

	var handler http.Handler
	handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func HTTPHandleCreateFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	if r.ContentLength > 0 {
		var location *FileLocation
		err := json.NewDecoder(r.Body).Decode(&location)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		handler := GetRouteHandler(ctx)
		err = handler.CopyFile(ctx, sourceID, filename, strings.TrimPrefix(location.Filename, sourceID))
		if err != nil {
			w.WriteHeader(errors.HttpStatus(err))
			return
		}
	}

	err := CreateDir(ctx, sourceID, filename)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func HTTPHandleListDir(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var opts ListDirOptions
	err := json.NewDecoder(r.Body).Decode(&opts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	content, err := ListDir(ctx, sourceID, filename, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	_ = json.NewEncoder(w).Encode(content)
}

func HTTPHandleGetFileInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var opts GetFileOptions
	opts.WithAttrs = r.URL.Query().Get("attrs") == "true"

	info, err := GetFile(ctx, sourceID, filename, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	_ = json.NewEncoder(w).Encode(info)
}

func HTTPHandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var opts DeleteFileOptions
	opts.Recursive = r.URL.Query().Get("recursive") == "true"

	err := DeleteFile(ctx, sourceID, filename, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func HTTPHandlePatchFileTree(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var patchInfo *TreePatchInfo
	err := json.NewDecoder(r.Body).Decode(&patchInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if patchInfo.Rename {
		err = RenameFile(ctx, sourceID, filename, patchInfo.Value)
	} else {
		err = MoveFile(ctx, sourceID, filename, strings.TrimPrefix(patchInfo.Value, sourceID))
	}
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func HTTPHandleSetFileAttributes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var attributes Attributes
	err := json.NewDecoder(r.Body).Decode(&attributes)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler := GetRouteHandler(ctx)
	err = handler.SetFileAttributes(ctx, sourceID, filename, attributes)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func HTTPHandleGetFileAttributes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	name := r.URL.Query().Get("name")

	handler := GetRouteHandler(ctx)
	attributes, err := handler.GetFileAttributes(ctx, sourceID, filename, name)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	_ = json.NewEncoder(w).Encode(attributes)
}

func HTTPHandleUploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	opts := WriteOptions{}
	opts.Append = r.URL.Query().Get("append") == "true"
	opts.Hash = r.Header.Get("X-Content-Hash")

	err := WriteFileContent(ctx, sourceID, filename, r.Body, r.ContentLength, opts)
	if err != nil {
		logs.Error("could not put file content", logs.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func HTTPHandleDownloadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var err error

	opts := ReadOptions{}
	opts.Range.Offset, err = common.Int64QueryParam(r, "offset")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	opts.Range.Length, err = common.Int64QueryParam(r, "length")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, size, err := ReadFileContent(ctx, sourceID, filename, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}

	defer func() {
		if cer := file.Close(); cer != nil {
			logs.Error("file stream close", logs.Err(err))
		}
	}()

	if size == 0 {
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	buf := make([]byte, 1024)
	done := false
	for !done {
		n, err := file.Read(buf)
		if err != nil {
			if done = err == io.EOF; !done {
				logs.Error("failed to read file content", logs.Err(err))
				return
			}
		}

		_, err = w.Write(buf[:n])
		if err != nil {
			logs.Error("failed to write file content", logs.Err(err))
			return
		}
	}
}

func HTTPHandleCreateSource(w http.ResponseWriter, r *http.Request) {
	var source *Source
	err := json.NewDecoder(r.Body).Decode(&source)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	err = CreateSource(ctx, source)
	if err != nil {
		logs.Error("could not create source", logs.Err(err))
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func HTTPHandleListSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sources, err := ListSources(ctx)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	_ = json.NewEncoder(w).Encode(sources)
}

func HTTPHandleGetSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := vars[common.ApiRouteVarId]

	ctx := r.Context()

	source, err := GetSource(ctx, sourceID)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	_ = json.NewEncoder(w).Encode(source)
}

func HTTPHandleDeleteSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := vars[common.ApiRouteVarId]

	ctx := r.Context()

	err := DeleteSource(ctx, sourceID)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

type middlewareRouteOptions struct {
	sourceManager  SourceManager
	fsProvider     FSProvider
	routerProvider RouterProvider
}

type MiddlewareOption func(options *middlewareRouteOptions)

func Middleware(opts ...MiddlewareOption) mux.MiddlewareFunc {
	var options middlewareRouteOptions
	for _, opt := range opts {
		opt(&options)
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if options.fsProvider != nil {
				ctx = context.WithValue(ctx, ctxFsProvider{}, options.fsProvider)
			}

			if options.sourceManager != nil {
				ctx = context.WithValue(ctx, ctxSourceManager{}, options.sourceManager)
			}

			if options.routerProvider != nil {
				ctx = context.WithValue(ctx, ctxRouterProvider{}, options.routerProvider)
			}

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}

}

func MiddlewareWithSourceManager(manager SourceManager) MiddlewareOption {
	return func(options *middlewareRouteOptions) {
		options.sourceManager = manager
	}
}

func MiddlewareWithFsProvider(provider FSProvider) MiddlewareOption {
	return func(options *middlewareRouteOptions) {
		options.fsProvider = provider
	}
}

func MiddlewareWithRouterProvider(provider RouterProvider) MiddlewareOption {
	return func(options *middlewareRouteOptions) {
		options.routerProvider = provider
	}
}
