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

func (h *handler) CreateCollection(ctx context.Context, request *pb.CreateCollectionRequest) (*pb.CreateCollectionResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	err = route.CreateCollection(ctx, request.Collection)
	return &pb.CreateCollectionResponse{}, err
}

func (h *handler) GetCollection(ctx context.Context, request *pb.GetCollectionRequest) (*pb.GetCollectionResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	collection, err := route.GetCollection(ctx, request.Id)
	return &pb.GetCollectionResponse{Collection: collection}, err
}

func (h *handler) ListCollections(ctx context.Context, request *pb.ListCollectionsRequest) (*pb.ListCollectionsResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	collections, err := route.ListCollections(ctx)
	return &pb.ListCollectionsResponse{Collections: collections}, err
}

func (h *handler) DeleteCollection(ctx context.Context, request *pb.DeleteCollectionRequest) (*pb.DeleteCollectionResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	err = route.DeleteCollection(ctx, request.Id)
	return &pb.DeleteCollectionResponse{}, err
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

	id, err := route.PutObject(ctx, "", request.Object, request.AccessSecurityRules, request.Indexes, pb.PutOptions{})
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

	return &pb.PatchObjectResponse{}, route.PatchObject(ctx, "", request.Patch, pb.PatchOptions{})
}

func (h *handler) MoveObject(ctx context.Context, request *pb.MoveObjectRequest) (*pb.MoveObjectResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.MoveObjectResponse{}, route.MoveObject(ctx, request.SourceCollection, request.ObjectId, request.TargetCollection, request.AccessSecurityRules, pb.MoveOptions{})
}

func (h *handler) GetObject(ctx context.Context, request *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	object, err := route.GetObject(ctx, "", request.ObjectId, pb.GetOptions{
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
	err = route.DeleteObject(ctx, "", request.ObjectId)
	return &pb.DeleteObjectResponse{}, err
}

func (h *handler) ObjectInfo(ctx context.Context, request *pb.ObjectInfoRequest) (*pb.ObjectInfoResponse, error) {
	route, err := router.NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	header, err := route.GetObjectHeader(ctx, "", request.ObjectId)
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
		At: request.At,
		DateOptions: pb.DateRangeOptions{
			Before: request.Before,
			After:  request.After,
		},
	}

	cursor, err := route.ListObjects(ctx, request.Collection, opts)
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

	route, err := router.NewRoute(ctx)
	if err != nil {
		return err
	}

	cursor, err := route.SearchObjects(ctx, request.Collection, request.Query)
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
