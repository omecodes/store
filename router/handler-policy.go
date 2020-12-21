package router

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
)

type policyHandler struct {
	BaseHandler
}

func (p *policyHandler) isAdmin(ctx context.Context) bool {
	authCEL := AuthInfo(ctx)
	if authCEL == nil {
		return false
	}
	return authCEL.Uid == "admin"
}

func (p *policyHandler) SetSettings(ctx context.Context, name string, value string, opts oms.SettingsOptions) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.BaseHandler.SetSettings(ctx, name, value, opts)
}

func (p *policyHandler) GetSettings(ctx context.Context, name string) (string, error) {
	if !p.isAdmin(ctx) {
		return "", errors.Forbidden
	}
	return p.BaseHandler.GetSettings(ctx, name)
}

func (p *policyHandler) DeleteSettings(ctx context.Context, name string) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.BaseHandler.DeleteSettings(ctx, name)
}

func (p *policyHandler) ClearSettings(ctx context.Context) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.BaseHandler.ClearSettings(ctx)
}

func (p *policyHandler) PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	ai := AuthInfo(ctx)
	if ai == nil {
		return "", errors.Forbidden
	}

	if !ai.Validated {
		return "", errors.Unauthorized
	}

	docRules := security.AccessRules["$"]
	if docRules == nil {
		docRules = &pb.AccessRules{}
		security.AccessRules["$"] = docRules
	}

	userDefaultRule := fmt.Sprintf("auth.validated && auth.uid=='%s'", ai.Uid)
	if len(docRules.Read) == 0 {
		docRules.Read = append(docRules.Read, userDefaultRule, "auth.worker", "auth.uid=='admin'")
	} else {
		docRules.Read = append(docRules.Write, "auth.worker", "auth.uid=='admin'")
	}

	if len(docRules.Write) == 0 {
		docRules.Write = append(docRules.Write, userDefaultRule, "auth.worker", "auth.uid=='admin'")
	} else {
		docRules.Write = append(docRules.Write, "auth.worker", "auth.uid=='admin'")
	}

	if len(docRules.Delete) == 0 {
		docRules.Delete = append(docRules.Delete, userDefaultRule)
	}

	object.SetCreatedBy(ai.Uid)
	return p.BaseHandler.PutObject(ctx, object, security, opts)
}

func (p *policyHandler) GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*oms.Object, error) {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, id, opts.Path)
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObject(ctx, id, opts)
}

func (p *policyHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_delete, patch.GetObjectID(), "")
	if err != nil {
		return err
	}

	return p.BaseHandler.PatchObject(ctx, patch, opts)
}

func (p *policyHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, id, "")
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObjectHeader(ctx, id)
}

func (p *policyHandler) DeleteObject(ctx context.Context, id string) error {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_delete, id, "")
	if err != nil {
		return err
	}
	return p.BaseHandler.DeleteObject(ctx, id)
}

func (p *policyHandler) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	opts.Filter = oms.FilterObjectFunc(func(o *oms.Object) (bool, error) {
		err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, o.ID(), opts.Path)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	return p.BaseHandler.ListObjects(ctx, opts)
}

func (p *policyHandler) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	if params.MatchedExpression == "false" {
		return &oms.ObjectList{
			Before: opts.Before,
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

	program, err := LoadProgramForSearch(&ctx, params.MatchedExpression)
	if err != nil {
		log.Error("Handler-policy.Search: failed to load CEL program", log.Err(err))
		return nil, err
	}

	lOpts := oms.ListOptions{
		Path:   opts.Path,
		Before: opts.Before,
		Count:  opts.Count,
	}
	lOpts.Filter = oms.FilterObjectFunc(func(o *oms.Object) (bool, error) {
		err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, o.ID(), opts.Path)
		if err != nil {
			return false, err
		}

		var object = map[string]interface{}{}
		err = json.NewDecoder(o.GetContent()).Decode(&object)
		if err != nil {
			return false, err
		}

		vars := map[string]interface{}{"o": object}
		out, details, err := program.Eval(vars)
		if err != nil {
			log.Error("cel execution", log.Field("details", details))
			return false, err
		}
		return out.Value().(bool), nil
	})
	return p.BaseHandler.ListObjects(ctx, lOpts)
}
