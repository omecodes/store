package oms

import (
	"context"
	"fmt"
	"github.com/omecodes/store/objects"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/router"
	"google.golang.org/grpc/metadata"
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

	id, err := route.PutObject(ctx, request.Object, request.AccessSecurityRules, objects.PutDataOptions{
		Indexes: request.Indexes,
	})
	if err != nil {
		return nil, err
	}

	return &pb.PutObjectResponse{
		ObjectId: id,
	}, nil
}

func (h *handler) UpdateObject(ctx context.Context, request *pb.UpdateObjectRequest) (*pb.UpdateObjectResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateObjectResponse{}, route.PatchObject(ctx, request.Patch, objects.PatchOptions{})
}

func (h *handler) GetObject(ctx context.Context, request *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	object, err := route.GetObject(ctx, request.ObjectId, objects.GetObjectOptions{Path: request.Path})

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

	opts := objects.ListOptions{
		Path:   request.Path,
		Before: request.Before,
		After:  request.After,
		Count:  int(request.Count),
	}

	items, err := route.ListObjects(ctx, opts)
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

func (h *handler) SearchObjects(request *pb.SearchObjectsRequest, stream pb.HandlerUnit_SearchObjectsServer) error {
	ctx := stream.Context()

	route, err := router.NewRoute(ctx)
	if err != nil {
		return err
	}

	params := objects.SearchParams{
		Condition: request.Condition,
	}

	opts := objects.SearchOptions{
		Before: request.Before,
		Count:  int(request.Count),
	}

	items, err := route.SearchObjects(ctx, params, opts)
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
