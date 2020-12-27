package router

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/oms"
	"github.com/omecodes/store/pb"
	"strconv"
	"time"
)

type ParamsHandler struct {
	BaseHandler
}

func (p *ParamsHandler) PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	if object == nil || object.Size() == 0 {
		return "", errors.BadInput
	}

	if security == nil {
		security = new(pb.PathAccessRules)
		security.AccessRules = map[string]*pb.AccessRules{}
	}

	settings := Settings(ctx)
	if settings == nil {
		return "", errors.Internal
	}

	s, err := settings.Get(oms.SettingsDataMaxSizePath)
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return "", errors.Internal
	}

	maxLength, err := strconv.ParseInt(s, 10, 64)
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

func (p *ParamsHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	if patch == nil || patch.GetObjectID() == "" || patch.Size() == 0 || patch.Path() == "" {
		return errors.BadInput
	}

	settings := Settings(ctx)
	s, err := settings.Get(oms.SettingsDataMaxSizePath)
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return errors.Internal
	}

	maxLength, err := strconv.ParseInt(s, 10, 64)
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

func (p *ParamsHandler) GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*oms.Object, error) {
	if id == "" {
		return nil, errors.BadInput
	}
	return p.BaseHandler.GetObject(ctx, id, opts)
}

func (p *ParamsHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	if id == "" {
		return nil, errors.BadInput
	}
	return p.BaseHandler.GetObjectHeader(ctx, id)
}

func (p *ParamsHandler) DeleteObject(ctx context.Context, id string) error {
	if id == "" {
		return errors.BadInput
	}
	return p.next.DeleteObject(ctx, id)
}

func (p *ParamsHandler) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	if opts.Count == 0 {
		opts.Count = 100
	}

	if opts.Before == 0 {
		opts.Before = time.Now().UnixNano() / 1e6
	}
	return p.BaseHandler.ListObjects(ctx, opts)
}

func (p *ParamsHandler) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	if params.MatchedExpression == "" {
		return nil, errors.BadInput
	}

	if opts.Before == 0 {
		opts.Before = time.Now().UnixNano() / 1e6
	}

	if opts.Count == 0 {
		opts.Count = 100
	}
	return p.BaseHandler.SearchObjects(ctx, params, opts)
}
