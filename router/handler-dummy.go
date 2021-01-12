package router

import (
	"context"
	"github.com/omecodes/store/pb"
)

type dummyHandler struct{}

func (d *dummyHandler) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	return nil
}

func (d *dummyHandler) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	return nil, nil
}

func (d *dummyHandler) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	return nil, nil
}

func (d *dummyHandler) DeleteCollection(ctx context.Context, id string) error {
	return nil
}

func (d *dummyHandler) PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.Index, opts pb.PutOptions) (string, error) {
	return "", nil
}

func (d *dummyHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts pb.PatchOptions) error {
	return nil
}

func (d *dummyHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts pb.MoveOptions) error {
	return nil
}

func (d *dummyHandler) GetObject(ctx context.Context, collection string, id string, opts pb.GetOptions) (*pb.Object, error) {
	return nil, nil
}

func (d *dummyHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error) {
	return nil, nil
}

func (d *dummyHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	return nil
}

func (d *dummyHandler) ListObjects(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error) {
	return nil, nil
}

func (d *dummyHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*pb.Cursor, error) {
	return nil, nil
}
