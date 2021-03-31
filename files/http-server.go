package files

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/libome/logs"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	pathVarId = "id"
)

func MuxRouter(middleware ...mux.MiddlewareFunc) http.Handler {
	r := mux.NewRouter()

	treeRoute := r.PathPrefix("/tree/").Subrouter()
	treeRoute.Name("CreateFile").Methods(http.MethodPut).Handler(http.StripPrefix("/tree/", http.HandlerFunc(APIHandleCreateFile)))
	treeRoute.Name("ListDir").Methods(http.MethodPost).Handler(http.StripPrefix("/tree/", http.HandlerFunc(APIHandleListDir)))
	treeRoute.Name("GetFileInfo").Methods(http.MethodGet).Handler(http.StripPrefix("/tree/", http.HandlerFunc(APIHandleGetFileInfo)))
	treeRoute.Name("DeleteFile").Methods(http.MethodDelete).Handler(http.StripPrefix("/tree/", http.HandlerFunc(APIHandleDeleteFile)))
	treeRoute.Name("PatchTree").Methods(http.MethodPatch).Handler(http.StripPrefix("/tree/", http.HandlerFunc(APIHandlePatchFileTree)))

	attrRoute := r.PathPrefix("/attr/").Subrouter()
	attrRoute.Name("GetFileAttributes").Methods(http.MethodGet).Handler(http.StripPrefix("/attr/", http.HandlerFunc(APIHandleGetFileAttributes)))
	attrRoute.Name("SetFileAttributes").Methods(http.MethodPut).Handler(http.StripPrefix("/attr/", http.HandlerFunc(APIHandleSetFileAttributes)))

	dataRoute := r.PathPrefix("/data/").Subrouter()
	dataRoute.Name("Download").Methods(http.MethodGet).Handler(http.StripPrefix("/data/", http.HandlerFunc(APIHandleDownloadFile)))
	dataRoute.Name("Upload").Methods(http.MethodPut, http.MethodPost).Handler(http.StripPrefix("/data/", http.HandlerFunc(APIHandleUploadFile)))

	r.Name("CreateSource").Path("/sources").Methods(http.MethodPut).HandlerFunc(APIHandleCreateSource)
	r.Name("ListSources").Path("/sources").Methods(http.MethodGet).HandlerFunc(APIHandleListSources)
	r.Name("GetSource").Path("/sources/{id}").Methods(http.MethodGet).HandlerFunc(APIHandleGetSource)
	r.Name("DeleterSource").Path("/sources/{id}").Methods(http.MethodDelete).HandlerFunc(APIHandleDeleteSource)

	var handler http.Handler
	handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func APIHandleCreateFile(w http.ResponseWriter, r *http.Request) {
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

func APIHandleListDir(w http.ResponseWriter, r *http.Request) {
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

func APIHandleGetFileInfo(w http.ResponseWriter, r *http.Request) {
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

func APIHandleDeleteFile(w http.ResponseWriter, r *http.Request) {
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

func APIHandlePatchFileTree(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var patchInfo *TreePatchInfo
	err := json.NewDecoder(r.Body).Decode(&patchInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if patchInfo.Rename {
		err = RenameFile(ctx, sourceID, filename, strings.TrimPrefix(patchInfo.Value, sourceID))
	} else {
		err = MoveFile(ctx, sourceID, filename, strings.TrimPrefix(patchInfo.Value, sourceID))
	}
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func APIHandleSetFileAttributes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var attributes Attributes
	err := json.NewDecoder(r.Body).Decode(&attributes)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler := GetRouteHandler(ctx)
	err = handler.SetFileMetaData(ctx, sourceID, filename, attributes)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func APIHandleGetFileAttributes(w http.ResponseWriter, r *http.Request) {
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

func APIHandleUploadFile(w http.ResponseWriter, r *http.Request) {
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

func APIHandleDownloadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var err error

	opts := ReadOptions{}
	opts.Range.Offset, err = Int64QueryParam(r, "offset")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	opts.Range.Length, err = Int64QueryParam(r, "length")
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

func APIHandleCreateSource(w http.ResponseWriter, r *http.Request) {
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

func APIHandleListSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sources, err := ListSources(ctx)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	_ = json.NewEncoder(w).Encode(sources)
}

func APIHandleGetSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := vars[pathVarId]

	ctx := r.Context()

	source, err := GetSource(ctx, sourceID)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	_ = json.NewEncoder(w).Encode(source)
}

func APIHandleDeleteSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := vars[pathVarId]

	ctx := r.Context()

	err := DeleteSource(ctx, sourceID)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func Int64QueryParam(r *http.Request, name string) (int64, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return strconv.ParseInt(param, 10, 64)
	} else {
		return 0, nil
	}
}
