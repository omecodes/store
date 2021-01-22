package oms

import (
	"context"
	"database/sql"

	"google.golang.org/grpc"

	"github.com/omecodes/bome"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/service"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/objects"
)

type StoreConfig struct {
	Name           string
	WorkingDir     string
	RegAddr        string
	CaAddr         string
	CaApiSecret    string
	CaApiKey       string
	CaCertFilename string
	Domain         string
	IP             string
	ACLPort        int
	ObjectsPort    int
	DBUri          string
}

func NewMSStore(config *StoreConfig) *MSStore {
	return &MSStore{config: config}
}

type MSStore struct {
	config      *StoreConfig
	box         *service.Box
	accessStore objects.ACLManager
	objects     objects.DB
}

func (s *MSStore) init() error {
	var err error

	db, err := sql.Open(bome.MySQL, s.config.DBUri)
	if err != nil {
		return err
	}

	s.accessStore, err = objects.NewSQLACLStore(db, bome.MySQL, "objects_acl")
	if err != nil {
		return err
	}

	s.objects, err = objects.NewSqlDB(db, bome.MySQL, "objects")
	if err != nil {
		return err
	}

	s.box = service.CreateBox(
		service.Dir(s.config.WorkingDir),
		service.Ip(s.config.IP),
		service.Domain(s.config.Domain),
		service.RegAddr(s.config.RegAddr),
		service.Name(s.config.Name),
		service.CAApiKey(s.config.CaApiKey),
		service.CAApiSecret(s.config.CaApiSecret),
		service.CAAddr(s.config.CaAddr),
		service.CACertFile(s.config.CaCertFilename),
	)
	return nil
}

func (s *MSStore) updateGrpcContext(ctx context.Context) (context.Context, error) {
	ctx = service.ContextWithBox(ctx, s.box)
	ctx = objects.ContextWithACLStore(ctx, s.accessStore)
	ctx = objects.ContextWithStore(ctx, s.objects)
	ctx = objects.WithRouterProvider(ctx, objects.RouterProvideFunc(
		func(ctx context.Context) objects.Router {
			return objects.NewCustomObjectsRouter(&objects.ExecHandler{})
		},
	))
	return ctx, nil
}

func (s *MSStore) startACLService() error {
	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			objects.RegisterACLServer(server, objects.NewUnitServerHandler())
		},
		ServiceType: objects.ServiceTypeACL,
		ServiceID:   s.config.Name + "-acl",
		Name:        s.config.Name + "-grpc",
		Meta:        nil,
	}
	opts := []service.NodeOption{service.Register(true),
		service.WithInterceptor(
			ome.GrpcContextUpdaterFunc(
				func(ctx context.Context) (context.Context, error) {
					ctx = objects.ContextWithACLStore(ctx, s.accessStore)
					return ctx, nil
				}),
		),
	}
	return s.box.StartNode(params, opts...)
}

func (s *MSStore) startObjectsService() error {
	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			objects.RegisterHandlerUnitServer(server, objects.NewStoreGrpcHandler())
		},
		ServiceType: objects.ServiceTypeObjects,
		ServiceID:   s.config.Name + "-objects",
		Name:        s.config.Name + "-grpc",
		Meta:        nil,
	}
	opts := []service.NodeOption{service.Register(true),
		service.WithInterceptor(
			ome.GrpcContextUpdaterFunc(s.updateGrpcContext),
			ome.GrpcContextUpdaterFunc(auth.UpdateFromMeta),
		),
	}
	return s.box.StartNode(params, opts...)
}

func (s *MSStore) Start() error {
	err := s.init()
	if err != nil {
		return err
	}

	err = s.startACLService()
	if err != nil {
		return err
	}

	return s.startObjectsService()
}

func (s *MSStore) Stop() error {
	s.box.Stop()
	return nil
}
