package objects

import (
	"bytes"
	"context"
	"fmt"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/omestore/router"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
)

func NewHandler() pb.HandlerUnitServer {
	return &handler{}
}

type handler struct {
	pb.UnimplementedHandlerUnitServer
}

func (h *handler) Put(ctx context.Context, request *pb.PutObjectRequest) (*pb.PutObjectResponse, error) {
	handler := router.NewRoute(ctx, router.SkipParamsCheck(), router.SkipPoliciesCheck())
	if handler == nil {
		return nil, errors.Internal
	}

	object := oms.NewObject()
	object.SetHeader(request.Header)
	object.SetContent(bytes.NewBuffer(request.Data))

	id, err := handler.PutObject(ctx, object, request.AccessSecurityRules, oms.PutDataOptions{})
	if err != nil {
		return nil, err
	}

	return &pb.PutObjectResponse{
		ObjectId: id,
	}, nil
}

func (h *handler) Update(ctx context.Context, request *pb.UpdateObjectRequest) (*pb.UpdateObjectResponse, error) {
	handler := router.NewRoute(ctx, router.SkipParamsCheck(), router.SkipPoliciesCheck())
	if handler == nil {
		return nil, errors.Internal
	}

	patch := oms.NewPatch(request.ObjectId, request.Path)
	patch.SetContent(bytes.NewBuffer(request.Data))

	return &pb.UpdateObjectResponse{}, handler.PatchObject(ctx, patch, oms.PatchOptions{})
}

func (h *handler) Get(ctx context.Context, request *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	handler := router.NewRoute(ctx, router.SkipParamsCheck(), router.SkipPoliciesCheck())
	if handler == nil {
		return nil, errors.Internal
	}

	object, err := handler.GetObject(ctx, "", oms.GetObjectOptions{})
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

func (h *handler) Delete(ctx context.Context, request *pb.DeleteObjectRequest) (*pb.DeleteObjectResponse, error) {
	handler := router.NewRoute(ctx, router.SkipParamsCheck(), router.SkipPoliciesCheck())
	if handler == nil {
		return nil, errors.Internal
	}
	return &pb.DeleteObjectResponse{}, handler.DeleteObject(ctx, request.ObjectId)
}

func (h *handler) Info(ctx context.Context, request *pb.ObjectInfoRequest) (*pb.ObjectInfoResponse, error) {
	handler := router.NewRoute(ctx, router.SkipParamsCheck(), router.SkipPoliciesCheck())
	if handler == nil {
		return nil, errors.Internal
	}

	header, err := handler.GetObjectHeader(ctx, request.ObjectId)
	if err != nil {
		return nil, err
	}

	return &pb.ObjectInfoResponse{Header: header}, nil
}

func (h *handler) List(request *pb.ListObjectsRequest, stream pb.HandlerUnit_ListObjectsServer) error {
	ctx := stream.Context()

	handler := router.NewRoute(ctx, router.SkipParamsCheck(), router.SkipPoliciesCheck())
	if handler == nil {
		return errors.Internal
	}

	opts := oms.ListOptions{
		Before: request.Before,
		Count:  int(request.Count),
	}

	items, err := handler.ListObjects(ctx, opts)
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

func (h *handler) Search(request *pb.SearchObjectsRequest, stream pb.HandlerUnit_SearchObjectsServer) error {
	ctx := stream.Context()

	handler := router.NewRoute(ctx, router.SkipParamsCheck(), router.SkipPoliciesCheck())
	if handler == nil {
		return errors.Internal
	}

	params := oms.SearchParams{
		MatchedExpression: request.MatchExpression,
	}

	opts := oms.SearchOptions{
		Before: request.Before,
		Count:  int(request.Count),
	}

	items, err := handler.SearchObjects(ctx, params, opts)
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
