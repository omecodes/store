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
	treeRoute.Name("CreateFile").Methods(http.MethodPut).Handler(http.StripPrefix("/tree/", http.HandlerFunc(CreateFile)))
	treeRoute.Name("ListDir").Methods(http.MethodPost).Handler(http.StripPrefix("/tree/", http.HandlerFunc(ListDir)))
	treeRoute.Name("GetFileInfo").Methods(http.MethodGet).Handler(http.StripPrefix("/tree/", http.HandlerFunc(GetFileInfo)))
	treeRoute.Name("DeleteFile").Methods(http.MethodDelete).Handler(http.StripPrefix("/tree/", http.HandlerFunc(DeleteFile)))
	treeRoute.Name("PatchTree").Methods(http.MethodPatch).Handler(http.StripPrefix("/tree/", http.HandlerFunc(PatchFileTree)))

	attrRoute := r.PathPrefix("/attr/").Subrouter()
	attrRoute.Name("GetFileAttributes").Methods(http.MethodGet).Handler(http.StripPrefix("/attr/", http.HandlerFunc(GetFileAttributes)))
	attrRoute.Name("SetFileAttributes").Methods(http.MethodPut).Handler(http.StripPrefix("/attr/", http.HandlerFunc(SetFileAttributes)))

	dataRoute := r.PathPrefix("/data/").Subrouter()
	dataRoute.Name("Download").Methods(http.MethodGet).Handler(http.StripPrefix("/data/", http.HandlerFunc(DownloadFile)))
	dataRoute.Name("Upload").Methods(http.MethodPut, http.MethodPost).Handler(http.StripPrefix("/data/", http.HandlerFunc(UploadFile)))

	r.Name("CreateSource").Path("/sources").Methods(http.MethodPut).HandlerFunc(CreateSource)
	r.Name("ListSources").Path("/sources").Methods(http.MethodGet).HandlerFunc(ListSources)
	r.Name("GetSource").Path("/sources/{id}").Methods(http.MethodGet).HandlerFunc(GetSource)
	r.Name("DeleterSource").Path("/sources/{id}").Methods(http.MethodDelete).HandlerFunc(DeleteSource)

	var handler http.Handler
	handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func CreateFile(w http.ResponseWriter, r *http.Request) {
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

	handler := GetRouteHandler(ctx)
	err := handler.CreateDir(ctx, sourceID, filename)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func ListDir(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var opts ListDirOptions
	err := json.NewDecoder(r.Body).Decode(&opts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler := GetRouteHandler(ctx)
	content, err := handler.ListDir(ctx, sourceID, filename, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	_ = json.NewEncoder(w).Encode(content)
}

func GetFileInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var opts GetFileInfoOptions
	opts.WithAttrs = r.URL.Query().Get("attrs") == "true"

	handler := GetRouteHandler(ctx)
	info, err := handler.GetFileInfo(ctx, sourceID, filename, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	_ = json.NewEncoder(w).Encode(info)
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var opts DeleteFileOptions
	opts.Recursive = r.URL.Query().Get("recursive") == "true"

	handler := GetRouteHandler(ctx)
	err := handler.DeleteFile(ctx, sourceID, filename, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func PatchFileTree(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	var patchInfo *TreePatchInfo
	err := json.NewDecoder(r.Body).Decode(&patchInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler := GetRouteHandler(ctx)
	if patchInfo.Rename {
		err = handler.RenameFile(ctx, sourceID, filename, strings.TrimPrefix(patchInfo.Value, sourceID))
	} else {
		err = handler.MoveFile(ctx, sourceID, filename, strings.TrimPrefix(patchInfo.Value, sourceID))
	}
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func SetFileAttributes(w http.ResponseWriter, r *http.Request) {
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

func GetFileAttributes(w http.ResponseWriter, r *http.Request) {
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

func UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceID, filename := Split(r.URL.Path)

	handler := GetRouteHandler(ctx)
	opts := WriteOptions{}
	opts.Append = r.URL.Query().Get("append") == "true"
	opts.Hash = r.Header.Get("X-Content-Hash")

	err := handler.WriteFileContent(ctx, sourceID, filename, r.Body, r.ContentLength, opts)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
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

	handler := GetRouteHandler(ctx)
	file, size, err := handler.ReadFileContent(ctx, sourceID, filename, opts)
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

func CreateSource(w http.ResponseWriter, r *http.Request) {
	var source *Source
	err := json.NewDecoder(r.Body).Decode(&source)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	handler := GetRouteHandler(ctx)
	err = handler.CreateSource(ctx, source)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
}

func ListSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handler := GetRouteHandler(ctx)
	sources, err := handler.ListSources(ctx)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	_ = json.NewEncoder(w).Encode(sources)
}

func GetSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := vars[pathVarId]

	ctx := r.Context()

	handler := GetRouteHandler(ctx)
	source, err := handler.GetSource(ctx, sourceID)
	if err != nil {
		w.WriteHeader(errors.HttpStatus(err))
		return
	}
	_ = json.NewEncoder(w).Encode(source)
}

func DeleteSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := vars[pathVarId]

	ctx := r.Context()

	handler := GetRouteHandler(ctx)
	err := handler.DeleteSource(ctx, sourceID)
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
