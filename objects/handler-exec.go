package objects

import (
	"context"
	"github.com/google/uuid"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	pb "github.com/omecodes/store/gen/go/proto"
)

type ExecHandler struct {
	BaseHandler
}

func (e *ExecHandler) CreateCollection(ctx context.Context, collection *pb.Collection, _ CreateCollectionOptions) error {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.CreateCollection: missing storage in context")
		return errors.Internal("missing objects storage")
	}

	return storage.CreateCollection(ctx, collection)
}

func (e *ExecHandler) GetCollection(ctx context.Context, id string, _ GetCollectionOptions) (*pb.Collection, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.GetCollection: missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	return storage.GetCollection(ctx, id)
}

func (e *ExecHandler) ListCollections(ctx context.Context, _ ListCollectionOptions) ([]*pb.Collection, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.ListCollections: missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	return storage.ListCollections(ctx)
}

func (e *ExecHandler) DeleteCollection(ctx context.Context, id string, _ DeleteCollectionOptions) error {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.PutObject: missing storage in context")
		return errors.Internal("missing objects storage")
	}

	return storage.DeleteCollection(ctx, id)
}

func (e *ExecHandler) PutObject(ctx context.Context, collection string, object *pb.Object, _ *pb.PathAccessRules, indexes []*pb.TextIndex, _ PutOptions) (string, error) {
	if object.Header.Id == "" {
		object.Header.Id = uuid.New().String()
	}

	storage := Get(ctx)
	if storage == nil {
		return "", errors.Internal("missing objects storage")
	}

	err := storage.Save(ctx, collection, object, indexes...)
	if err != nil {
		logs.Error("could not save object", logs.Err(err))
		return "", err
	}

	return object.Header.Id, nil
}

func (e *ExecHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, _ PatchOptions) error {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("missing storage in context")
		return errors.Internal("missing objects storage")
	}
	return storage.Patch(ctx, collection, patch)
}

func (e *ExecHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, _ *pb.PathAccessRules, _ MoveOptions) error {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.PutObject: missing storage in context")
		return errors.Internal("missing objects storage")
	}

	object, err := storage.Get(ctx, collection, objectID, GetObjectOptions{})
	if err != nil {
		return err
	}

	return storage.Save(ctx, targetCollection, object)
}

func (e *ExecHandler) GetObject(ctx context.Context, collection string, id string, opts GetObjectOptions) (*pb.Object, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("missing DB in context")
		return nil, errors.Internal("missing objects storage")
	}

	return storage.Get(ctx, collection, id, opts)
}

func (e *ExecHandler) GetObjectHeader(ctx context.Context, collection string, id string, _ GetHeaderOptions) (*pb.Header, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("missing DB in context")
		return nil, errors.Internal("missing objects storage")
	}
	return storage.Info(ctx, collection, id)
}

func (e *ExecHandler) DeleteObject(ctx context.Context, collection string, id string, _ DeleteObjectOptions) error {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("exec-handler.DeleteObjet: missing DB in context")
		return errors.Internal("missing objects storage")
	}

	return storage.Delete(ctx, collection, id)
}

func (e *ExecHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("missing DB in context")
		return nil, errors.Internal("missing objects storage")
	}

	return storage.List(ctx, collection, opts)
}

func (e *ExecHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery, _ SearchObjectsOptions) (*Cursor, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Error("exec-handler.SearchObjects: missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	return storage.Search(ctx, collection, query)
}
