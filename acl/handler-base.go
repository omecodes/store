package acl

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
)

type BaseHandler struct {
	next Handler
}

func (b *BaseHandler) SaveNamespaceConfig(ctx context.Context, cfg *pb.NamespaceConfig, opts SaveNamespaceConfigOptions) error {
	return b.next.SaveNamespaceConfig(ctx, cfg, opts)
}

func (b *BaseHandler) GetNamespaceConfig(ctx context.Context, namespaceID string, opts SaveNamespaceOptions) (*pb.NamespaceConfig, error) {
	return b.next.GetNamespaceConfig(ctx, namespaceID, opts)
}

func (b *BaseHandler) DeleteNamespaceConfig(ctx context.Context, namespaceID string, opts DeleteNamespaceOptions) error {
	return b.next.DeleteNamespaceConfig(ctx, namespaceID, opts)
}

func (b *BaseHandler) SaveACL(ctx context.Context, a *pb.ACL, opts SaveACLOptions) error {
	return b.next.SaveACL(ctx, a, opts)
}

func (b *BaseHandler) DeleteACL(ctx context.Context, a *pb.ACL, opts DeleteACLOptions) error {
	return b.next.DeleteACL(ctx, a, opts)
}

func (b *BaseHandler) CheckACL(ctx context.Context, subjectName string, set *pb.SubjectSet, opts CheckACLOptions) (bool, error) {
	return b.next.CheckACL(ctx, subjectName, set, opts)
}

func (b *BaseHandler) GetSubjectSet(ctx context.Context, set *pb.SubjectSet, opts GetSubjectSetOptions) ([]string, error) {
	return b.next.GetSubjectSet(ctx, set, opts)
}

func (b *BaseHandler) GetObjectSet(ctx context.Context, set *pb.ObjectSet, opts GetObjectSetOptions) ([]string, error) {
	return b.next.GetObjectSet(ctx, set, opts)
}
