package objects

import (
	"context"
	"github.com/omecodes/store/auth"
	pb "github.com/omecodes/store/gen/go/proto"
)

// NewGRPCObjectsClientHandler creates a router ObjectsHandler that embed that calls a gRPC service to perform final actions
func NewGRPCObjectsClientHandler(nodeType uint32) Handler {
	return &gRPCClientHandler{
		nodeType: nodeType,
	}
}

type gRPCClientHandler struct {
	nodeType uint32
	BaseHandler
}

func (g *gRPCClientHandler) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return err
	}

	_, err = client.CreateCollection(ctx, &pb.CreateCollectionRequest{Collection: collection})
	return err
}

func (g *gRPCClientHandler) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetCollection(ctx, &pb.GetCollectionRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return rsp.Collection, err
}

func (g *gRPCClientHandler) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ListCollections(ctx, &pb.ListCollectionsRequest{})
	if err != nil {
		return nil, err
	}
	return rsp.Collections, err
}

func (g *gRPCClientHandler) DeleteCollection(ctx context.Context, id string) error {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return err
	}

	_, err = client.DeleteCollection(ctx, &pb.DeleteCollectionRequest{Id: id})
	return err
}

func (g *gRPCClientHandler) PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.TextIndex, opts PutOptions) (string, error) {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return "", err
	}

	newCtx, err := auth.ContextWithMeta(ctx)
	if err != nil {
		return "", err
	}

	rsp, err := client.PutObject(newCtx, &pb.PutObjectRequest{
		Collection:            collection,
		Object:                object,
		Indexes:               indexes,
		ActionAuthorizedUsers: accessSecurityRules,
	})
	if err != nil {
		return "", err
	}

	return rsp.ObjectId, nil
}

func (g *gRPCClientHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts PatchOptions) error {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return err
	}

	newCtx, err := auth.ContextWithMeta(ctx)
	if err != nil {
		return err
	}

	_, err = client.PatchObject(newCtx, &pb.PatchObjectRequest{
		Collection: collection,
		Patch:      patch,
	})
	return err
}

func (g *gRPCClientHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts MoveOptions) error {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return err
	}

	newCtx, err := auth.ContextWithMeta(ctx)
	if err != nil {
		return err
	}
	_, err = client.MoveObject(newCtx, &pb.MoveObjectRequest{
		SourceCollection:    collection,
		ObjectId:            objectID,
		TargetCollection:    targetCollection,
		AccessSecurityRules: accessSecurityRules,
	})
	return err
}

func (g *gRPCClientHandler) GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*pb.Object, error) {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return nil, err
	}

	newCtx, err := auth.ContextWithMeta(ctx)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetObject(newCtx, &pb.GetObjectRequest{
		Collection: collection,
		ObjectId:   id,
		At:         opts.At,
		InfoOnly:   opts.Info,
	})
	if err != nil {
		return nil, err
	}

	return rsp.Object, nil
}

func (g *gRPCClientHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error) {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return nil, err
	}

	newCtx, err := auth.ContextWithMeta(ctx)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ObjectInfo(newCtx, &pb.ObjectInfoRequest{
		Collection: collection,
		ObjectId:   id,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Header, nil
}

func (g *gRPCClientHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return err
	}

	newCtx, err := auth.ContextWithMeta(ctx)
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(newCtx, &pb.DeleteObjectRequest{
		Collection: collection,
		ObjectId:   id,
	})
	return err
}

func (g *gRPCClientHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return nil, err
	}

	newCtx, err := auth.ContextWithMeta(ctx)
	if err != nil {
		return nil, err
	}

	stream, err := client.ListObjects(newCtx, &pb.ListObjectsRequest{
		Offset:     opts.Offset,
		At:         opts.At,
		Collection: collection,
	})
	if err != nil {
		return nil, err
	}

	closer := CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := BrowseFunc(func() (*pb.Object, error) {
		return stream.Recv()
	})

	return NewCursor(browser, closer), nil
}

func (g *gRPCClientHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*Cursor, error) {
	client, err := grpcClient(ctx, g.nodeType)
	if err != nil {
		return nil, err
	}

	newCtx, err := auth.ContextWithMeta(ctx)
	if err != nil {
		return nil, err
	}

	stream, err := client.SearchObjects(newCtx, &pb.SearchObjectsRequest{
		Collection: collection,
		Query:      query,
	})
	if err != nil {
		return nil, err
	}

	closer := CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := BrowseFunc(func() (*pb.Object, error) {
		return stream.Recv()
	})

	return NewCursor(browser, closer), nil
}
