package oms

import (
	"context"
	"database/sql"
	"github.com/omecodes/bome"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/omestore/acl"
	"github.com/omecodes/omestore/auth"
	"github.com/omecodes/omestore/common"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/omestore/router"
	"github.com/omecodes/service"
	"google.golang.org/grpc"
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
	accessStore acl.Store
	objects     oms.Objects
}

func (s *MSStore) init() error {
	var err error

	db, err := sql.Open(bome.MySQL, s.config.DBUri)
	if err != nil {
		return err
	}

	s.objects, err = oms.NewSQLObjects(db, bome.MySQL, "objects")
	if err != nil {
		return err
	}

	s.accessStore, err = acl.NewSQLStore(db, bome.MySQL, "objects_acl")
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
	ctx = acl.ContextWithStore(ctx, s.accessStore)
	ctx = oms.ContextWithStore(ctx, s.objects)
	ctx = router.WithRouterProvider(ctx, router.ProviderFunc(
		func(ctx context.Context) router.Router {
			return router.NewCustomRouter(&router.ExecHandler{})
		},
	))
	return ctx, nil
}

func (s *MSStore) startACLService() error {
	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			pb.RegisterACLServer(server, acl.NewUnitServerHandler())
		},
		ServiceType: common.ServiceTypeACL,
		ServiceID:   s.config.Name + "-acl",
		Name:        s.config.Name + "-grpc",
		Meta:        nil,
	}
	opts := []service.NodeOption{service.Register(true),
		service.WithInterceptor(
			ome.GrpcContextUpdaterFunc(
				func(ctx context.Context) (context.Context, error) {
					ctx = acl.ContextWithStore(ctx, s.accessStore)
					return ctx, nil
				}),
		),
	}
	return s.box.StartNode(params, opts...)
}

func (s *MSStore) startObjectsService() error {
	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			pb.RegisterHandlerUnitServer(server, oms.NewStoreGrpcHandler())
		},
		ServiceType: common.ServiceTypeObjects,
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
