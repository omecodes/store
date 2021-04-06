package service

import (
	"context"
	"github.com/gorilla/mux"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/service"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/objects"
	"google.golang.org/grpc"
	"net/http"
)

type AccessConfig struct {
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

type Access struct {
	config     *AccessConfig
	box        *service.Box
	httpServer *http.Server
	gRPCServer *grpc.Server
}

func (a *Access) init() error {
	a.box = service.CreateBox(
		service.Dir(a.config.WorkingDir),
		service.Ip(a.config.IP),
		service.Domain(a.config.Domain),
		service.RegAddr(a.config.RegistryAddress),
		service.Name(a.config.Name),
		service.CAApiKey(a.config.CAAccess),
		service.CAApiSecret(a.config.CASecret),
		service.CAAddr(a.config.CAAddress),
		service.CACertFile(a.config.CACertFilename),
	)
	return nil
}

func (a *Access) updateIncomingRequestContext(ctx context.Context) context.Context {
	ctx = service.ContextWithBox(ctx, a.box)

	ctx = files.ContextWithRouterProvider(ctx, files.RouterProvideFunc(a.provideFilesRouter))
	ctx = files.ContextWithSourceManager(ctx, files.NewSourcesManagerServiceClient())
	ctx = files.ContextWithSourcesServiceClientProvider(ctx, &files.DefaultSourcesServiceClientProvider{})
	ctx = files.ContextWithTransfersServiceClientProvider(ctx, &files.DefaultTransfersServiceClientProvider{})
	ctx = files.ContextWithClientProvider(ctx, &files.DefaultClientProvider{})

	ctx = objects.ContextWithRouterProvider(ctx, objects.RouterProvideFunc(a.provideObjectsRouter))
	ctx = objects.ContextWithACLManager(ctx, objects.NewACLManagerServiceClient())
	ctx = objects.WithACLGrpcClientProvider(ctx, &objects.DefaultACLGrpcProvider{})
	ctx = objects.WithObjectsGrpcClientProvider(ctx, &objects.DefaultClientProvider{})

	return ctx
}

func (a *Access) updateIncomingGrpcRequestContext(ctx context.Context) (context.Context, error) {
	return a.updateIncomingRequestContext(ctx), nil
}

func (a *Access) middlewareUpdateContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(a.updateIncomingRequestContext(r.Context())))
	})
}

func (a *Access) provideFilesRouter(ctx context.Context) files.Router {
	return files.NewCustomRouter(
		files.NewHandlerServiceClient(common.ServiceTypeFilesStorage),
		files.WithDefaultPolicyHandler(),
	)
}

func (a *Access) provideObjectsRouter(ctx context.Context) objects.Router {
	return objects.NewCustomRouter(
		objects.NewGRPCObjectsClientHandler(common.ServiceTypeObjectsStorage),
		objects.WithDefaultPolicyHandler(),
	)
}

func (a *Access) filesTransferHandler() *mux.Router {
	r := mux.NewRouter()
	dataRoute := r.PathPrefix("/api/files/data/").Subrouter()
	dataRoute.Name("Download").Methods(http.MethodGet).Handler(http.StripPrefix("/api/files/data/", http.HandlerFunc(files.HTTPHandleDownloadFile)))
	dataRoute.Name("Upload").Methods(http.MethodPut, http.MethodPost).Handler(http.StripPrefix("/api/files/data/", http.HandlerFunc(files.HTTPHandleUploadFile)))
	return r
}

func (a *Access) startHTTPTransferServer() error {
	params := &service.HTTPServerParams{
		MiddlewareList: []mux.MiddlewareFunc{
			a.middlewareUpdateContext,
			auth.ServiceMiddleware,
		},
		ProvideRouter: a.filesTransferHandler,
		Security:      ome.Security_MutualTls,
		ServiceType:   common.ServiceTypeACL,
		ServiceID:     a.config.Name,
		Name:          a.config.Name + "-http",
		Meta:          nil,
	}
	return a.box.StartHTTPServer(params, service.Register(true))
}

func (a *Access) startGRPCServer() error {
	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			objects.RegisterObjectsServer(server, objects.NewGRPCHandler())
			files.RegisterFilesServer(server, files.NewFilesServerHandler())
		},
		ServiceType: common.ServiceTypeACL,
		ServiceID:   a.config.Name,
		Name:        a.config.Name + "-grpc",
		Meta:        nil,
	}

	opts := []service.NodeOption{
		service.Register(true),
		service.WithInterceptor(
			ome.GrpcContextUpdaterFunc(a.updateIncomingGrpcRequestContext),
			ome.GrpcContextUpdaterFunc(auth.ParseMetaInNewContext),
		),
	}
	return a.box.StartNode(params, opts...)
}

func (a *Access) Start() error {
	err := a.init()
	if err != nil {
		return err
	}

	err = a.startGRPCServer()
	if err != nil {
		return err
	}

	return a.startHTTPTransferServer()
}
