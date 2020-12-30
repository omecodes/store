package oms

import (
	"bytes"
	"context"
	"fmt"
	"github.com/omecodes/store/oms"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/router"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
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

	object := oms.NewObject()
	object.SetHeader(request.Header)
	object.SetContent(bytes.NewBufferString(request.Data))

	if request.AccessSecurityRules == nil {
		request.AccessSecurityRules = &pb.PathAccessRules{}
	}
	if request.AccessSecurityRules.AccessRules == nil {
		request.AccessSecurityRules.AccessRules = map[string]*pb.AccessRules{}
	}

	id, err := route.PutObject(ctx, object, request.AccessSecurityRules, oms.PutDataOptions{
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

	patch := oms.NewPatch(request.ObjectId, request.Path)
	patch.SetContent(bytes.NewBuffer(request.Data))

	return &pb.UpdateObjectResponse{}, route.PatchObject(ctx, patch, oms.PatchOptions{})
}

func (h *handler) GetObject(ctx context.Context, request *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	object, err := route.GetObject(ctx, "", oms.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(object.GetContent())
	if err != nil {
		return nil, err
	}

	rsp := &pb.GetObjectResponse{
		Data: &pb.DataObject{
			Header: object.Header(),
			Data:   data,
		},
	}
	return rsp, nil
}

func (h *handler) DeleteObject(ctx context.Context, request *pb.DeleteObjectRequest) (*pb.DeleteObjectResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteObjectResponse{}, route.DeleteObject(ctx, request.ObjectId)
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

	opts := oms.ListOptions{
		Before: request.Before,
		Count:  int(request.Count),
	}

	items, err := route.ListObjects(ctx, opts)
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

func (h *handler) SearchObjects(request *pb.SearchObjectsRequest, stream pb.HandlerUnit_SearchObjectsServer) error {
	ctx := stream.Context()

	route, err := router.NewRoute(ctx)
	if err != nil {
		return err
	}

	params := oms.SearchParams{
		MatchedExpression: request.MatchExpression,
	}

	opts := oms.SearchOptions{
		Before: request.Before,
		Count:  int(request.Count),
	}

	items, err := route.SearchObjects(ctx, params, opts)
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
