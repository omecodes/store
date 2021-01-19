package objects

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	se "github.com/omecodes/store/search-engine"
	"io"
	"strconv"
)

type ParamsHandler struct {
	BaseHandler
}

func (p *ParamsHandler) CreateCollection(ctx context.Context, collection *Collection) error {
	if collection == nil || collection.DefaultAccessSecurityRules == nil || collection.Id == "" {
		return errors.BadInput
	}
	return p.BaseHandler.CreateCollection(ctx, collection)
}

func (p *ParamsHandler) GetCollection(ctx context.Context, id string) (*Collection, error) {
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

func (p *ParamsHandler) PutObject(ctx context.Context, collection string, object *Object, accessSecurityRules *PathAccessRules, indexes []*se.TextIndex, opts PutOptions) (string, error) {
	if collection == "" || object == nil || object.Header == nil || object.Header.Size == 0 {
		return "", errors.BadInput
	}

	settings := Settings(ctx)
	if settings == nil {
		return "", errors.Internal
	}

	s, err := settings.Get(SettingsDataMaxSizePath)
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

func (p *ParamsHandler) PatchObject(ctx context.Context, collection string, patch *Patch, opts PatchOptions) error {
	if collection == "" || patch == nil || patch.ObjectId == "" || len(patch.Data) == 0 || patch.At == "" {
		return errors.BadInput
	}

	settings := Settings(ctx)
	s, err := settings.Get(SettingsDataMaxSizePath)
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

func (p *ParamsHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *PathAccessRules, opts MoveOptions) error {
	if collection == "" || objectID == "" || targetCollection == "" {
		return errors.BadInput
	}
	return p.next.MoveObject(ctx, collection, objectID, targetCollection, accessSecurityRules, opts)
}

func (p *ParamsHandler) GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*Object, error) {
	if collection == "" || id == "" {
		return nil, errors.BadInput
	}
	return p.BaseHandler.GetObject(ctx, collection, id, opts)
}

func (p *ParamsHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*Header, error) {
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

func (p *ParamsHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	if collection == "" {
		return nil, errors.BadInput
	}

	settings := Settings(ctx)
	if settings == nil {
		return nil, errors.Internal
	}

	s, err := settings.Get(SettingsObjectListMaxCount)
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
	limitedBrowser := BrowseFunc(func() (*Object, error) {
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

func (p *ParamsHandler) SearchObjects(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error) {
	if collection == "" || query == nil {
		return nil, errors.BadInput
	}
	return p.BaseHandler.SearchObjects(ctx, collection, query)
}
