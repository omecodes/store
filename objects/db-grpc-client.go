package objects

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/common"
	se "github.com/omecodes/store/search-engine"
)

func NewDBGrpcClient() DB {
	return &dbClient{}
}

type dbClient struct{}

func (d *dbClient) CreateCollection(ctx context.Context, collection *Collection) error {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return err
	}

	_, err = objects.CreateCollection(ctx, &CreateCollectionRequest{
		Collection: collection})
	return err
}

func (d *dbClient) GetCollection(ctx context.Context, id string) (*Collection, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.GetCollection(ctx, &GetCollectionRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Collection, err
}

func (d *dbClient) ListCollections(ctx context.Context) ([]*Collection, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.ListCollections(ctx, &ListCollectionsRequest{})
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

	_, err = objects.DeleteCollection(ctx, &DeleteCollectionRequest{Id: id})
	return err
}

func (d *dbClient) Save(ctx context.Context, collection string, object *Object, index ...*se.TextIndex) error {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return err
	}

	_, err = objects.PutObject(ctx, &PutObjectRequest{
		Collection: collection,
		Object:     object,
		Indexes:    index,
	})
	return err
}

func (d *dbClient) Patch(ctx context.Context, collection string, patch *Patch) error {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return err
	}

	_, err = objects.PatchObject(ctx, &PatchObjectRequest{
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

	_, err = objects.DeleteObject(ctx, &DeleteObjectRequest{
		Collection: collection,
		ObjectId:   objectID,
	})
	return err
}

func (d *dbClient) Get(ctx context.Context, collection string, objectID string, opts GetOptions) (*Object, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.GetObject(ctx, &GetObjectRequest{
		Collection: collection, ObjectId: objectID})
	if err != nil {
		return nil, err
	}

	return rsp.Object, nil
}

func (d *dbClient) Info(ctx context.Context, collection string, objectID string) (*Header, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.ObjectInfo(ctx, &ObjectInfoRequest{
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

	stream, err := objects.ListObjects(ctx, &ListObjectsRequest{
		Collection: collection,
		Offset:     opts.Offset,
		At:         opts.At,
	})

	closer := CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := BrowseFunc(func() (*Object, error) {
		return stream.Recv()
	})

	return NewCursor(browser, closer), nil
}

func (d *dbClient) Search(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error) {
	objects, err := grpcClient(ctx, common.ServiceTypeFilesStorage)
	if err != nil {
		return nil, err
	}

	stream, err := objects.SearchObjects(ctx, &SearchObjectsRequest{
		Collection: collection,
		Query:      query,
	})

	closer := CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := BrowseFunc(func() (*Object, error) {
		return stream.Recv()
	})

	return NewCursor(browser, closer), nil
}

func (d *dbClient) Clear() error {
	return errors.Unauthorized("this operation is not authorize")
}
