package objects

import (
	"context"
	"github.com/omecodes/store/auth"
	"io"

	"github.com/omecodes/libome/logs"
)

func NewGRPCHandler() ObjectsServer {
	return &gRPCGatewayHandler{}
}

func NewHandler() ObjectsServer {
	return &handler{}
}

type gRPCGatewayHandler struct {
	UnimplementedObjectsServer
}

func (h *gRPCGatewayHandler) CreateCollection(ctx context.Context, request *CreateCollectionRequest) (*CreateCollectionResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	err = CreateCollection(ctx, request.Collection)
	return &CreateCollectionResponse{}, err
}

func (h *gRPCGatewayHandler) GetCollection(ctx context.Context, request *GetCollectionRequest) (*GetCollectionResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	collection, err := GetCollection(ctx, request.Id)
	return &GetCollectionResponse{Collection: collection}, err
}

func (h *gRPCGatewayHandler) ListCollections(ctx context.Context, _ *ListCollectionsRequest) (*ListCollectionsResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	collections, err := ListCollections(ctx)
	return &ListCollectionsResponse{Collections: collections}, err
}

func (h *gRPCGatewayHandler) DeleteCollection(ctx context.Context, request *DeleteCollectionRequest) (*DeleteCollectionResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	err = DeleteCollection(ctx, request.Id)
	return &DeleteCollectionResponse{}, err
}

func (h *gRPCGatewayHandler) PutObject(ctx context.Context, request *PutObjectRequest) (*PutObjectResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	if request.AccessSecurityRules == nil {
		request.AccessSecurityRules = &PathAccessRules{}
	}
	if request.AccessSecurityRules.AccessRules == nil {
		request.AccessSecurityRules.AccessRules = map[string]*AccessRules{}
	}

	id, err := PutObject(ctx, "", request.Object, request.AccessSecurityRules, request.Indexes, PutOptions{})
	if err != nil {
		return nil, err
	}

	return &PutObjectResponse{
		ObjectId: id,
	}, nil
}

func (h *gRPCGatewayHandler) PatchObject(ctx context.Context, request *PatchObjectRequest) (*PatchObjectResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	return &PatchObjectResponse{}, PatchObject(ctx, "", request.Patch, PatchOptions{})
}

func (h *gRPCGatewayHandler) MoveObject(ctx context.Context, request *MoveObjectRequest) (*MoveObjectResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	return &MoveObjectResponse{}, MoveObject(ctx, request.SourceCollection, request.ObjectId, request.TargetCollection, request.AccessSecurityRules, MoveOptions{})
}

func (h *gRPCGatewayHandler) GetObject(ctx context.Context, request *GetObjectRequest) (*GetObjectResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	object, err := GetObject(ctx, "", request.ObjectId, GetOptions{
		At:   request.At,
		Info: request.InfoOnly,
	})

	return &GetObjectResponse{
		Object: object,
	}, err
}

func (h *gRPCGatewayHandler) DeleteObject(ctx context.Context, request *DeleteObjectRequest) (*DeleteObjectResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	err = DeleteObject(ctx, "", request.ObjectId)
	return &DeleteObjectResponse{}, err
}

func (h *gRPCGatewayHandler) ObjectInfo(ctx context.Context, request *ObjectInfoRequest) (*ObjectInfoResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	header, err := GetObjectHeader(ctx, "", request.ObjectId)
	if err != nil {
		return nil, err
	}

	return &ObjectInfoResponse{Header: header}, nil
}

func (h *gRPCGatewayHandler) ListObjects(request *ListObjectsRequest, stream Objects_ListObjectsServer) error {
	ctx, err := auth.ParseMetaInNewContext(stream.Context())
	if err != nil {
		return err
	}

	opts := ListOptions{
		At:     request.At,
		Offset: request.Offset,
	}

	cursor, err := ListObjects(ctx, request.Collection, opts)
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

func (h *gRPCGatewayHandler) SearchObjects(request *SearchObjectsRequest, stream Objects_SearchObjectsServer) error {
	ctx, err := auth.ParseMetaInNewContext(stream.Context())
	if err != nil {
		return err
	}

	cursor, err := SearchObjects(ctx, request.Collection, request.Query)
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
