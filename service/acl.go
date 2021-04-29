package service

import (
	"context"
	"database/sql"
	"github.com/omecodes/bome"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/service"
	"github.com/omecodes/store/common"
	"google.golang.org/grpc"
)

type ACLConfig struct {
	Name string

	Domain string
	IP     string

	CAAddress      string
	CAAccess       string
	CASecret       string
	CACertFilename string

	RegistryAddress string
	Port            int

	Database   string
	WorkingDir string
}

func NewACL(config ACLConfig) *ACL {
	return &ACL{
		config: &config,
	}
}

type ACL struct {
	config *ACLConfig
	box    *service.Box
	//manager objects.ACLManager
	db *sql.DB
}

func (a *ACL) init() error {
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

	a.db = common.GetDB(bome.MySQL, a.config.Database)

	var err error
	//a.manager, err = objects.NewACLSQLManager(a.db, bome.MySQL, strings.ToLower(a.config.Name)+"_objects_access_rules")
	return err
}

func (a *ACL) updateGrpcContext(ctx context.Context) (context.Context, error) {
	//ctx = objects.ContextWithACLManager(ctx, a.manager)
	return ctx, nil
}

func (a *ACL) Start() error {
	err := a.init()
	if err != nil {
		return err
	}

	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			//pb.RegisterACLServer(server, objects.NewACLManagerServiceServerHandler())
		},
		ServiceType: common.ServiceTypeACLStore,
		ServiceID:   a.config.Name,
		Name:        a.config.Name + "-grpc",
		Meta:        nil,
	}

	opts := []service.NodeOption{service.Register(true),
		service.WithInterceptor(
			ome.GrpcContextUpdaterFunc(a.updateGrpcContext),
		)}

	return a.box.StartNode(params, opts...)
}

func (a *ACL) Stop() error {
	return a.db.Close()
}
