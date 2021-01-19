package objects

import (
	"context"
	"github.com/google/uuid"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	se "github.com/omecodes/store/search-engine"
)

type ExecHandler struct {
	BaseHandler
}

func (e *ExecHandler) CreateCollection(ctx context.Context, collection *Collection) error {
	storage := Get(ctx)
	if storage == nil {
		log.Error("exec-handler.CreateCollection: missing storage in context")
		return errors.Internal
	}

	return storage.CreateCollection(ctx, collection)
}

func (e *ExecHandler) GetCollection(ctx context.Context, id string) (*Collection, error) {
	storage := Get(ctx)
	if storage == nil {
		log.Error("exec-handler.GetCollection: missing storage in context")
		return nil, errors.Internal
	}

	return storage.GetCollection(ctx, id)
}

func (e *ExecHandler) ListCollections(ctx context.Context) ([]*Collection, error) {
	storage := Get(ctx)
	if storage == nil {
		log.Error("exec-handler.ListCollections: missing storage in context")
		return nil, errors.Internal
	}

	return storage.ListCollections(ctx)
}

func (e *ExecHandler) DeleteCollection(ctx context.Context, id string) error {
	storage := Get(ctx)
	if storage == nil {
		log.Error("exec-handler.PutObject: missing storage in context")
		return errors.Internal
	}

	return storage.DeleteCollection(ctx, id)
}

func (e *ExecHandler) PutObject(ctx context.Context, collection string, object *Object, security *PathAccessRules, indexes []*se.TextIndex, opts PutOptions) (string, error) {
	if object.Header.Id == "" {
		object.Header.Id = uuid.New().String()
	}

	accessStore := GetACLStore(ctx)
	if accessStore == nil {
		log.Info("exec-handler.PutObject: missing access store in context")
		return "", errors.Internal
	}

	err := accessStore.SaveRules(ctx, collection, object.Header.Id, security)
	if err != nil {
		log.Error("exec-handler.PutObject: failed to save object access security rules", log.Err(err))
		return "", errors.Internal
	}

	storage := Get(ctx)
	if storage == nil {
		log.Error("exec-handler.PutObject: missing storage in context")
		if err2 := accessStore.Delete(ctx, collection, object.Header.Id); err2 != nil {
			log.Error("exec-handler.PutObject: failed to clear access rules", log.Err(err2))
		}
		return "", errors.Internal
	}

	err = storage.Save(ctx, collection, object, indexes...)
	if err != nil {
		if err2 := accessStore.Delete(ctx, collection, object.Header.Id); err2 != nil {
			log.Error("exec-handler.PutObject: failed to clear access rules", log.Err(err2))
		}
		return "", err
	}

	return object.Header.Id, nil
}

func (e *ExecHandler) PatchObject(ctx context.Context, collection string, patch *Patch, opts PatchOptions) error {
	storage := Get(ctx)
	if storage == nil {
		log.Info("missing storage in context")
		return errors.Internal
	}
	return storage.Patch(ctx, collection, patch)
}

func (e *ExecHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *PathAccessRules, opts MoveOptions) error {
	accessStore := GetACLStore(ctx)
	if accessStore == nil {
		log.Info("exec-handler.MoveObject: missing access store in context")
		return errors.Internal
	}

	storage := Get(ctx)
	if storage == nil {
		log.Error("exec-handler.PutObject: missing storage in context")
		return errors.Internal
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
		log.Info("missing DB in context")
		return nil, errors.Internal
	}

	return storage.Get(ctx, collection, id, opts)
}

func (e *ExecHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*Header, error) {
	storage := Get(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.Internal
	}
	return storage.Info(ctx, collection, id)
}

func (e *ExecHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	storage := Get(ctx)
	if storage == nil {
		log.Info("exec-handler.DeleteObjet: missing DB in context")
		return errors.Internal
	}

	err := storage.Delete(ctx, collection, id)
	if err != nil {
		log.Error("exec-handler.DeleteObjet: failed to delete object from storage", log.Err(err))
		return err
	}

	accessStore := GetACLStore(ctx)
	if accessStore == nil {
		log.Info("exec-handler.DeleteObjet: missing access store in context")
		return errors.Internal
	}

	return accessStore.Delete(ctx, collection, id)
}

func (e *ExecHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	storage := Get(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.Internal
	}

	return storage.List(ctx, collection, opts)
}

func (e *ExecHandler) SearchObjects(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error) {
	storage := Get(ctx)
	if storage == nil {
		log.Error("exec-handler.SearchObjects: missing storage in context")
		return nil, errors.Internal
	}

	return storage.Search(ctx, collection, query)
}
