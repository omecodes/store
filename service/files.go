package service

import (
	"context"
	"github.com/gorilla/mux"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/service"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/files"
	"google.golang.org/grpc"
	"net/http"
)

type FilesConfig struct {
	Name            string
	Domain          string
	IP              string
	CAAddress       string
	CAAccess        string
	CASecret        string
	CACertFilename  string
	RegistryAddress string
	WorkingDir      string
}

func NewFiles(config FilesConfig) *Files {
	return &Files{config: &config}
}

type Files struct {
	config *FilesConfig
	box    *service.Box
}

func (f *Files) init() error {
	f.box = service.CreateBox(
		service.Dir(f.config.WorkingDir),
		service.Ip(f.config.IP),
		service.Domain(f.config.Domain),
		service.RegAddr(f.config.RegistryAddress),
		service.Name(f.config.Name),
		service.CAApiKey(f.config.CAAccess),
		service.CAApiSecret(f.config.CASecret),
		service.CAAddr(f.config.CAAddress),
		service.CACertFile(f.config.CACertFilename),
	)
	return nil
}

func (f *Files) updateIncomingRequestContext(ctx context.Context) context.Context {
	ctx = service.ContextWithBox(ctx, f.box)
	ctx = files.ContextWithSourceManager(ctx, files.NewSourcesManagerServiceClient(common.ServiceTypeFileSources))
	ctx = files.ContextWithSourcesServiceClientProvider(ctx, &files.DefaultSourcesServiceClientProvider{})
	ctx = files.ContextWithRouterProvider(ctx, files.RouterProvideFunc(
		func(ctx context.Context) files.Router {
			return files.NewCustomRouter(&files.ExecHandler{})
		},
	))
	return ctx
}

func (f *Files) updateIncomingGrpcRequestContext(ctx context.Context) (context.Context, error) {
	return f.updateIncomingRequestContext(ctx), nil
}

func (f *Files) middlewareUpdateContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(f.updateIncomingRequestContext(r.Context())))
	})
}

func (f *Files) filesTransferHandler() *mux.Router {
	r := mux.NewRouter()
	dataRoute := r.PathPrefix("/api/files/data/").Subrouter()
	dataRoute.Name("Download").Methods(http.MethodGet).Handler(http.StripPrefix("/api/files/data/", http.HandlerFunc(files.HTTPHandleDownloadFile)))
	dataRoute.Name("Upload").Methods(http.MethodPut, http.MethodPost).Handler(http.StripPrefix("/api/files/data/", http.HandlerFunc(files.HTTPHandleUploadFile)))
	return r
}

func (f *Files) startHTTPTransferServer() error {
	params := &service.HTTPServerParams{
		MiddlewareList: []mux.MiddlewareFunc{
			f.middlewareUpdateContext,
			auth.ServiceMiddleware,
		},
		ProvideRouter: f.filesTransferHandler,
		Security:      ome.Security_MutualTls,
		ServiceType:   common.ServiceTypeFilesStorage,
		ServiceID:     f.config.Name,
		Name:          f.config.Name + "-http",
		Meta:          nil,
	}
	return f.box.StartHTTPServer(params, service.Register(true))
}

func (f *Files) startGRPCServer() error {
	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			files.RegisterFilesServer(server, files.NewFilesServerHandler())
		},
		ServiceType: common.ServiceTypeFilesStorage,
		ServiceID:   f.config.Name,
		Name:        f.config.Name + "-grpc",
		Meta:        nil,
	}

	opts := []service.NodeOption{
		service.Register(true),
		service.WithInterceptor(
			ome.GrpcContextUpdaterFunc(f.updateIncomingGrpcRequestContext),
			ome.GrpcContextUpdaterFunc(auth.ParseMetaInNewContext),
		),
	}
	return f.box.StartNode(params, opts...)
}

func (f *Files) Start() error {
	err := f.init()
	if err != nil {
		return err
	}

	err = f.startGRPCServer()
	if err != nil {
		return err
	}

	return f.startHTTPTransferServer()
}
