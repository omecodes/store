package oms

import (
	"bytes"
	"context"
	"fmt"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/pb"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
)

func NewStoreGrpcHandler() pb.HandlerUnitServer {
	return &handler{}
}

type handler struct {
	pb.UnimplementedHandlerUnitServer
}

func (h *handler) PutObject(ctx context.Context, request *pb.PutObjectRequest) (*pb.PutObjectResponse, error) {
	object := NewObject()
	object.SetHeader(request.Header)
	object.SetContent(bytes.NewBufferString(request.Data))

	storage := Get(ctx)
	if storage == nil {
		logs.Info("Objects server • missing storage in context")
		return nil, errors.Internal
	}

	err := storage.Save(ctx, object, request.Indexes...)
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

	data, err := ioutil.ReadAll(o.GetContent())
	if err != nil {
		return nil, err
	}

	return &pb.GetObjectResponse{
		Data: &pb.DataObject{
			Header: o.Header(),
			Data:   data,
		}}, err
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

	items, err := storage.List(ctx, request.Before, int(request.Count), nil)
	if err != nil {
		return err
	}

	md := metadata.MD{}
	md.Set("before", fmt.Sprintf("%d", items.Before))
	md.Set("count", fmt.Sprintf("%d", items.Count))
	err = stream.SetHeader(md)
	if err != nil {
		return err
	}

	for _, object := range items.Objects {
		data, err := ioutil.ReadAll(object.GetContent())
		if err != nil {
			return err
		}

		err = stream.Send(&pb.DataObject{
			Header: object.Header(),
			Data:   data,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
