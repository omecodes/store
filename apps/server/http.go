package server

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"

	"github.com/omecodes/store/accounts"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/objects"
)

func filesRouter(r *mux.Router, pathPrefix string, middleware ...mux.MiddlewareFunc) http.Handler {
	treePrefix := path.Join(pathPrefix, "/tree/")

	treeRoute := r.PathPrefix(treePrefix).Subrouter()
	treeRoute.Name("CreateFile").Methods(http.MethodPut).Handler(http.StripPrefix(treePrefix, http.HandlerFunc(files.CreateFile)))
	treeRoute.Name("ListDir").Methods(http.MethodPost).Handler(http.StripPrefix(treePrefix, http.HandlerFunc(files.ListDir)))
	treeRoute.Name("GetFileInfo").Methods(http.MethodGet).Handler(http.StripPrefix(treePrefix, http.HandlerFunc(files.GetFileInfo)))
	treeRoute.Name("DeleteFile").Methods(http.MethodDelete).Handler(http.StripPrefix(treePrefix, http.HandlerFunc(files.DeleteFile)))
	treeRoute.Name("PatchTree").Methods(http.MethodPatch).Handler(http.StripPrefix(treePrefix, http.HandlerFunc(files.PatchFileTree)))

	attrPrefix := path.Join(pathPrefix, "/tree/")
	attrRoute := r.PathPrefix(attrPrefix).Subrouter()
	attrRoute.Name("GetFileAttributes").Methods(http.MethodGet).Handler(http.StripPrefix(attrPrefix, http.HandlerFunc(files.GetFileAttributes)))
	attrRoute.Name("SetFileAttributes").Methods(http.MethodPut).Handler(http.StripPrefix(attrPrefix, http.HandlerFunc(files.SetFileAttributes)))

	dataPrefix := path.Join(pathPrefix, "/data/")
	dataRoute := r.PathPrefix(attrPrefix).Subrouter()
	dataRoute.Name("Download").Methods(http.MethodGet).Handler(http.StripPrefix(dataPrefix, http.HandlerFunc(files.DownloadFile)))
	dataRoute.Name("Upload").Methods(http.MethodPut, http.MethodPost).Handler(http.StripPrefix(dataPrefix, http.HandlerFunc(files.UploadFile)))

	sourcePrefix := path.Join(pathPrefix, "/sources/")
	r.Name("CreateSource").Path(sourcePrefix).Methods(http.MethodPut).HandlerFunc(files.CreateSource)
	r.Name("ListSources").Path(path.Join(pathPrefix, "/sources/")).Methods(http.MethodGet).HandlerFunc(files.ListSources)
	r.Name("GetSource").Path(path.Join(sourcePrefix, "{id}")).Methods(http.MethodGet).HandlerFunc(files.GetSource)
	r.Name("DeleterSource").Path(path.Join(sourcePrefix, "{id}")).Methods(http.MethodDelete).HandlerFunc(files.DeleteSource)

	var handler http.Handler
	handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func objectsRouter(r *mux.Router, pathPrefix string, middleware ...mux.MiddlewareFunc) http.Handler {
	r.Name("SetSettings").Methods(http.MethodPut).Path(path.Join(pathPrefix, "settings")).Handler(http.HandlerFunc(objects.SetSettings))
	r.Name("GetSettings").Methods(http.MethodGet).Path(path.Join(pathPrefix, "settings")).Handler(http.HandlerFunc(objects.GetSettings))

	r.Name("CreateCollection").Methods(http.MethodPut).Path(path.Join(pathPrefix, "collections")).Handler(http.HandlerFunc(objects.CreateCollection))
	r.Name("ListCollections").Methods(http.MethodGet).Path(path.Join(pathPrefix, "collections")).Handler(http.HandlerFunc(objects.ListCollections))
	r.Name("DeleteCollection").Methods(http.MethodGet).Path(path.Join(pathPrefix, "collections", "{id}")).Handler(http.HandlerFunc(objects.DeleteCollection))
	r.Name("GetCollection").Methods(http.MethodGet).Path(path.Join(pathPrefix, "collections", "{id}")).Handler(http.HandlerFunc(objects.GetCollection))

	r.Name("PutObject").Methods(http.MethodPut).Path(path.Join(pathPrefix, "data", "{collection}")).Handler(http.HandlerFunc(objects.PutObject))
	r.Name("PatchObject").Methods(http.MethodPatch).Path(path.Join(pathPrefix, "data", "{collection}", "{id}")).Handler(http.HandlerFunc(objects.PatchObject))
	r.Name("MoveObject").Methods(http.MethodPost).Path(path.Join(pathPrefix, "data", "{collection}", "{id}")).Handler(http.HandlerFunc(objects.MoveObject))
	r.Name("GetObject").Methods(http.MethodGet).Path(path.Join(pathPrefix, "data", "{collection}", "{id}")).Handler(http.HandlerFunc(objects.GetObject))
	r.Name("DeleteObject").Methods(http.MethodDelete).Path(path.Join(pathPrefix, "data", "{collection}", "{id}")).Handler(http.HandlerFunc(objects.DeleteObject))
	r.Name("GetObjects").Methods(http.MethodGet).Path(path.Join(pathPrefix, "data", "{collection}")).Handler(http.HandlerFunc(objects.ListObjects))
	r.Name("SearchObjects").Methods(http.MethodPost).Path(path.Join(pathPrefix, "data", "{collection}")).Handler(http.HandlerFunc(objects.SearchObjects))

	var handler http.Handler
	handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func authRouter(r *mux.Router, pathPrefix string, middleware ...mux.MiddlewareFunc) http.Handler {
	r.Name("SaveAuthProvider").Methods(http.MethodPut).Path(path.Join(pathPrefix, "providers")).Handler(http.HandlerFunc(auth.SaveProvider))
	r.Name("GetAuthProvider").Methods(http.MethodGet).Path(path.Join(pathPrefix, "providers", "{name}")).Handler(http.HandlerFunc(auth.GetProvider))
	r.Name("DeleteAuthProvider").Methods(http.MethodDelete).Path(path.Join(pathPrefix, "providers", "{name}")).Handler(http.HandlerFunc(auth.DeleteProvider))
	r.Name("ListProviders").Methods(http.MethodGet).Path(path.Join(pathPrefix, "providers")).Handler(http.HandlerFunc(auth.ListProviders))
	r.Name("CreateAccess").Methods(http.MethodPut).Path(path.Join(pathPrefix, "access")).Handler(http.HandlerFunc(auth.CreateAccess))
	r.Name("ListAccesses").Methods(http.MethodGet).Path(path.Join(pathPrefix, "accesses")).Handler(http.HandlerFunc(auth.ListAccesses))
	r.Name("DeleteAccess").Methods(http.MethodDelete).Path(path.Join(pathPrefix, "providers", "{key}")).Handler(http.HandlerFunc(auth.DeleteAccess))

	var handler http.Handler
	handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func accountRouter(r *mux.Router, pathPrefix string, middleware ...mux.MiddlewareFunc) http.Handler {
	r.Name("GetAccount").Methods(http.MethodGet).Path(path.Join(pathPrefix, "accounts", "{name}")).Handler(http.HandlerFunc(accounts.GetAccount))
	r.Name("FindAccount").Methods(http.MethodPost).Path(path.Join(pathPrefix, "accounts", "{name}")).Handler(http.HandlerFunc(accounts.FindAccount))
	r.Name("CreateAccount").Methods(http.MethodPut).Path(path.Join(pathPrefix, "accounts", "{name}")).Handler(http.HandlerFunc(accounts.CreateAccount))
	var handler http.Handler
	handler = r
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}
