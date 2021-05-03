package acl

import (
	"context"
	"github.com/omecodes/errors"
	pb "github.com/omecodes/store/gen/go/proto"
)

type ExecHandler struct{}

func (e *ExecHandler) SaveNamespaceConfig(ctx context.Context, cfg *pb.NamespaceConfig, _ SaveNamespaceConfigOptions) error {
	man := GetManager(ctx)
	if man == nil {
		return errors.Internal("acl.SaveNamespaceConfig: could not load acl manager from context")
	}
	return man.SaveNamespaceConfig(ctx, cfg)
}

func (e *ExecHandler) GetNamespaceConfig(ctx context.Context, namespaceID string, opts GetNamespaceOptions) (*pb.NamespaceConfig, error) {
	man := GetManager(ctx)
	if man == nil {
		return nil, errors.Internal("acl.GetNamespaceConfig: could not load acl manager from context")
	}
	return man.GetNamespaceConfig(ctx, namespaceID)
}

func (e *ExecHandler) DeleteNamespaceConfig(ctx context.Context, namespaceID string, opts DeleteNamespaceOptions) error {
	man := GetManager(ctx)
	if man == nil {
		return errors.Internal("acl.DeleteNamespaceConfig: could not load acl manager from context")
	}
	return man.DeleteNamespaceConfig(ctx, namespaceID)
}

func (e *ExecHandler) SaveACL(ctx context.Context, a *pb.ACL, opts SaveACLOptions) error {
	man := GetManager(ctx)
	if man == nil {
		return errors.Internal("acl.SaveACL: could not load acl manager from context")
	}
	return man.SaveACL(ctx, a)
}

func (e *ExecHandler) DeleteACL(ctx context.Context, a *pb.ACL, opts DeleteACLOptions) error {
	man := GetManager(ctx)
	if man == nil {
		return errors.Internal("acl.DeleteACL: could not load acl manager from context")
	}
	return man.DeleteACL(ctx, a)
}

func (e *ExecHandler) CheckACL(ctx context.Context, subjectName string, set *pb.SubjectSet, opts CheckACLOptions) (bool, error) {
	man := GetManager(ctx)
	if man == nil {
		return false, errors.Internal("acl.CheckACL: could not load acl manager from context")
	}
	return man.CheckACL(ctx, subjectName, set)
}

func (e *ExecHandler) GetSubjectNames(ctx context.Context, set *pb.SubjectSet, opts GetSubjectsNamesOptions) ([]string, error) {
	man := GetManager(ctx)
	if man == nil {
		return nil, errors.Internal("acl.GetSubjectsNames: could not load acl manager from context")
	}
	return man.GetSubjectsNames(ctx, set)
}

func (e *ExecHandler) GetObjectNames(ctx context.Context, set *pb.ObjectSet, opts GetObjectsSetOptions) ([]string, error) {
	man := GetManager(ctx)
	if man == nil {
		return nil, errors.Internal("acl.GetObjectsNames: could not load acl manager from context")
	}
	return man.GetObjectsNames(ctx, set)
}
