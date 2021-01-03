package oms

import (
	"bytes"
	"context"
	"fmt"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/pb"
	"google.golang.org/grpc/metadata"
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

func (h *handler) UpdateObject(ctx context.Context, request *pb.UpdateObjectRequest) (*pb.UpdateObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	patch := NewPatch(request.ObjectId, request.Path)
	patch.SetContent(bytes.NewBufferString(request.Data))

	err := storage.Patch(ctx, patch)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateObjectResponse{}, nil
}

func (h *handler) GetObject(ctx context.Context, request *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	o, err := storage.Get(ctx, request.ObjectId)
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

	opts := ListOptions{
		Path:   request.Path,
		Before: request.Before,
		After:  request.After,
		Count:  int(request.Count),
	}

	items, err := storage.List(ctx, nil, opts)
	if err != nil {
		return err
	}

	md := metadata.MD{}
	md.Set("before", fmt.Sprintf("%d", items.Before))
	md.Set("count", fmt.Sprintf("%d", len(items.Objects)))
	err = stream.SetHeader(md)
	if err != nil {
		return err
	}

	for _, object := range items.Objects {
		err = stream.Send(object)
		if err != nil {
			return err
		}
	}
	return nil
}
