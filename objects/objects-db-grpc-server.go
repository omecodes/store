package objects

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/pb"
	"io"
)

func NewStoreGrpcHandler() pb.HandlerUnitServer {
	return &handler{}
}

type handler struct {
	pb.UnimplementedHandlerUnitServer
}

func (h *handler) CreateCollection(ctx context.Context, request *pb.CreateCollectionRequest) (*pb.CreateCollectionResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	err := storage.CreateCollection(ctx, request.Collection)
	return &pb.CreateCollectionResponse{}, err
}

func (h *handler) GetCollection(ctx context.Context, request *pb.GetCollectionRequest) (*pb.GetCollectionResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	collection, err := storage.GetCollection(ctx, request.Id)
	return &pb.GetCollectionResponse{Collection: collection}, err
}

func (h *handler) ListCollections(ctx context.Context, request *pb.ListCollectionsRequest) (*pb.ListCollectionsResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	collections, err := storage.ListCollections(ctx)
	return &pb.ListCollectionsResponse{Collections: collections}, err
}

func (h *handler) DeleteCollection(ctx context.Context, request *pb.DeleteCollectionRequest) (*pb.DeleteCollectionResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	err := storage.DeleteCollection(ctx, request.Id)
	return &pb.DeleteCollectionResponse{}, err
}

func (h *handler) PutObject(ctx context.Context, request *pb.PutObjectRequest) (*pb.PutObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	err := storage.Save(ctx, request.Collection, request.Object, request.Indexes...)
	if err != nil {
		return nil, err
	}

	return &pb.PutObjectResponse{}, nil
}

func (h *handler) PatchObject(ctx context.Context, request *pb.PatchObjectRequest) (*pb.PatchObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	err := storage.Patch(ctx, request.Collection, request.Patch)
	if err != nil {
		return nil, err
	}

	return &pb.PatchObjectResponse{}, nil
}

func (h *handler) GetObject(ctx context.Context, request *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	o, err := storage.Get(ctx, request.Collection, request.ObjectId, pb.GetOptions{
		At:   request.At,
		Info: request.InfoOnly,
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetObjectResponse{
		Object: o}, err
}

func (h *handler) DeleteObject(ctx context.Context, request *pb.DeleteObjectRequest) (*pb.DeleteObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	err := storage.Delete(ctx, request.Collection, request.ObjectId)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteObjectResponse{}, nil
}

func (h *handler) ObjectInfo(ctx context.Context, request *pb.ObjectInfoRequest) (*pb.ObjectInfoResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	header, err := storage.Info(ctx, request.Collection, request.ObjectId)
	if err != nil {
		return nil, err
	}

	return &pb.ObjectInfoResponse{Header: header}, nil
}

func (h *handler) ListObjects(request *pb.ListObjectsRequest, stream pb.HandlerUnit_ListObjectsServer) error {
	ctx := stream.Context()

	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return errors.Internal
	}

	opts := pb.ListOptions{
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

func (h *handler) SearchObjects(request *pb.SearchObjectsRequest, stream pb.HandlerUnit_SearchObjectsServer) error {
	ctx := stream.Context()

	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return errors.Internal
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
