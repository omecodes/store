package router

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
	"strconv"
	"time"
)

type paramsHandler struct {
	base
}

func (p *paramsHandler) RegisterWorker(ctx context.Context, info *oms.JSON) error {
	if info == nil {
		return errors.BadInput
	}
	return p.base.RegisterWorker(ctx, info)
}

func (p *paramsHandler) SetSettings(ctx context.Context, name string, value string, opts oms.SettingsOptions) error {
	if name == "" || value == "" {
		return errors.BadInput
	}
	return p.base.SetSettings(ctx, name, value, opts)
}

func (p *paramsHandler) DeleteSettings(ctx context.Context, name string) error {
	if name == "" {
		return errors.BadInput
	}
	return p.base.DeleteSettings(ctx, name)
}

func (p *paramsHandler) GetSettings(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", errors.BadInput
	}
	return p.base.GetSettings(ctx, name)
}

func (p *paramsHandler) PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	if object == nil || object.Size() == 0 {
		return "", errors.BadInput
	}

	if security == nil {
		security = new(pb.PathAccessRules)
		security.AccessRules = map[string]*pb.AccessRules{}
	}

	route := Route(SkipPoliciesCheck(), SkipParamsCheck())
	s, err := route.GetSettings(ctx, oms.SettingsDataMaxSizePath)
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

func (p *paramsHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	if patch == nil || patch.GetObjectID() == "" || patch.Size() == 0 || patch.Path() == "" {
		return errors.BadInput
	}

	route := Route(SkipPoliciesCheck(), SkipParamsCheck())
	s, err := route.GetSettings(ctx, oms.SettingsDataMaxSizePath)
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

func (p *paramsHandler) GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*oms.Object, error) {
	if id == "" {
		return nil, errors.BadInput
	}
	return p.base.GetObject(ctx, id, opts)
}

func (p *paramsHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
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

func (p *paramsHandler) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	if opts.Count == 0 {
		opts.Count = 100
	}

	if opts.Before == 0 {
		opts.Before = time.Now().UnixNano() / 1e6
	}
	return p.base.ListObjects(ctx, opts)
}

func (p *paramsHandler) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	if params.MatchedExpression == "" {
		return nil, errors.BadInput
	}

	if opts.Before == 0 {
		opts.Before = time.Now().UnixNano() / 1e6
	}

	if opts.Count == 0 {
		opts.Count = 100
	}
	return p.base.SearchObjects(ctx, params, opts)
}
