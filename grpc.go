package oms

import (
	"context"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/router"
	"io"
)

func NewHandler() pb.HandlerUnitServer {
	return &handler{}
}

type handler struct {
	pb.UnimplementedHandlerUnitServer
}

func (h *handler) PutObject(ctx context.Context, request *pb.PutObjectRequest) (*pb.PutObjectResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	if request.AccessSecurityRules == nil {
		request.AccessSecurityRules = &pb.PathAccessRules{}
	}
	if request.AccessSecurityRules.AccessRules == nil {
		request.AccessSecurityRules.AccessRules = map[string]*pb.AccessRules{}
	}

	id, err := route.PutObject(ctx, request.Object, request.AccessSecurityRules, request.Indexes, pb.PutOptions{})
	if err != nil {
		return nil, err
	}

	return &pb.PutObjectResponse{
		ObjectId: id,
	}, nil
}

func (h *handler) PatchObject(ctx context.Context, request *pb.PatchObjectRequest) (*pb.PatchObjectResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.PatchObjectResponse{}, route.PatchObject(ctx, request.Patch, pb.PatchOptions{})
}

func (h *handler) GetObject(ctx context.Context, request *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	object, err := route.GetObject(ctx, request.ObjectId, pb.GetOptions{
		At:   request.At,
		Info: request.InfoOnly,
	})

	return &pb.GetObjectResponse{
		Object: object,
	}, err
}

func (h *handler) DeleteObject(ctx context.Context, request *pb.DeleteObjectRequest) (*pb.DeleteObjectResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}
	err = route.DeleteObject(ctx, request.ObjectId)
	return &pb.DeleteObjectResponse{}, err
}

func (h *handler) ObjectInfo(ctx context.Context, request *pb.ObjectInfoRequest) (*pb.ObjectInfoResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	header, err := route.GetObjectHeader(ctx, request.ObjectId)
	if err != nil {
		return nil, err
	}

	return &pb.ObjectInfoResponse{Header: header}, nil
}

func (h *handler) ListObjects(request *pb.ListObjectsRequest, stream pb.HandlerUnit_ListObjectsServer) error {
	ctx := stream.Context()

	route, err := router.NewRoute(ctx)
	if err != nil {
		return err
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

	cursor, err := route.ListObjects(ctx, opts)
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
