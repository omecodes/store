package objects

import (
	"context"
	"io"

	"github.com/omecodes/libome/logs"
)

func NewGRPCHandler() HandlerUnitServer {
	return &gRPCGatewayHandler{}
}

func NewHandler() HandlerUnitServer {
	return &handler{}
}

type gRPCGatewayHandler struct {
	UnimplementedHandlerUnitServer
}

func (h *gRPCGatewayHandler) CreateCollection(ctx context.Context, request *CreateCollectionRequest) (*CreateCollectionResponse, error) {
	route, err := NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	err = route.CreateCollection(ctx, request.Collection)
	return &CreateCollectionResponse{}, err
}

func (h *gRPCGatewayHandler) GetCollection(ctx context.Context, request *GetCollectionRequest) (*GetCollectionResponse, error) {
	route, err := NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	collection, err := route.GetCollection(ctx, request.Id)
	return &GetCollectionResponse{Collection: collection}, err
}

func (h *gRPCGatewayHandler) ListCollections(ctx context.Context, _ *ListCollectionsRequest) (*ListCollectionsResponse, error) {
	route, err := NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	collections, err := route.ListCollections(ctx)
	return &ListCollectionsResponse{Collections: collections}, err
}

func (h *gRPCGatewayHandler) DeleteCollection(ctx context.Context, request *DeleteCollectionRequest) (*DeleteCollectionResponse, error) {
	route, err := NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	err = route.DeleteCollection(ctx, request.Id)
	return &DeleteCollectionResponse{}, err
}

func (h *gRPCGatewayHandler) PutObject(ctx context.Context, request *PutObjectRequest) (*PutObjectResponse, error) {
	route, err := NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	if request.AccessSecurityRules == nil {
		request.AccessSecurityRules = &PathAccessRules{}
	}
	if request.AccessSecurityRules.AccessRules == nil {
		request.AccessSecurityRules.AccessRules = map[string]*AccessRules{}
	}

	id, err := route.PutObject(ctx, "", request.Object, request.AccessSecurityRules, request.Indexes, PutOptions{})
	if err != nil {
		return nil, err
	}

	return &PutObjectResponse{
		ObjectId: id,
	}, nil
}

func (h *gRPCGatewayHandler) PatchObject(ctx context.Context, request *PatchObjectRequest) (*PatchObjectResponse, error) {
	route, err := NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	return &PatchObjectResponse{}, route.PatchObject(ctx, "", request.Patch, PatchOptions{})
}

func (h *gRPCGatewayHandler) MoveObject(ctx context.Context, request *MoveObjectRequest) (*MoveObjectResponse, error) {
	route, err := NewRoute(ctx)
	if err != nil {
		return nil, err
	}
	return &MoveObjectResponse{}, route.MoveObject(ctx, request.SourceCollection, request.ObjectId, request.TargetCollection, request.AccessSecurityRules, MoveOptions{})
}

func (h *gRPCGatewayHandler) GetObject(ctx context.Context, request *GetObjectRequest) (*GetObjectResponse, error) {
	route, err := NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	object, err := route.GetObject(ctx, "", request.ObjectId, GetOptions{
		At:   request.At,
		Info: request.InfoOnly,
	})

	return &GetObjectResponse{
		Object: object,
	}, err
}

func (h *gRPCGatewayHandler) DeleteObject(ctx context.Context, request *DeleteObjectRequest) (*DeleteObjectResponse, error) {
	route, err := NewRoute(ctx)
	if err != nil {
		return nil, err
	}
	err = route.DeleteObject(ctx, "", request.ObjectId)
	return &DeleteObjectResponse{}, err
}

func (h *gRPCGatewayHandler) ObjectInfo(ctx context.Context, request *ObjectInfoRequest) (*ObjectInfoResponse, error) {
	route, err := NewRoute(ctx)
	if err != nil {
		return nil, err
	}

	header, err := route.GetObjectHeader(ctx, "", request.ObjectId)
	if err != nil {
		return nil, err
	}

	return &ObjectInfoResponse{Header: header}, nil
}

func (h *gRPCGatewayHandler) ListObjects(request *ListObjectsRequest, stream HandlerUnit_ListObjectsServer) error {
	ctx := stream.Context()

	route, err := NewRoute(ctx)
	if err != nil {
		return err
	}

	opts := ListOptions{
		At:     request.At,
		Offset: request.Offset,
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

func (h *gRPCGatewayHandler) SearchObjects(request *SearchObjectsRequest, stream HandlerUnit_SearchObjectsServer) error {
	ctx := stream.Context()

	route, err := NewRoute(ctx)
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
