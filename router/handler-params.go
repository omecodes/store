package router

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/objects"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/utime"
	"strconv"
)

type ParamsHandler struct {
	BaseHandler
}

func (p *ParamsHandler) PutObject(ctx context.Context, object *pb.Object, security *pb.PathAccessRules, indexes []*pb.Index, opts pb.PutOptions) (string, error) {
	if object == nil || object.Header == nil || object.Header.Size == 0 {
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

	s, err := settings.Get(objects.SettingsDataMaxSizePath)
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return "", errors.Internal
	}

	maxLength, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return "", errors.Internal
	}

	if object.Header.Size > maxLength {
		log.Error("could not process request. Object too big", log.Field("max", maxLength), log.Field("received", object.Header.Size))
		return "", errors.BadInput
	}

	return p.next.PutObject(ctx, object, security, indexes, opts)
}

func (p *ParamsHandler) PatchObject(ctx context.Context, patch *pb.Patch, opts pb.PatchOptions) error {
	if patch == nil || patch.ObjectId == "" || len(patch.Data) == 0 || patch.At == "" {
		return errors.BadInput
	}

	settings := Settings(ctx)
	s, err := settings.Get(objects.SettingsDataMaxSizePath)
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return errors.Internal
	}

	maxLength, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return errors.Internal
	}

	if int64(len(patch.Data)) > maxLength {
		log.Error("could not process request. Object too big", log.Field("max", maxLength), log.Field("received", len(patch.Data)))
		return errors.BadInput
	}

	return p.next.PatchObject(ctx, patch, opts)
}

func (p *ParamsHandler) GetObject(ctx context.Context, id string, opts pb.GetOptions) (*pb.Object, error) {
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

func (p *ParamsHandler) ListObjects(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	if opts.DateOptions.Before == 0 {
		opts.DateOptions.Before = utime.Now()
	}

	if opts.DateOptions.After == 0 {
		opts.DateOptions.After = 1
	}

	/*if opts.Count > 5 || opts.Count == 0 {
		opts.Count = 5
	} */

	return p.BaseHandler.ListObjects(ctx, opts)
}
