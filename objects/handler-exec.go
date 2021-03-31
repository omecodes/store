package objects

import (
	"context"
	"github.com/google/uuid"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	se "github.com/omecodes/store/search-engine"
)

type ExecHandler struct {
	BaseHandler
}

func (e *ExecHandler) CreateCollection(ctx context.Context, collection *Collection) error {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.CreateCollection: missing storage in context")
		return errors.Internal("missing objects storage")
	}

	return storage.CreateCollection(ctx, collection)
}

func (e *ExecHandler) GetCollection(ctx context.Context, id string) (*Collection, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.GetCollection: missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	return storage.GetCollection(ctx, id)
}

func (e *ExecHandler) ListCollections(ctx context.Context) ([]*Collection, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.ListCollections: missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	return storage.ListCollections(ctx)
}

func (e *ExecHandler) DeleteCollection(ctx context.Context, id string) error {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.PutObject: missing storage in context")
		return errors.Internal("missing objects storage")
	}

	return storage.DeleteCollection(ctx, id)
}

func (e *ExecHandler) PutObject(ctx context.Context, collection string, object *Object, security *PathAccessRules, indexes []*se.TextIndex, opts PutOptions) (string, error) {
	if object.Header.Id == "" {
		object.Header.Id = uuid.New().String()
	}

	accessStore := GetACLStore(ctx)
	if accessStore == nil {
		logs.Info("exec-handler.PutObject: missing access store in context")
		return "", errors.Internal("missing objects storage")
	}

	err := accessStore.SaveRules(ctx, collection, object.Header.Id, security)
	if err != nil {
		logs.Error("exec-handler.PutObject: failed to save object access security rules", logs.Err(err))
		return "", errors.Internal("missing objects storage")
	}

	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.PutObject: missing storage in context")
		if err2 := accessStore.Delete(ctx, collection, object.Header.Id); err2 != nil {
			logs.Error("exec-handler.PutObject: failed to clear access rules", logs.Err(err2))
		}
		return "", errors.Internal("missing objects storage")
	}

	err = storage.Save(ctx, collection, object, indexes...)
	if err != nil {
		logs.Error("could not save object", logs.Err(err))
		if err2 := accessStore.Delete(ctx, collection, object.Header.Id); err2 != nil {
			logs.Error("exec-handler.PutObject: failed to clear access rules", logs.Err(err2))
		}
		return "", err
	}

	return object.Header.Id, nil
}

func (e *ExecHandler) PatchObject(ctx context.Context, collection string, patch *Patch, opts PatchOptions) error {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("missing storage in context")
		return errors.Internal("missing objects storage")
	}
	return storage.Patch(ctx, collection, patch)
}

func (e *ExecHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *PathAccessRules, opts MoveOptions) error {
	accessStore := GetACLStore(ctx)
	if accessStore == nil {
		logs.Info("exec-handler.MoveObject: missing access store in context")
		return errors.Internal("missing ACL store")
	}

	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.PutObject: missing storage in context")
		return errors.Internal("missing objects storage")
	}

	object, err := storage.Get(ctx, collection, objectID, GetOptions{})
	if err != nil {
		return err
	}

	err = storage.Save(ctx, targetCollection, object)
	if err != nil {
		return err
	}

	err = accessStore.SaveRules(ctx, targetCollection, objectID, accessSecurityRules)
	if err != nil {
		return err
	}

	return accessStore.Delete(ctx, collection, objectID)
}

func (e *ExecHandler) GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*Object, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("missing DB in context")
		return nil, errors.Internal("missing objects storage")
	}

	return storage.Get(ctx, collection, id, opts)
}

func (e *ExecHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*Header, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("missing DB in context")
		return nil, errors.Internal("missing objects storage")
	}
	return storage.Info(ctx, collection, id)
}

func (e *ExecHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("exec-handler.DeleteObjet: missing DB in context")
		return errors.Internal("missing objects storage")
	}

	err := storage.Delete(ctx, collection, id)
	if err != nil {
		logs.Error("exec-handler.DeleteObjet: failed to delete object from storage", logs.Err(err))
		return err
	}

	accessStore := GetACLStore(ctx)
	if accessStore == nil {
		logs.Info("exec-handler.DeleteObjet: missing access store in context")
		return errors.Internal("missing ACL store")
	}

	return accessStore.Delete(ctx, collection, id)
}

func (e *ExecHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("missing DB in context")
		return nil, errors.Internal("missing objects storage")
	}

	return storage.List(ctx, collection, opts)
}

func (e *ExecHandler) SearchObjects(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.SearchObjects: missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	return storage.Search(ctx, collection, query)
}
