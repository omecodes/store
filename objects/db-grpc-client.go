package objects

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/common"
	pb "github.com/omecodes/store/gen/go/proto"
)

func NewDBGrpcClient() DB {
	return &dbClient{}
}

type dbClient struct{}

func (d *dbClient) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return err
	}

	_, err = objects.CreateCollection(ctx, &pb.CreateCollectionRequest{
		Collection: collection})
	return err
}

func (d *dbClient) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.GetCollection(ctx, &pb.GetCollectionRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Collection, err
}

func (d *dbClient) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.ListCollections(ctx, &pb.ListCollectionsRequest{})
	if err != nil {
		return nil, err
	}
	return rsp.Collections, err
}

func (d *dbClient) DeleteCollection(ctx context.Context, id string) error {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return err
	}

	_, err = objects.DeleteCollection(ctx, &pb.DeleteCollectionRequest{Id: id})
	return err
}

func (d *dbClient) Save(ctx context.Context, collection string, object *pb.Object, index ...*pb.TextIndex) error {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return err
	}

	_, err = objects.PutObject(ctx, &pb.PutObjectRequest{
		Collection: collection,
		Object:     object,
		Indexes:    index,
	})
	return err
}

func (d *dbClient) Patch(ctx context.Context, collection string, patch *pb.Patch) error {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return err
	}

	_, err = objects.PatchObject(ctx, &pb.PatchObjectRequest{
		Collection: collection,
		Patch:      patch,
	})
	return err
}

func (d *dbClient) Delete(ctx context.Context, collection string, objectID string) error {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return err
	}

	_, err = objects.DeleteObject(ctx, &pb.DeleteObjectRequest{
		Collection: collection,
		ObjectId:   objectID,
	})
	return err
}

func (d *dbClient) Get(ctx context.Context, collection string, objectID string, opts GetObjectOptions) (*pb.Object, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.GetObject(ctx, &pb.GetObjectRequest{
		Collection: collection, ObjectId: objectID})
	if err != nil {
		return nil, err
	}

	return rsp.Object, nil
}

func (d *dbClient) Info(ctx context.Context, collection string, objectID string) (*pb.Header, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.ObjectInfo(ctx, &pb.ObjectInfoRequest{
		Collection: collection, ObjectId: objectID})
	if err != nil {
		return nil, err
	}
	return rsp.Header, nil
}

func (d *dbClient) List(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	stream, err := objects.ListObjects(ctx, &pb.ListObjectsRequest{
		Collection: collection,
		Offset:     opts.Offset,
		At:         opts.At,
	})

	closer := CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := BrowseFunc(func() (*pb.Object, error) {
		return stream.Recv()
	})

	return NewCursor(browser, closer), nil
}

func (d *dbClient) Search(ctx context.Context, collection string, query *pb.SearchQuery) (*Cursor, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	stream, err := objects.SearchObjects(ctx, &pb.SearchObjectsRequest{
		Collection: collection,
		Query:      query,
	})

	closer := CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := BrowseFunc(func() (*pb.Object, error) {
		return stream.Recv()
	})

	return NewCursor(browser, closer), nil
}

func (d *dbClient) Clear() error {
	return errors.Unauthorized("this operation is not authorize")
}
