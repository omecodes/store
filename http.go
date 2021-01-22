package oms

import (
	"github.com/gorilla/mux"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/objects"
	"net/http"
)

type routeOptions struct {
	filesDir  string
	objects   bool
	appsDir   string
	staticDir string
}

type RouteOption func(options *routeOptions)

func WithFiles(rootDir string) RouteOption {
	return func(options *routeOptions) {
		options.filesDir = rootDir
	}
}

func WithWebApp(appsDir string) RouteOption {
	return func(options *routeOptions) {
		options.appsDir = appsDir
	}
}

func WithObjects() RouteOption {
	return func(options *routeOptions) {
		options.objects = true
	}
}

func WithStaticFiles(staticDir string) RouteOption {
	return func(options *routeOptions) {
		options.staticDir = staticDir
	}
}

func httpRouter(opts ...RouteOption) *mux.Router {
	var options routeOptions
	for _, opt := range opts {
		opt(&options)
	}

	r := mux.NewRouter()

	if options.objects {
		r.Name("SetSettings").Methods(http.MethodPut).Path("/objects/settings").Handler(http.HandlerFunc(objects.SetSettings))
		r.Name("GetSettings").Methods(http.MethodGet).Path("/objects/settings").Handler(http.HandlerFunc(objects.GetSettings))

		r.Name("CreateCollection").Methods(http.MethodPut).Path("/objects/collections").Handler(http.HandlerFunc(objects.CreateCollection))
		r.Name("ListCollections").Methods(http.MethodGet).Path("/objects/collections").Handler(http.HandlerFunc(objects.ListCollections))
		r.Name("DeleteCollection").Methods(http.MethodGet).Path("/objects/collections/{id}").Handler(http.HandlerFunc(objects.DeleteCollection))
		r.Name("GetCollection").Methods(http.MethodGet).Path("/objects/collections/{id}").Handler(http.HandlerFunc(objects.GetCollection))

		r.Name("PutObject").Methods(http.MethodPut).Path("/objects/data/{collection}").Handler(http.HandlerFunc(objects.PutObject))
		r.Name("PatchObject").Methods(http.MethodPatch).Path("/objects/data/{collection}/{id}").Handler(http.HandlerFunc(objects.PatchObject))
		r.Name("MoveObject").Methods(http.MethodPost).Path("/objects/data/{collection}/{id}").Handler(http.HandlerFunc(objects.MoveObject))
		r.Name("GetObject").Methods(http.MethodGet).Path("/objects/data/{collection}/{id}").Handler(http.HandlerFunc(objects.GetObject))
		r.Name("DeleteObject").Methods(http.MethodDelete).Path("/objects/data/{collection}/{id}").Handler(http.HandlerFunc(objects.DeleteObject))
		r.Name("GetObjects").Methods(http.MethodGet).Path("/objects/data/{collection}").Handler(http.HandlerFunc(objects.ListObjects))
		r.Name("SearchObjects").Methods(http.MethodPost).Path("/objects/data/{collection}").Handler(http.HandlerFunc(objects.SearchObjects))
	}

	if options.appsDir != "" {
	}

	if options.filesDir != "" {
		treeRoute := r.PathPrefix("/files/tree/").Subrouter()
		treeRoute.Name("CreateFile").Methods(http.MethodPut).Handler(http.StripPrefix("/files/tree/", http.HandlerFunc(files.CreateFile)))
		treeRoute.Name("ListDir").Methods(http.MethodPost).Handler(http.StripPrefix("/files/tree/", http.HandlerFunc(files.ListDir)))
		treeRoute.Name("GetFileInfo").Methods(http.MethodGet).Handler(http.StripPrefix("/files/tree/", http.HandlerFunc(files.GetFileInfo)))
		treeRoute.Name("DeleteFile").Methods(http.MethodDelete).Handler(http.StripPrefix("/files/tree/", http.HandlerFunc(files.DeleteFile)))
		treeRoute.Name("PatchTree").Methods(http.MethodPatch).Handler(http.StripPrefix("/files/tree/", http.HandlerFunc(files.PatchFileTree)))

		attrRoute := r.PathPrefix("/files/attr/").Subrouter()
		attrRoute.Name("GetFileAttributes").Methods(http.MethodGet).Handler(http.StripPrefix("/files/attr/", http.HandlerFunc(files.GetFileAttributes)))
		attrRoute.Name("SetFileAttributes").Methods(http.MethodPut).Handler(http.StripPrefix("/files/attr/", http.HandlerFunc(files.SetFileAttributes)))

		dataRoute := r.PathPrefix("/files/data/").Subrouter()
		dataRoute.Name("Download").Methods(http.MethodGet).Handler(http.StripPrefix("/files/data/", http.HandlerFunc(files.DownloadFile)))
		dataRoute.Name("Upload").Methods(http.MethodPut, http.MethodPost).Handler(http.StripPrefix("/files/data/", http.HandlerFunc(files.UploadFile)))

		r.Name("CreateSource").Path("/files/sources").Methods(http.MethodPut).HandlerFunc(files.CreateSource)
		r.Name("ListSources").Path("/files/sources").Methods(http.MethodPost).HandlerFunc(files.ListSources)
		r.Name("GetSource").Path("/files/sources/{id}").Methods(http.MethodGet).HandlerFunc(files.GetSource)
		r.Name("DeleterSource").Path("/files/sources/{id}").Methods(http.MethodDelete).HandlerFunc(files.DeleteSource)
	}

	if options.staticDir != "" {
		staticFilesRouter := http.FileServer(http.Dir(options.staticDir))
		r.NotFoundHandler = staticFilesRouter
	}

	return r
}
