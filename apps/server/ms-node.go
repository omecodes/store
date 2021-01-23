package server

import (
	"context"

	"google.golang.org/grpc"

	ome "github.com/omecodes/libome"
	"github.com/omecodes/service"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/objects"
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
	config      *NodeConfig
	box         *service.Box
	accessStore objects.ACLManager
	objects     objects.DB
	reg         ome.Registry
}

func (n *MSNode) init() error {
	var err error
	n.accessStore = objects.NewACLGrpcClient()
	n.objects = objects.NewDBGrpcClient()

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
	ctx = WithRegistry(ctx, n.reg)
	ctx = objects.ContextWithACLStore(ctx, n.accessStore)
	ctx = objects.ContextWithStore(ctx, n.objects)
	ctx = service.ContextWithBox(ctx, n.box)
	ctx = objects.WithRouterProvider(ctx, n)
	ctx = objects.WithACLGrpcClientProvider(ctx, &objects.DefaultACLGrpcProvider{})
	ctx = objects.WithRouterGrpcClientProvider(ctx, &objects.DefaultRouterGrpcProvider{})
	return ctx, nil
}

func (n *MSNode) GetRouter(ctx context.Context) objects.Router {
	return objects.NewCustomObjectsRouter(
		&objects.ExecHandler{},
		objects.WithDefaultObjectsPolicyHandler(),
	)
}

func (n *MSNode) Start() error {
	err := n.init()
	if err != nil {
		return err
	}

	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			objects.RegisterHandlerUnitServer(server, objects.NewHandler())
		},
		ServiceType: objects.ServiceTypeHandler,
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
