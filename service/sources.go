package service

import (
	"context"
	"database/sql"
	"github.com/omecodes/bome"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/service"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/objects"
	"google.golang.org/grpc"
	"strings"
)

type SourcesConfig struct {
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

type Sources struct {
	config  *SourcesConfig
	box     *service.Box
	manager objects.ACLManager
	db      *sql.DB
}

func (s *Sources) init() error {
	s.box = service.CreateBox(
		service.Dir(s.config.WorkingDir),
		service.Ip(s.config.IP),
		service.Domain(s.config.Domain),
		service.RegAddr(s.config.RegistryAddress),
		service.Name(s.config.Name),
		service.CAApiKey(s.config.CAAccess),
		service.CAApiSecret(s.config.CASecret),
		service.CAAddr(s.config.CAAddress),
		service.CACertFile(s.config.CACertFilename),
	)

	s.db = common.GetDB(bome.MySQL, s.config.Database)

	var err error
	s.manager, err = objects.NewACLSQLManager(s.db, bome.MySQL, strings.ToLower(s.config.Name)+"_objects_access_rules")
	return err
}

func (s *Sources) updateGrpcContext(ctx context.Context) (context.Context, error) {
	ctx = objects.ContextWithACLManager(ctx, s.manager)
	return ctx, nil
}

func (s *Sources) Start() error {
	err := s.init()
	if err != nil {
		return err
	}

	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			files.RegisterSourcesServer(server, files.NewSourceServerHandler())
		},
		ServiceType: common.ServiceTypeACL,
		ServiceID:   s.config.Name,
		Name:        s.config.Name + "-grpc",
		Meta:        nil,
	}

	opts := []service.NodeOption{service.Register(true),
		service.WithInterceptor(
			ome.GrpcContextUpdaterFunc(s.updateGrpcContext),
		)}

	return s.box.StartNode(params, opts...)
}

func (s *Sources) Stop() error {
	return s.db.Close()
}
