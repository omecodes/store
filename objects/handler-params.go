package objects

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/common"
	se "github.com/omecodes/store/search-engine"
	"io"
	"strconv"
)

type ParamsHandler struct {
	BaseHandler
}

func (p *ParamsHandler) CreateCollection(ctx context.Context, collection *Collection) error {
	if collection == nil || collection.DefaultAccessSecurityRules == nil || collection.Id == "" {
		return errors.BadRequest("requires a collection with an ID and default security rules")
	}
	return p.BaseHandler.CreateCollection(ctx, collection)
}

func (p *ParamsHandler) GetCollection(ctx context.Context, id string) (*Collection, error) {
	if id == "" {
		return nil, errors.BadRequest("requires a collection ID")
	}
	return p.BaseHandler.GetCollection(ctx, id)
}

func (p *ParamsHandler) DeleteCollection(ctx context.Context, id string) error {
	if id == "" {
		return errors.BadRequest("requires a collection ID")
	}
	return p.BaseHandler.DeleteCollection(ctx, id)
}

func (p *ParamsHandler) PutObject(ctx context.Context, collection string, object *Object, accessSecurityRules *PathAccessRules, indexes []*se.TextIndex, opts PutOptions) (string, error) {
	if collection == "" || object == nil || len(object.Data) == 0 {
		logs.Error("missing collection or object")
		return "", errors.BadRequest("requires a collection ID and object with header and data")
	}

	settings := common.Settings(ctx)
	if settings == nil {
		return "", errors.Internal("missing settings in context")
	}

	s, err := settings.Get(common.SettingsDataMaxSizePath)
	if err != nil {
		logs.Error("could not get data max length from settings", logs.Err(err))
		return "", errors.Internal("could not get data-max-size config")
	}

	maxLength, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		logs.Error("could not get data max length from settings", logs.Err(err))
		return "", errors.Internal("could not get data-max-size config")
	}

	if object.Header == nil {
		object.Header = new(Header)
	}

	object.Header.Size = int64(len(object.Data))

	if object.Header.Size > maxLength {
		logs.Error("could not process request. Object too big", logs.Details("max", maxLength), logs.Details("received", object.Header.Size))
		return "", errors.BadRequest("object data size exceeds limit")
	}

	return p.next.PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func (p *ParamsHandler) PatchObject(ctx context.Context, collection string, patch *Patch, opts PatchOptions) error {
	if collection == "" || patch == nil || patch.ObjectId == "" || len(patch.Data) == 0 || patch.At == "" {
		return errors.BadRequest("requires a collection ID a patch, with object ID content data and the path")
	}

	settings := common.Settings(ctx)
	if settings == nil {
		return errors.Internal("missing settings in context")
	}

	s, err := settings.Get(common.SettingsDataMaxSizePath)
	if err != nil {
		logs.Error("could not get data max length from settings", logs.Err(err))
		return errors.Internal("could not get data-max-size config")
	}

	maxLength, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		logs.Error("could not get data max length from settings", logs.Err(err))
		return errors.Internal("could not get data-max-size config")
	}

	if int64(len(patch.Data)) > maxLength {
		logs.Error("could not process request. Object too big", logs.Details("max", maxLength), logs.Details("received", len(patch.Data)))
		return errors.BadRequest("data size exceeds limit")
	}

	return p.next.PatchObject(ctx, collection, patch, opts)
}

func (p *ParamsHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *PathAccessRules, opts MoveOptions) error {
	if collection == "" || objectID == "" || targetCollection == "" {
		return errors.BadRequest("requires a collection ID an object ID and the target collection ID")
	}
	return p.next.MoveObject(ctx, collection, objectID, targetCollection, accessSecurityRules, opts)
}

func (p *ParamsHandler) GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*Object, error) {
	if collection == "" || id == "" {
		return nil, errors.BadRequest("requires a collection ID and an object ID")
	}
	return p.BaseHandler.GetObject(ctx, collection, id, opts)
}

func (p *ParamsHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*Header, error) {
	if collection == "" || id == "" {
		return nil, errors.BadRequest("requires a collection ID and an object ID")
	}
	return p.BaseHandler.GetObjectHeader(ctx, collection, id)
}

func (p *ParamsHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	if collection == "" || id == "" {
		return errors.BadRequest("requires a collection ID and an object ID")
	}
	return p.next.DeleteObject(ctx, collection, id)
}

func (p *ParamsHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	if collection == "" {
		return nil, errors.BadRequest("requires a collection ID ")
	}

	settings := common.Settings(ctx)
	if settings == nil {
		return nil, errors.Internal("missing settings in context")
	}

	s, err := settings.Get(common.SettingsObjectListMaxCount)
	if err != nil {
		logs.Error("could not get data max length from settings", logs.Err(err))
		return nil, errors.Internal("could not get data-max-size config")
	}

	maxLength, err := strconv.Atoi(s)
	if err != nil {
		logs.Error("could not get data max length from settings", logs.Err(err))
		return nil, errors.Internal("could not get data-max-size config")
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
		return nil, errors.BadRequest("requires a collection id and a query object")
	}
	return p.BaseHandler.SearchObjects(ctx, collection, query)
}
