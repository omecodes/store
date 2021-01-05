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

func (g *gRPCClientHandler) PutObject(ctx context.Context, object *pb.Object, security *pb.PathAccessRules, indexes []*pb.Index, opts pb.PutOptions) (string, error) {
	client, err := clients.RouterGrpc(ctx, g.nodeType)
	if err != nil {
		return "", err
	}

	rsp, err := client.PutObject(auth.SetMetaWithExisting(ctx), &pb.PutObjectRequest{
		AccessSecurityRules: security,
		Object:              object,
		Indexes:             indexes,
	})
	if err != nil {
		return "", err
	}

	return rsp.ObjectId, nil
}

func (g *gRPCClientHandler) PatchObject(ctx context.Context, patch *pb.Patch, opts pb.PatchOptions) error {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return err
	}

	_, err = client.PatchObject(auth.SetMetaWithExisting(ctx), &pb.PatchObjectRequest{
		Patch: patch,
	})
	return err
}

func (g *gRPCClientHandler) GetObject(ctx context.Context, id string, opts pb.GetOptions) (*pb.Object, error) {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetObject(auth.SetMetaWithExisting(ctx), &pb.GetObjectRequest{
		ObjectId: id,
		At:       opts.At,
		InfoOnly: opts.Info,
	})
	if err != nil {
		return nil, err
	}

	return rsp.Object, nil
}

func (g *gRPCClientHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ObjectInfo(auth.SetMetaWithExisting(ctx), &pb.ObjectInfoRequest{
		ObjectId: id,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Header, nil
}

func (g *gRPCClientHandler) DeleteObject(ctx context.Context, id string) error {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(auth.SetMetaWithExisting(ctx), &pb.DeleteObjectRequest{
		ObjectId: id,
	})
	return err
}

func (g *gRPCClientHandler) ListObjects(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	stream, err := client.ListObjects(auth.SetMetaWithExisting(ctx), &pb.ListObjectsRequest{
		Before:     opts.DateOptions.Before,
		After:      opts.DateOptions.After,
		At:         opts.At,
		Collection: opts.CollectionOptions.Name,
		FullObject: opts.CollectionOptions.FullObject,
		Condition:  opts.Condition,
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
