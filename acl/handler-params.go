package acl

import (
	"context"
	"github.com/omecodes/errors"
	pb "github.com/omecodes/store/gen/go/proto"
)

type ParamsHandler struct {
	BaseHandler
}

func (p *ParamsHandler) SaveNamespaceConfig(ctx context.Context, cfg *pb.NamespaceConfig, opts SaveNamespaceConfigOptions) error {
	if cfg == nil {
		return errors.BadRequest("acl.SaveNamespaceConfig: a namespace config is required")
	}

	if cfg.Namespace == "" {
		return errors.BadRequest("acl.SaveNamespaceConfig: the namespace config name field is required")
	}

	if len(cfg.Relations) == 0 {
		return errors.BadRequest("acl.SaveNamespaceConfig: the namespace config must have at least on relation defined")
	}

	for _, rel := range cfg.Relations {
		if rel.Name == "" {
			return errors.BadRequest("acl.SaveNamespaceConfig: the namespace config has one or more relation definitions with an empty name")
		}
	}

	return p.BaseHandler.SaveNamespaceConfig(ctx, cfg, opts)
}

func (p *ParamsHandler) GetNamespaceConfig(ctx context.Context, namespaceID string, opts SaveNamespaceOptions) (*pb.NamespaceConfig, error) {
	if namespaceID == "" {
		return nil, errors.BadRequest("acl.GetNamespaceConfig: a namespace id is required")
	}

	return p.BaseHandler.GetNamespaceConfig(ctx, namespaceID, opts)
}

func (p *ParamsHandler) DeleteNamespaceConfig(ctx context.Context, namespaceID string, opts DeleteNamespaceOptions) error {
	if namespaceID == "" {
		return errors.BadRequest("acl.DeleteNamespaceConfig: a namespace id is required")
	}
	return p.BaseHandler.DeleteNamespaceConfig(ctx, namespaceID, opts)
}

func (p *ParamsHandler) SaveACL(ctx context.Context, a *pb.ACL, opts SaveACLOptions) error {
	if a == nil {
		return errors.BadRequest("acl.SaveACL: an acl is required")
	}

	if a.Subject == "" || a.Relation == "" || a.Object == "" {
		return errors.BadRequest("acl.SaveACL: the request acl has one or more empty fields")
	}

	return p.BaseHandler.SaveACL(ctx, a, opts)
}

func (p *ParamsHandler) DeleteACL(ctx context.Context, a *pb.ACL, opts DeleteACLOptions) error {
	if a == nil {
		return errors.BadRequest("acl.DeleteACL: an acl is required")
	}

	if a.Subject == "" || a.Relation == "" || a.Object == "" {
		return errors.BadRequest("acl.DeleteACL: the request acl has one or more empty fields")
	}

	return p.BaseHandler.DeleteACL(ctx, a, opts)
}

func (p *ParamsHandler) CheckACL(ctx context.Context, subject string, set *pb.SubjectSet, opts CheckACLOptions) (bool, error) {
	if subject == "" {
		return false, errors.BadRequest("acl.CheckACL: a subject is required")
	}

	if set == nil {
		return false, errors.BadRequest("acl.CheckACL: a subject set is required")
	}

	if set.Relation == "" || set.Object == "" {
		return false, errors.BadRequest("acl.CheckACL: the subject set has one or more empty fields")
	}

	return p.BaseHandler.CheckACL(ctx, subject, set, opts)
}

func (p *ParamsHandler) GetSubjectSet(ctx context.Context, set *pb.SubjectSet, opts GetSubjectSetOptions) ([]string, error) {
	if set == nil {
		return nil, errors.BadRequest("acl.GetSubjectSet: a subject set is required")
	}

	if set.Relation == "" || set.Object == "" {
		return nil, errors.BadRequest("acl.GetSubjectSet: the subject set has one or more empty fields")
	}

	return p.BaseHandler.GetSubjectSet(ctx, set, opts)
}

func (p *ParamsHandler) GetObjectSet(ctx context.Context, set *pb.ObjectSet, opts GetObjectSetOptions) ([]string, error) {
	if set == nil {
		return nil, errors.BadRequest("acl.GetObjectSet: an object set is required")
	}

	if set.Relation == "" || set.Subject == "" {
		return nil, errors.BadRequest("acl.GetObjectSet: the object set has one or more empty field")
	}

	return p.BaseHandler.GetObjectSet(ctx, set, opts)
}
