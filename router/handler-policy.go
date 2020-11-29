package router

import (
	"context"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/jsonpb"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/libome"
	"github.com/omecodes/omestore/oms"
)

type policyHandler struct {
	base
}

func (p *policyHandler) isAdmin(ctx context.Context) bool {
	authCEL := authInfo(ctx)
	if authCEL == nil {
		return false
	}
	return authCEL.Uid == "admin"
}

// Route Handler methods
func (p *policyHandler) SetSettings(ctx context.Context, value *oms.JSON, opts oms.SettingsOptions) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}

	env := celPolicyEnv(ctx)
	if env == nil {
		return errors.Internal
	}
	return p.base.SetSettings(ctx, value, opts)
}

func (p *policyHandler) GetSettings(ctx context.Context, opts oms.SettingsOptions) (*oms.JSON, error) {
	if !p.isAdmin(ctx) {
		return nil, errors.Forbidden
	}
	return p.base.GetSettings(ctx, opts)
}

func (p *policyHandler) RegisterWorker(ctx context.Context, info *oms.JSON) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.base.RegisterWorker(ctx, info)
}

func (p *policyHandler) ListWorkers(ctx context.Context) ([]*oms.JSON, error) {
	if !p.isAdmin(ctx) {
		cred := ome.CredentialsFromContext(ctx)
		if cred == nil {
			return nil, errors.Forbidden
		}
	}
	return p.base.ListWorkers(ctx)
}

func (p *policyHandler) PutObject(ctx context.Context, object *oms.Object, security *oms.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	ai := authInfo(ctx)
	if ai == nil {
		return "", errors.Forbidden
	}

	if !ai.Validated {
		return "", errors.Unauthorized
	}

	docRules := security.AccessRules["$"]
	if docRules == nil {
		docRules = &oms.AccessRules{}
		security.AccessRules["$"] = docRules
	}

	if len(docRules.Read) == 0 && len(docRules.Write) == 0 {
		docRules.Write = append(docRules.Write, "auth.user==\"admin\", auth.validated && data.created_by=auth.user", "auth.worker")
		docRules.Read = append(docRules.Read, "auth.user==\"admin\",  auth.validated && data.created_by=auth.user", "auth.worker")
	} else {
		docRules.Read = append(docRules.Write, "auth.worker")
		docRules.Write = append(docRules.Write, "auth.worker")
	}

	return p.base.PutObject(ctx, object, security, opts)
}

func (p *policyHandler) GetObject(ctx context.Context, id string, opts oms.GetDataOptions) (*oms.Object, error) {
	if !p.isAdmin(ctx) {
		err := assetActionAllowedOnObject(&ctx, oms.AllowedTo_read, id, opts.Path)
		if err != nil {
			return nil, err
		}
	}
	return p.base.GetObject(ctx, id, opts)
}

func (p *policyHandler) GetObjectHeader(ctx context.Context, id string) (*oms.Header, error) {
	if p.isAdmin(ctx) {
		return p.base.GetObjectHeader(ctx, id)
	}

	err := assetActionAllowedOnObject(&ctx, oms.AllowedTo_read, id, "")
	if err != nil {
		return nil, err
	}
	return p.base.GetObjectHeader(ctx, id)
}

func (p *policyHandler) DeleteObject(ctx context.Context, id string) error {
	if !p.isAdmin(ctx) {
		err := assetActionAllowedOnObject(&ctx, oms.AllowedTo_write, id, "")
		if err != nil {
			return err
		}
	}
	return p.base.DeleteObject(ctx, id)
}

func (p *policyHandler) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	if !p.isAdmin(ctx) {
		return p.base.ListObjects(ctx, opts)
	}
	newOpts := oms.ListOptions{
		Filter: oms.FilterObjectFunc(func(o *oms.Object) (bool, error) {
			err := assetActionAllowedOnObject(&ctx, oms.AllowedTo_read, o.ID(), opts.Path)
			if err != nil {
				if err == errors.Unauthorized || err == errors.Forbidden {
					return false, nil
				}
				return false, err
			}
			return true, nil
		}),
		Path:   opts.Path,
		Before: opts.Before,
		Count:  opts.Count,
	}
	return p.base.ListObjects(ctx, newOpts)
}

func (p *policyHandler) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	if p.isAdmin(ctx) {
		return p.base.SearchObjects(ctx, params, opts)
	}

	if params.MatchedExpression == "false" {
		return &oms.ObjectList{
			Before: opts.Before,
			Offset: opts.Offset,
			Count:  0,
		}, nil
	}

	if params.MatchedExpression == "true" {
		return p.ListObjects(ctx, oms.ListOptions{
			Path:   opts.Path,
			Before: opts.Before,
			Count:  opts.Count,
		})
	}

	searchEnv := celSearchEnv(ctx)
	ast, issues := searchEnv.Compile(params.MatchedExpression)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	matchProgram, err := searchEnv.Program(ast)
	if err != nil {
		return nil, err
	}

	lOpts := oms.ListOptions{
		Path:   opts.Path,
		Before: opts.Before,
		Count:  opts.Count,
	}
	lOpts.Filter = oms.FilterObjectFunc(func(o *oms.Object) (bool, error) {
		err := assetActionAllowedOnObject(&ctx, oms.AllowedTo_read, o.ID(), opts.Path)
		if err != nil {
			if err == errors.Unauthorized || err == errors.Forbidden {
				return false, nil
			}
			return false, err
		}

		var object types.Struct
		err = jsonpb.Unmarshal(o.Content(), &object)
		if err != nil {
			return false, err
		}

		vars := map[string]interface{}{"o": object}
		out, details, err := matchProgram.Eval(vars)
		if err != nil {
			log.Error("cel execution", log.Field("details", details))
			return false, err
		}

		return out.Value().(bool), nil
	})

	return p.base.ListObjects(ctx, lOpts)
}
