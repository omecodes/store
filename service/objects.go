package service

import (
	"context"
	"database/sql"
	"github.com/omecodes/bome"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/service"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/objects"
	"google.golang.org/grpc"
)

type ObjectsConfig struct {
	Name            string
	Domain          string
	IP              string
	CAAddress       string
	CAAccess        string
	CASecret        string
	CACertFilename  string
	RegistryAddress string
	WorkingDir      string
	Database        string
}

func NewObjects(config ObjectsConfig) *Objects {
	return &Objects{config: &config}
}

type Objects struct {
	config *ObjectsConfig

	box       *service.Box
	db        *sql.DB
	objectsDB objects.DB
}

func (o *Objects) init() error {
	o.db = common.GetDB(bome.MySQL, o.config.Database)

	var err error
	o.objectsDB, err = objects.NewSqlDB(o.db, bome.MySQL, "objects")
	if err != nil {
		return err
	}

	o.box = service.CreateBox(
		service.Dir(o.config.WorkingDir),
		service.Ip(o.config.IP),
		service.Domain(o.config.Domain),
		service.RegAddr(o.config.RegistryAddress),
		service.Name(o.config.Name),
		service.CAApiKey(o.config.CAAccess),
		service.CAApiSecret(o.config.CASecret),
		service.CAAddr(o.config.CAAddress),
		service.CACertFile(o.config.CACertFilename),
	)
	return nil
}

func (o *Objects) updatedGRPCIncomingContext(ctx context.Context) (context.Context, error) {
	ctx = service.ContextWithBox(ctx, o.box)
	ctx = objects.ContextWithStore(ctx, o.objectsDB)
	ctx = objects.ContextWithACLManager(ctx, objects.NewACLManagerServiceClient())
	ctx = objects.WithACLGrpcClientProvider(ctx, &objects.DefaultACLGrpcProvider{})
	ctx = objects.WithACLGrpcClientProvider(ctx, objects.NewDefaultACLGRPCClientProvider(common.ServiceTypeACLStore))
	ctx = objects.ContextWithRouterProvider(ctx, objects.RouterProvideFunc(
		func(ctx context.Context) objects.Router {
			return objects.NewCustomRouter(&objects.ExecHandler{})
		},
	))
	return ctx, nil
}

func (o *Objects) Start() error {
	err := o.init()
	if err != nil {
		return err
	}

	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			objects.RegisterObjectsServer(server, objects.NewGRPCHandler())
		},
		ServiceType: common.ServiceTypeObjectsStorage,
		ServiceID:   o.config.Name,
		Name:        o.config.Name + "-grpc",
		Meta:        nil,
	}

	opts := []service.NodeOption{
		service.Register(true),
		service.WithInterceptor(
			ome.GrpcContextUpdaterFunc(o.updatedGRPCIncomingContext),
			ome.GrpcContextUpdaterFunc(auth.ParseMetaInNewContext),
		),
	}
	return o.box.StartNode(params, opts...)
}
