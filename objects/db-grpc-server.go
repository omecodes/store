package objects

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"io"
)

func NewStoreGrpcHandler() ObjectsServer {
	return &handler{}
}

type handler struct {
	UnimplementedObjectsServer
}

func (h *handler) CreateCollection(ctx context.Context, request *CreateCollectionRequest) (*CreateCollectionResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	err := storage.CreateCollection(ctx, request.Collection)
	return &CreateCollectionResponse{}, err
}

func (h *handler) GetCollection(ctx context.Context, request *GetCollectionRequest) (*GetCollectionResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	collection, err := storage.GetCollection(ctx, request.Id)
	return &GetCollectionResponse{Collection: collection}, err
}

func (h *handler) ListCollections(ctx context.Context, request *ListCollectionsRequest) (*ListCollectionsResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	collections, err := storage.ListCollections(ctx)
	return &ListCollectionsResponse{Collections: collections}, err
}

func (h *handler) DeleteCollection(ctx context.Context, request *DeleteCollectionRequest) (*DeleteCollectionResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	err := storage.DeleteCollection(ctx, request.Id)
	return &DeleteCollectionResponse{}, err
}

func (h *handler) PutObject(ctx context.Context, request *PutObjectRequest) (*PutObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	err := storage.Save(ctx, request.Collection, request.Object, request.Indexes...)
	if err != nil {
		return nil, err
	}

	return &PutObjectResponse{}, nil
}

func (h *handler) PatchObject(ctx context.Context, request *PatchObjectRequest) (*PatchObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	err := storage.Patch(ctx, request.Collection, request.Patch)
	if err != nil {
		return nil, err
	}

	return &PatchObjectResponse{}, nil
}

func (h *handler) GetObject(ctx context.Context, request *GetObjectRequest) (*GetObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	o, err := storage.Get(ctx, request.Collection, request.ObjectId, GetOptions{
		At:   request.At,
		Info: request.InfoOnly,
	})
	if err != nil {
		return nil, err
	}

	return &GetObjectResponse{
		Object: o}, err
}

func (h *handler) DeleteObject(ctx context.Context, request *DeleteObjectRequest) (*DeleteObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal("missing object storage")
	}

	err := storage.Delete(ctx, request.Collection, request.ObjectId)
	if err != nil {
		return nil, err
	}
	return &DeleteObjectResponse{}, nil
}

func (h *handler) ObjectInfo(ctx context.Context, request *ObjectInfoRequest) (*ObjectInfoResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal("missing objects storage")
	}

	header, err := storage.Info(ctx, request.Collection, request.ObjectId)
	if err != nil {
		return nil, err
	}

	return &ObjectInfoResponse{Header: header}, nil
}

func (h *handler) ListObjects(request *ListObjectsRequest, stream Objects_ListObjectsServer) error {
	ctx := stream.Context()

	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return errors.Internal("missing objects storage")
	}

	opts := ListOptions{
		At:     request.At,
		Offset: request.Offset,
	}

	cursor, err := storage.List(ctx, request.Collection, opts)
	if err != nil {
		return err
	}

	defer func() {
		if ce := cursor.Close(); ce != nil {
			logs.Error("closed cursor with error", logs.Err(err))
		}
	}()

	for {
		o, err := cursor.Browse()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		err = stream.Send(o)
		if err != nil {
			return err
		}
	}
}

func (h *handler) SearchObjects(request *SearchObjectsRequest, stream Objects_SearchObjectsServer) error {
	ctx := stream.Context()

	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return errors.Internal("missing objects storage")
	}

	cursor, err := storage.Search(ctx, request.Collection, request.Query)
	if err != nil {
		return err
	}

	defer func() {
		if ce := cursor.Close(); ce != nil {
			logs.Error("closed cursor with error", logs.Err(err))
		}
	}()

	for {
		o, err := cursor.Browse()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		err = stream.Send(o)
		if err != nil {
			return err
		}
	}
}
