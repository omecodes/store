package router

import (
	"context"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/clients"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/pb"
)

// NewGRPCClientHandler creates a router Handler that embed that calls a gRPC service to perform final actions
func NewGRPCClientHandler(nodeType uint32) Handler {
	return &gRPCClientHandler{
		nodeType: nodeType,
	}
}

type gRPCClientHandler struct {
	nodeType uint32
	BaseHandler
}

func (g *gRPCClientHandler) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	client, err := clients.RouterGrpc(ctx, g.nodeType)
	if err != nil {
		return err
	}

	_, err = client.CreateCollection(ctx, &pb.CreateCollectionRequest{Collection: collection})
	return err
}

func (g *gRPCClientHandler) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	client, err := clients.RouterGrpc(ctx, g.nodeType)
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
	client, err := clients.RouterGrpc(ctx, g.nodeType)
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
	client, err := clients.RouterGrpc(ctx, g.nodeType)
	if err != nil {
		return err
	}

	_, err = client.DeleteCollection(ctx, &pb.DeleteCollectionRequest{Id: id})
	return err
}

func (g *gRPCClientHandler) PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.Index, opts pb.PutOptions) (string, error) {
	client, err := clients.RouterGrpc(ctx, g.nodeType)
	if err != nil {
		return "", err
	}

	rsp, err := client.PutObject(auth.SetMetaWithExisting(ctx), &pb.PutObjectRequest{
		Collection:          collection,
		Object:              object,
		Indexes:             indexes,
		AccessSecurityRules: accessSecurityRules,
	})
	if err != nil {
		return "", err
	}

	return rsp.ObjectId, nil
}

func (g *gRPCClientHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts pb.PatchOptions) error {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return err
	}

	_, err = client.PatchObject(auth.SetMetaWithExisting(ctx), &pb.PatchObjectRequest{
		Collection: collection,
		Patch:      patch,
	})
	return err
}

func (g *gRPCClientHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts pb.MoveOptions) error {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return err
	}

	_, err = client.MoveObject(auth.SetMetaWithExisting(ctx), &pb.MoveObjectRequest{
		SourceCollection:    collection,
		ObjectId:            objectID,
		TargetCollection:    targetCollection,
		AccessSecurityRules: accessSecurityRules,
	})
	return err
}

func (g *gRPCClientHandler) GetObject(ctx context.Context, collection string, id string, opts pb.GetOptions) (*pb.Object, error) {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetObject(auth.SetMetaWithExisting(ctx), &pb.GetObjectRequest{
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
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ObjectInfo(auth.SetMetaWithExisting(ctx), &pb.ObjectInfoRequest{
		Collection: collection,
		ObjectId:   id,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Header, nil
}

func (g *gRPCClientHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(auth.SetMetaWithExisting(ctx), &pb.DeleteObjectRequest{
		Collection: collection,
		ObjectId:   id,
	})
	return err
}

func (g *gRPCClientHandler) ListObjects(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error) {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	stream, err := client.ListObjects(auth.SetMetaWithExisting(ctx), &pb.ListObjectsRequest{
		Before:     opts.DateOptions.Before,
		After:      opts.DateOptions.After,
		At:         opts.At,
		Collection: collection,
	})
	if err != nil {
		return nil, err
	}

	closer := pb.CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := pb.BrowseFunc(func() (*pb.Object, error) {
		return stream.Recv()
	})

	return pb.NewCursor(browser, closer), nil
}

func (g *gRPCClientHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*pb.Cursor, error) {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	stream, err := client.SearchObjects(auth.SetMetaWithExisting(ctx), &pb.SearchObjectsRequest{
		Collection: collection,
		Query:      query,
	})
	if err != nil {
		return nil, err
	}

	closer := pb.CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := pb.BrowseFunc(func() (*pb.Object, error) {
		return stream.Recv()
	})

	return pb.NewCursor(browser, closer), nil
}
