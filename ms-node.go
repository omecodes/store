package oms

import (
	"context"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/omestore/common"
	context2 "github.com/omecodes/omestore/context"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/omestore/router"
	"github.com/omecodes/omestore/services/acl"
	"github.com/omecodes/omestore/services/objects"
	"github.com/omecodes/service"
	"google.golang.org/grpc"
)

type NodeConfig struct {
	Name           string
	WorkingDir     string
	RegAddr        string
	CaAddr         string
	CaApiSecret    string
	CaApiKey       string
	CaCertFilename string
	Domain         string
	IP             string
	Port           int
}

func NewMSNode(config *NodeConfig) *MSNode {
	return &MSNode{config: config}
}

type MSNode struct {
	config       *NodeConfig
	box          *service.Box
	celPolicyEnv *cel.Env
	celSearchEnv *cel.Env
	accessStore  acl.Store
	objects      oms.Objects
	reg          ome.Registry
}

func (n *MSNode) init() error {
	var err error
	n.accessStore = acl.NewStoreClient()
	n.objects = objects.NewStoreClient()

	n.celPolicyEnv, err = cel.NewEnv(
		cel.Declarations(
			decls.NewVar("auth", decls.NewMapType(decls.String, decls.Dyn)),
			decls.NewVar("data", decls.NewMapType(decls.String, decls.Dyn)),
		),
	)
	if err != nil {
		return err
	}

	n.celSearchEnv, err = cel.NewEnv(
		cel.Declarations(decls.NewVar("o", decls.NewMapType(decls.String, decls.Dyn))))
	if err != nil {
		return err
	}

	n.box = service.CreateBox(
		service.Dir(n.config.WorkingDir),
		service.Ip(n.config.IP),
		service.Domain(n.config.Domain),
		service.RegAddr(n.config.RegAddr),
		service.Name(n.config.Name),
		service.CAApiKey(n.config.CaApiKey),
		service.CAApiSecret(n.config.CaApiSecret),
		service.CAAddr(n.config.CaAddr),
		service.CACertFile(n.config.CaCertFilename),
	)

	n.reg, err = n.box.Registry()
	return err
}

func (n *MSNode) updateGrpcContext(ctx context.Context) (context.Context, error) {
	ctx = context2.WithRegistry(ctx, n.reg)
	ctx = router.WithAccessStore(n.accessStore)(ctx)
	ctx = router.WithCelPolicyEnv(n.celPolicyEnv)(ctx)
	ctx = router.WithCelSearchEnv(n.celSearchEnv)(ctx)
	ctx = router.WithObjectsStore(n.objects)(ctx)
	ctx = service.ContextWithBox(ctx, n.box)
	return ctx, nil
}

func (n *MSNode) GetRouter(ctx context.Context) router.Router {
	return router.NewCustomRouter(
		&router.ExecHandler{},
		router.WithDefaultPoliciesHandler(),
		router.WithDefaultParamsHandler(),
	)
}

func (n *MSNode) Start() error {
	err := n.init()
	if err != nil {
		return err
	}

	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			pb.RegisterHandlerUnitServer(server, NewHandler())
		},
		ServiceType: common.ServiceTypeHandler,
		ServiceID:   n.config.Name,
		Name:        n.config.Name + "-grpc",
		Meta:        nil,
	}
	opts := []service.NodeOption{service.Register(true),
		service.WithInterceptor(ome.GrpcContextUpdaterFunc(n.updateGrpcContext))}

	return n.box.StartNode(params, opts...)
}

func (n *MSNode) Stop() error {
	return nil
}
