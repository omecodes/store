package objects

import (
	"context"
	"github.com/omecodes/store/auth"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"

	"github.com/omecodes/libome/logs"
)

func NewGRPCHandler() pb.ObjectsServer {
	return &gRPCGatewayHandler{}
}

func NewHandler() pb.ObjectsServer {
	return &handler{}
}

type gRPCGatewayHandler struct {
	pb.UnimplementedObjectsServer
}

func (h *gRPCGatewayHandler) CreateCollection(ctx context.Context, request *pb.CreateCollectionRequest) (*pb.CreateCollectionResponse, error) {
	var err error
	err = CreateCollection(ctx, request.Collection)
	return &pb.CreateCollectionResponse{}, err
}

func (h *gRPCGatewayHandler) GetCollection(ctx context.Context, request *pb.GetCollectionRequest) (*pb.GetCollectionResponse, error) {
	var err error
	collection, err := GetCollection(ctx, request.Id)
	return &pb.GetCollectionResponse{Collection: collection}, err
}

func (h *gRPCGatewayHandler) ListCollections(ctx context.Context, _ *pb.ListCollectionsRequest) (*pb.ListCollectionsResponse, error) {
	var err error
	collections, err := ListCollections(ctx)
	return &pb.ListCollectionsResponse{Collections: collections}, err
}

func (h *gRPCGatewayHandler) DeleteCollection(ctx context.Context, request *pb.DeleteCollectionRequest) (*pb.DeleteCollectionResponse, error) {
	var err error
	err = DeleteCollection(ctx, request.Id)
	return &pb.DeleteCollectionResponse{}, err
}

func (h *gRPCGatewayHandler) PutObject(ctx context.Context, request *pb.PutObjectRequest) (*pb.PutObjectResponse, error) {
	var err error
	if request.ActionAuthorizedUsers == nil {
		request.ActionAuthorizedUsers = &pb.PathAccessRules{}
	}
	if request.ActionAuthorizedUsers.AccessRules == nil {
		request.ActionAuthorizedUsers.AccessRules = map[string]*pb.ObjectActionsUsers{}
	}

	id, err := PutObject(ctx, "", request.Object, request.ActionAuthorizedUsers, request.Indexes, PutOptions{})
	if err != nil {
		return nil, err
	}

	return &pb.PutObjectResponse{
		ObjectId: id,
	}, nil
}

func (h *gRPCGatewayHandler) PatchObject(ctx context.Context, request *pb.PatchObjectRequest) (*pb.PatchObjectResponse, error) {
	return &pb.PatchObjectResponse{}, PatchObject(ctx, "", request.Patch, PatchOptions{})
}

func (h *gRPCGatewayHandler) MoveObject(ctx context.Context, request *pb.MoveObjectRequest) (*pb.MoveObjectResponse, error) {
	return &pb.MoveObjectResponse{}, MoveObject(ctx, request.SourceCollection, request.ObjectId, request.TargetCollection, request.AccessSecurityRules, MoveOptions{})
}

func (h *gRPCGatewayHandler) GetObject(ctx context.Context, request *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	var err error
	object, err := GetObject(ctx, "", request.ObjectId, GetOptions{
		At:   request.At,
		Info: request.InfoOnly,
	})

	return &pb.GetObjectResponse{
		Object: object,
	}, err
}

func (h *gRPCGatewayHandler) DeleteObject(ctx context.Context, request *pb.DeleteObjectRequest) (*pb.DeleteObjectResponse, error) {
	var err error
	err = DeleteObject(ctx, "", request.ObjectId)
	return &pb.DeleteObjectResponse{}, err
}

func (h *gRPCGatewayHandler) ObjectInfo(ctx context.Context, request *pb.ObjectInfoRequest) (*pb.ObjectInfoResponse, error) {
	var err error
	header, err := GetObjectHeader(ctx, "", request.ObjectId)
	if err != nil {
		return nil, err
	}

	return &pb.ObjectInfoResponse{Header: header}, nil
}

func (h *gRPCGatewayHandler) ListObjects(request *pb.ListObjectsRequest, stream pb.Objects_ListObjectsServer) error {
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

func (h *gRPCGatewayHandler) SearchObjects(request *pb.SearchObjectsRequest, stream pb.Objects_SearchObjectsServer) error {
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
