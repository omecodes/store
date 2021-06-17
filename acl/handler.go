package acl

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
)

type Handler interface {
	SaveNamespaceConfig(ctx context.Context, cfg *pb.NamespaceConfig, opts SaveNamespaceConfigOptions) error
	GetNamespaceConfig(ctx context.Context, namespaceID string, opts GetNamespaceOptions) (*pb.NamespaceConfig, error)
	DeleteNamespaceConfig(ctx context.Context, namespaceID string, opts DeleteNamespaceOptions) error

	SaveACL(ctx context.Context, a *pb.ACL, opts SaveACLOptions) error
	DeleteACL(ctx context.Context, a *pb.ACL, opts DeleteACLOptions) error
	CheckACL(ctx context.Context, subjectName string, set *pb.SubjectSet, opts CheckACLOptions) (bool, error)
	GetObjectACL(ctx context.Context, objectID string, opts GetObjectACLOptions) ([]*pb.ACL, error)
	GetSubjectACL(ctx context.Context, subjectID string, opts GetSubjectACLOptions) ([]*pb.ACL, error)

	GetSubjectNames(ctx context.Context, set *pb.SubjectSet, opts GetSubjectsNamesOptions) ([]string, error)
	GetObjectNames(ctx context.Context, set *pb.ObjectSet, opts GetObjectsSetOptions) ([]string, error)
}
