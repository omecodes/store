package router

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/oms"
)

type paramsHandler struct {
	base
}

func (p *paramsHandler) SetSettings(ctx context.Context, value *oms.JSON, opts oms.SettingsOptions) error {
	if value == nil {
		return errors.BadInput
	}

	if opts.Path != "" {
		_, allowed := oms.SettingsPathFormats[opts.Path]
		if !allowed {
			return errors.BadInput
		}
	}

	return p.base.SetSettings(ctx, value, opts)
}

func (p *paramsHandler) GetSettings(ctx context.Context, opts oms.SettingsOptions) (*oms.JSON, error) {
	if opts.Path != "" {
		_, exists := oms.SettingsPathFormats[opts.Path]
		if !exists {
			return nil, errors.BadInput
		}
	}
	return p.base.GetSettings(ctx, opts)
}

func (p *paramsHandler) RegisterWorker(ctx context.Context, info *oms.JSON) error {
	if info == nil {
		return errors.BadInput
	}
	return p.base.RegisterWorker(ctx, info)
}

func (p *paramsHandler) PutObject(ctx context.Context, object *oms.Object, security *oms.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	if object == nil || object.ID() == "" || object.Size() == 0 {
		return "", errors.BadInput
	}

	if security == nil {
		security = new(oms.PathAccessRules)
		security.AccessRules = map[string]*oms.AccessRules{}
	}

	route := Route(SkipPoliciesCheck(), SkipParamsCheck())
	s, err := route.GetSettings(ctx, oms.SettingsOptions{Path: oms.SettingsDataMaxSizePath})
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return "", errors.Internal
	}

	maxLength, err := s.ToInt64()
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return "", errors.Internal
	}

	if object.Size() > maxLength {
		log.Error("could not process request. Object too big", log.Field("max", maxLength), log.Field("received", object.Size))
		return "", errors.BadInput
	}

	return p.next.PutObject(ctx, object, security, opts)
}

func (p *paramsHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	if patch.GetObjectID() == "" || patch.Size() == 0 || patch.Path() == "" {
		return errors.BadInput
	}

	route := Route(SkipPoliciesCheck(), SkipParamsCheck())
	s, err := route.GetSettings(ctx, oms.SettingsOptions{Path: oms.SettingsDataMaxSizePath})
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return errors.Internal
	}

	maxLength, err := s.ToInt64()
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return errors.Internal
	}

	if patch.Size() > maxLength {
		log.Error("could not process request. Object too big", log.Field("max", maxLength), log.Field("received", patch.Size()))
		return errors.BadInput
	}

	return p.next.PatchObject(ctx, patch, opts)
}

func (p *paramsHandler) GetObject(ctx context.Context, id string, opts oms.GetDataOptions) (*oms.Object, error) {
	if id == "" {
		return nil, errors.BadInput
	}
	return p.base.GetObject(ctx, id, opts)
}

func (p *paramsHandler) GetObjectHeader(ctx context.Context, id string) (*oms.Header, error) {
	if id == "" {
		return nil, errors.BadInput
	}
	return p.base.GetObjectHeader(ctx, id)
}

func (p *paramsHandler) DeleteObject(ctx context.Context, id string) error {
	if id == "" {
		return errors.BadInput
	}
	return p.next.DeleteObject(ctx, id)
}

func (p *paramsHandler) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	if params.MatchedExpression == "" {
		return nil, errors.BadInput
	}
	return p.base.SearchObjects(ctx, params, opts)
}
