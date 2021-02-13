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
	handler := GetRouterHandler(ctx)
	err := handler.CreateCollection(ctx, request.Collection)
	return &CreateCollectionResponse{}, err
}

func (h *gRPCGatewayHandler) GetCollection(ctx context.Context, request *GetCollectionRequest) (*GetCollectionResponse, error) {
	handler := GetRouterHandler(ctx)
	collection, err := handler.GetCollection(ctx, request.Id)
	return &GetCollectionResponse{Collection: collection}, err
}

func (h *gRPCGatewayHandler) ListCollections(ctx context.Context, _ *ListCollectionsRequest) (*ListCollectionsResponse, error) {
	handler := GetRouterHandler(ctx)

	collections, err := handler.ListCollections(ctx)
	return &ListCollectionsResponse{Collections: collections}, err
}

func (h *gRPCGatewayHandler) DeleteCollection(ctx context.Context, request *DeleteCollectionRequest) (*DeleteCollectionResponse, error) {
	handler := GetRouterHandler(ctx)

	err := handler.DeleteCollection(ctx, request.Id)
	return &DeleteCollectionResponse{}, err
}

func (h *gRPCGatewayHandler) PutObject(ctx context.Context, request *PutObjectRequest) (*PutObjectResponse, error) {
	handler := GetRouterHandler(ctx)

	if request.AccessSecurityRules == nil {
		request.AccessSecurityRules = &PathAccessRules{}
	}
	if request.AccessSecurityRules.AccessRules == nil {
		request.AccessSecurityRules.AccessRules = map[string]*AccessRules{}
	}

	id, err := handler.PutObject(ctx, "", request.Object, request.AccessSecurityRules, request.Indexes, PutOptions{})
	if err != nil {
		return nil, err
	}

	return &PutObjectResponse{
		ObjectId: id,
	}, nil
}

func (h *gRPCGatewayHandler) PatchObject(ctx context.Context, request *PatchObjectRequest) (*PatchObjectResponse, error) {
	handler := GetRouterHandler(ctx)

	return &PatchObjectResponse{}, handler.PatchObject(ctx, "", request.Patch, PatchOptions{})
}

func (h *gRPCGatewayHandler) MoveObject(ctx context.Context, request *MoveObjectRequest) (*MoveObjectResponse, error) {
	handler := GetRouterHandler(ctx)
	return &MoveObjectResponse{}, handler.MoveObject(ctx, request.SourceCollection, request.ObjectId, request.TargetCollection, request.AccessSecurityRules, MoveOptions{})
}

func (h *gRPCGatewayHandler) GetObject(ctx context.Context, request *GetObjectRequest) (*GetObjectResponse, error) {
	handler := GetRouterHandler(ctx)

	object, err := handler.GetObject(ctx, "", request.ObjectId, GetOptions{
		At:   request.At,
		Info: request.InfoOnly,
	})

	return &GetObjectResponse{
		Object: object,
	}, err
}

func (h *gRPCGatewayHandler) DeleteObject(ctx context.Context, request *DeleteObjectRequest) (*DeleteObjectResponse, error) {
	handler := GetRouterHandler(ctx)
	err := handler.DeleteObject(ctx, "", request.ObjectId)
	return &DeleteObjectResponse{}, err
}

func (h *gRPCGatewayHandler) ObjectInfo(ctx context.Context, request *ObjectInfoRequest) (*ObjectInfoResponse, error) {
	handler := GetRouterHandler(ctx)

	header, err := handler.GetObjectHeader(ctx, "", request.ObjectId)
	if err != nil {
		return nil, err
	}

	return &ObjectInfoResponse{Header: header}, nil
}

func (h *gRPCGatewayHandler) ListObjects(request *ListObjectsRequest, stream HandlerUnit_ListObjectsServer) error {
	ctx := stream.Context()

	handler := GetRouterHandler(ctx)

	opts := ListOptions{
		At:     request.At,
		Offset: request.Offset,
	}

	cursor, err := handler.ListObjects(ctx, request.Collection, opts)
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

	handler := GetRouterHandler(ctx)

	cursor, err := handler.SearchObjects(ctx, request.Collection, request.Query)
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
