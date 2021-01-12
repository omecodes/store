package oms

import (
	"context"
	"github.com/google/cel-go/cel"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/service"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/cenv"
	"github.com/omecodes/store/clients"
	"github.com/omecodes/store/common"
	context2 "github.com/omecodes/store/context"
	"github.com/omecodes/store/objects"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/router"
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
	accessStore  acl.Store
	objects      objects.Objects
	reg          ome.Registry
}

func (n *MSNode) init() error {
	var err error
	n.accessStore = acl.NewStoreClient()
	n.objects = objects.NewStoreGrpcClient()

	n.celPolicyEnv, err = cenv.ACLEnv()
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
	ctx = acl.ContextWithStore(ctx, n.accessStore)
	ctx = router.WithCelPolicyEnv(n.celPolicyEnv)(ctx)
	ctx = objects.ContextWithStore(ctx, n.objects)
	ctx = service.ContextWithBox(ctx, n.box)
	ctx = router.WithRouterProvider(ctx, n)
	ctx = clients.WithACLGrpcClientProvider(ctx, &clients.DefaultACLGrpcProvider{})
	ctx = clients.WithRouterGrpcClientProvider(ctx, &clients.DefaultRouterGrpcProvider{})
	return ctx, nil
}

func (n *MSNode) GetRouter(ctx context.Context) router.Router {
	return router.NewCustomRouter(
		&router.ExecHandler{},
		router.WithDefaultPoliciesHandler(),
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
		service.WithInterceptor(
			ome.GrpcContextUpdaterFunc(n.updateGrpcContext),
			ome.GrpcContextUpdaterFunc(auth.UpdateFromMeta),
		)}

	return n.box.StartNode(params, opts...)
}

func (n *MSNode) Stop() error {
	return nil
}
