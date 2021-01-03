package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/objects"
	"github.com/omecodes/store/pb"
)

type PolicyHandler struct {
	BaseHandler
}

func (p *PolicyHandler) isAdmin(ctx context.Context) bool {
	authCEL := auth.Get(ctx)
	if authCEL == nil {
		return false
	}
	return authCEL.Uid == "admin"
}

func (p *PolicyHandler) PutObject(ctx context.Context, object *pb.Object, security *pb.PathAccessRules, opts objects.PutDataOptions) (string, error) {
	ai := auth.Get(ctx)
	if ai == nil {
		return "", errors.Forbidden
	}

	docRules := security.AccessRules["$"]
	if docRules == nil {
		docRules = &pb.AccessRules{}
		security.AccessRules["$"] = docRules
	}

	userDefaultRule := fmt.Sprintf("auth.uid=='%s'", ai.Uid)
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

	object.Header.CreatedBy = ai.Uid
	return p.BaseHandler.PutObject(ctx, object, security, opts)
}

func (p *PolicyHandler) GetObject(ctx context.Context, id string, opts objects.GetObjectOptions) (*pb.Object, error) {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, id, opts.Path)
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObject(ctx, id, opts)
}

func (p *PolicyHandler) PatchObject(ctx context.Context, patch *pb.Patch, opts objects.PatchOptions) error {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_delete, patch.ObjectId, "")
	if err != nil {
		return err
	}

	return p.BaseHandler.PatchObject(ctx, patch, opts)
}

func (p *PolicyHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, id, "")
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObjectHeader(ctx, id)
}

func (p *PolicyHandler) DeleteObject(ctx context.Context, id string) error {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_delete, id, "")
	if err != nil {
		return err
	}
	return p.BaseHandler.DeleteObject(ctx, id)
}

func (p *PolicyHandler) ListObjects(ctx context.Context, opts objects.ListOptions) (*pb.ObjectList, error) {
	opts.Filter = objects.FilterObjectFunc(func(o *pb.Object) (bool, error) {
		err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, o.Header.Id, opts.Path)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	return p.BaseHandler.ListObjects(ctx, opts)
}

func (p *PolicyHandler) SearchObjects(ctx context.Context, params objects.SearchParams, opts objects.SearchOptions) (*pb.ObjectList, error) {
	lOpts := objects.ListOptions{
		Path:   opts.Path,
		Before: opts.Before,
		After:  opts.After,
		Count:  opts.Count,
	}

	if params.Condition == "false" {
		return &pb.ObjectList{
			Before: opts.Before,
		}, nil
	}

	if params.Condition == "true" {
		return p.ListObjects(ctx, lOpts)
	}

	program, err := LoadProgramForSearch(&ctx, params.Condition)
	if err != nil {
		log.Error("Handler-policy.Search: failed to load CEL program", log.Err(err))
		return nil, err
	}

	lOpts.Filter = objects.FilterObjectFunc(func(o *pb.Object) (bool, error) {
		err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, o.Header.Id, opts.Path)
		if err != nil {
			return false, err
		}

		var object = map[string]interface{}{}
		err = json.NewDecoder(bytes.NewBufferString(o.Data)).Decode(&object)
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
