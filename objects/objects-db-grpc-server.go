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

func (h *handler) PutObject(ctx context.Context, request *pb.PutObjectRequest) (*pb.PutObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	err := storage.Save(ctx, request.Object, request.Indexes...)
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

	err := storage.Patch(ctx, request.Patch)
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

	o, err := storage.Get(ctx, request.ObjectId, pb.GetOptions{
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

	err := storage.Delete(ctx, request.ObjectId)
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

	header, err := storage.Info(ctx, request.ObjectId)
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
		CollectionOptions: pb.CollectionOptions{
			Name:       request.Collection,
			FullObject: request.FullObject,
		},
		Condition: request.Condition,
		At:        request.At,
		DateOptions: pb.DateRangeOptions{
			Before: request.Before,
			After:  request.After,
		},
	}

	cursor, err := storage.List(ctx, opts)
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
