package service

import (
	"context"
	"database/sql"
	"github.com/omecodes/bome"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/service"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/files"
	pb "github.com/omecodes/store/gen/go/proto"
	"github.com/omecodes/store/objects"
	"google.golang.org/grpc"
	"strings"
)

type FileAccessesConfig struct {
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

func NewSources(config FileAccessesConfig) *FileAccesses {
	return &FileAccesses{config: &config}
}

type FileAccesses struct {
	config  *FileAccessesConfig
	box     *service.Box
	manager objects.ACLManager
	db      *sql.DB
}

func (s *FileAccesses) init() error {
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

func (s *FileAccesses) updateGrpcContext(ctx context.Context) (context.Context, error) {
	ctx = objects.ContextWithACLManager(ctx, s.manager)
	return ctx, nil
}

func (s *FileAccesses) Start() error {
	err := s.init()
	if err != nil {
		return err
	}

	params := &service.NodeParams{
		RegisterHandlerFunc: func(server *grpc.Server) {
			pb.RegisterAccessManagerServer(server, files.NewAccessServerHandler())
		},
		ServiceType: common.ServiceTypeFileSources,
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

func (s *FileAccesses) Stop() error {
	return s.db.Close()
}
