package router

import (
	"context"
	"github.com/google/uuid"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/objects"
	"github.com/omecodes/store/pb"
)

type ObjectsExecHandler struct {
	ObjectsBaseHandler
}

func (e *ObjectsExecHandler) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	storage := objects.Get(ctx)
	if storage == nil {
		log.Error("exec-handler.CreateCollection: missing storage in context")
		return errors.Internal
	}

	return storage.CreateCollection(ctx, collection)
}

func (e *ObjectsExecHandler) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	storage := objects.Get(ctx)
	if storage == nil {
		log.Error("exec-handler.GetCollection: missing storage in context")
		return nil, errors.Internal
	}

	return storage.GetCollection(ctx, id)
}

func (e *ObjectsExecHandler) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	storage := objects.Get(ctx)
	if storage == nil {
		log.Error("exec-handler.ListCollections: missing storage in context")
		return nil, errors.Internal
	}

	return storage.ListCollections(ctx)
}

func (e *ObjectsExecHandler) DeleteCollection(ctx context.Context, id string) error {
	storage := objects.Get(ctx)
	if storage == nil {
		log.Error("exec-handler.PutObject: missing storage in context")
		return errors.Internal
	}

	return storage.DeleteCollection(ctx, id)
}

func (e *ObjectsExecHandler) PutObject(ctx context.Context, collection string, object *pb.Object, security *pb.PathAccessRules, indexes []*pb.TextIndex, opts pb.PutOptions) (string, error) {
	if object.Header.Id == "" {
		object.Header.Id = uuid.New().String()
	}

	accessStore := acl.GetStore(ctx)
	if accessStore == nil {
		log.Info("exec-handler.PutObject: missing access store in context")
		return "", errors.Internal
	}

	err := accessStore.SaveRules(ctx, collection, object.Header.Id, security)
	if err != nil {
		log.Error("exec-handler.PutObject: failed to save object access security rules", log.Err(err))
		return "", errors.Internal
	}

	storage := objects.Get(ctx)
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

func (e *ObjectsExecHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts pb.PatchOptions) error {
	storage := objects.Get(ctx)
	if storage == nil {
		log.Info("missing storage in context")
		return errors.Internal
	}
	return storage.Patch(ctx, collection, patch)
}

func (e *ObjectsExecHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts pb.MoveOptions) error {
	accessStore := acl.GetStore(ctx)
	if accessStore == nil {
		log.Info("exec-handler.MoveObject: missing access store in context")
		return errors.Internal
	}

	storage := objects.Get(ctx)
	if storage == nil {
		log.Error("exec-handler.PutObject: missing storage in context")
		return errors.Internal
	}

	object, err := storage.Get(ctx, collection, objectID, pb.GetOptions{})
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

func (e *ObjectsExecHandler) GetObject(ctx context.Context, collection string, id string, opts pb.GetOptions) (*pb.Object, error) {
	storage := objects.Get(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.Internal
	}

	return storage.Get(ctx, collection, id, opts)
}

func (e *ObjectsExecHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error) {
	storage := objects.Get(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.Internal
	}
	return storage.Info(ctx, collection, id)
}

func (e *ObjectsExecHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	storage := objects.Get(ctx)
	if storage == nil {
		log.Info("exec-handler.DeleteObjet: missing DB in context")
		return errors.Internal
	}

	err := storage.Delete(ctx, collection, id)
	if err != nil {
		log.Error("exec-handler.DeleteObjet: failed to delete object from storage", log.Err(err))
		return err
	}

	accessStore := acl.GetStore(ctx)
	if accessStore == nil {
		log.Info("exec-handler.DeleteObjet: missing access store in context")
		return errors.Internal
	}

	return accessStore.Delete(ctx, collection, id)
}

func (e *ObjectsExecHandler) ListObjects(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error) {
	storage := objects.Get(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.Internal
	}

	return storage.List(ctx, collection, opts)
}

func (e *ObjectsExecHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*pb.Cursor, error) {
	storage := objects.Get(ctx)
	if storage == nil {
		log.Error("exec-handler.SearchObjects: missing storage in context")
		return nil, errors.Internal
	}

	return storage.Search(ctx, collection, query)
}
