package router

import (
	"context"
	"github.com/omecodes/store/pb"
)

type objectsDummyHandler struct{}

func (d *objectsDummyHandler) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	return nil
}

func (d *objectsDummyHandler) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	return nil, nil
}

func (d *objectsDummyHandler) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	return nil, nil
}

func (d *objectsDummyHandler) DeleteCollection(ctx context.Context, id string) error {
	return nil
}

func (d *objectsDummyHandler) PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.TextIndex, opts pb.PutOptions) (string, error) {
	return "", nil
}

func (d *objectsDummyHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts pb.PatchOptions) error {
	return nil
}

func (d *objectsDummyHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts pb.MoveOptions) error {
	return nil
}

func (d *objectsDummyHandler) GetObject(ctx context.Context, collection string, id string, opts pb.GetOptions) (*pb.Object, error) {
	return nil, nil
}

func (d *objectsDummyHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error) {
	return nil, nil
}

func (d *objectsDummyHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	return nil
}

func (d *objectsDummyHandler) ListObjects(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error) {
	return nil, nil
}

func (d *objectsDummyHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*pb.Cursor, error) {
	return nil, nil
}
