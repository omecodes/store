package objects

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/store/clients"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/pb"
)

func NewStoreGrpcClient() Objects {
	return &dbClient{}
}

type dbClient struct {
	pb.UnimplementedHandlerUnitServer
}

func (d *dbClient) Save(ctx context.Context, object *pb.Object, index ...*pb.Index) error {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return err
	}

	_, err = objects.PutObject(ctx, &pb.PutObjectRequest{
		Object:  object,
		Indexes: index,
	})
	return err
}

func (d *dbClient) Patch(ctx context.Context, patch *pb.Patch) error {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return err
	}

	_, err = objects.PatchObject(ctx, &pb.PatchObjectRequest{
		Patch: patch,
	})
	return err
}

func (d *dbClient) Delete(ctx context.Context, objectID string) error {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return err
	}

	_, err = objects.DeleteObject(ctx, &pb.DeleteObjectRequest{
		ObjectId: objectID,
	})
	return err
}

func (d *dbClient) List(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}

	stream, err := objects.ListObjects(ctx, &pb.ListObjectsRequest{
		Before:     opts.DateOptions.Before,
		After:      opts.DateOptions.After,
		At:         opts.At,
		Collection: opts.CollectionOptions.Name,
		FullObject: opts.CollectionOptions.FullObject,
		Condition:  opts.Condition,
	})

	closer := pb.CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := pb.BrowseFunc(func() (*pb.Object, error) {
		return stream.Recv()
	})

	return pb.NewCursor(browser, closer), nil
}

func (d *dbClient) Get(ctx context.Context, objectID string, opts pb.GetOptions) (*pb.Object, error) {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.GetObject(ctx, &pb.GetObjectRequest{ObjectId: objectID})
	if err != nil {
		return nil, err
	}

	return rsp.Object, nil
}

func (d *dbClient) Info(ctx context.Context, objectID string) (*pb.Header, error) {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.ObjectInfo(ctx, &pb.ObjectInfoRequest{ObjectId: objectID})
	if err != nil {
		return nil, err
	}
	return rsp.Header, nil
}

func (d *dbClient) Clear() error {
	return errors.Forbidden
}
