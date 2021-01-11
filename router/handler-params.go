package router

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/objects"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/utime"
	"io"
	"strconv"
)

type ParamsHandler struct {
	BaseHandler
}

func (p *ParamsHandler) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	if collection == nil || collection.DefaultAccessSecurityRules == nil || collection.Id == "" {
		return errors.BadInput
	}
	return p.BaseHandler.CreateCollection(ctx, collection)
}

func (p *ParamsHandler) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	if id == "" {
		return nil, errors.BadInput
	}
	return p.BaseHandler.GetCollection(ctx, id)
}

func (p *ParamsHandler) DeleteCollection(ctx context.Context, id string) error {
	if id == "" {
		return errors.BadInput
	}
	return p.BaseHandler.DeleteCollection(ctx, id)
}

func (p *ParamsHandler) PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.Index, opts pb.PutOptions) (string, error) {
	if collection == "" || object == nil || object.Header == nil || object.Header.Size == 0 {
		return "", errors.BadInput
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

	return p.next.PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func (p *ParamsHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts pb.PatchOptions) error {
	if collection == "" || patch == nil || patch.ObjectId == "" || len(patch.Data) == 0 || patch.At == "" {
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

	return p.next.PatchObject(ctx, collection, patch, opts)
}

func (p *ParamsHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts pb.MoveOptions) error {
	if collection == "" || objectID == "" || targetCollection == "" {
		return errors.BadInput
	}
	return p.next.MoveObject(ctx, collection, objectID, targetCollection, accessSecurityRules, opts)
}

func (p *ParamsHandler) GetObject(ctx context.Context, collection string, id string, opts pb.GetOptions) (*pb.Object, error) {
	if collection == "" || id == "" {
		return nil, errors.BadInput
	}
	return p.BaseHandler.GetObject(ctx, collection, id, opts)
}

func (p *ParamsHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error) {
	if collection == "" || id == "" {
		return nil, errors.BadInput
	}
	return p.BaseHandler.GetObjectHeader(ctx, collection, id)
}

func (p *ParamsHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	if collection == "" || id == "" {
		return errors.BadInput
	}
	return p.next.DeleteObject(ctx, collection, id)
}

func (p *ParamsHandler) ListObjects(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error) {
	if collection == "" {
		return nil, errors.BadInput
	}

	if opts.DateOptions.Before == 0 {
		opts.DateOptions.Before = utime.Now()
	}

	if opts.DateOptions.After == 0 {
		opts.DateOptions.After = 1
	}

	settings := Settings(ctx)
	if settings == nil {
		return nil, errors.Internal
	}

	s, err := settings.Get(objects.SettingsObjectListMaxCount)
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return nil, errors.Internal
	}

	maxLength, err := strconv.Atoi(s)
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return nil, errors.Internal
	}

	cursor, err := p.BaseHandler.ListObjects(ctx, collection, opts)
	if err != nil {
		return nil, err
	}

	browser := cursor.GetBrowser()
	count := 0
	limitedBrowser := pb.BrowseFunc(func() (*pb.Object, error) {
		if count == maxLength {
			return nil, io.EOF
		}
		o, err := browser.Browse()
		if err == nil {
			count++
		}
		return o, err
	})

	cursor.SetBrowser(limitedBrowser)
	return cursor, nil
}

func (p *ParamsHandler) SearchObjects(ctx context.Context, collection string, exp *pb.BooleanExp) (*pb.Cursor, error) {
	if collection == "" || exp == nil {
		return nil, errors.BadInput
	}
	return p.BaseHandler.SearchObjects(ctx, collection, exp)
}
