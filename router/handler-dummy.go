package router

import (
	"context"
	"github.com/omecodes/store/pb"
)

type dummyHandler struct{}

func (d *dummyHandler) PutObject(ctx context.Context, object *pb.Object, security *pb.PathAccessRules, indexes []*pb.Index, opts pb.PutOptions) (string, error) {
	return "", nil
}

func (d *dummyHandler) PatchObject(ctx context.Context, patch *pb.Patch, opts pb.PatchOptions) error {
	return nil
}

func (d *dummyHandler) GetObject(ctx context.Context, id string, opts pb.GetOptions) (*pb.Object, error) {
	return nil, nil
}

func (d *dummyHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	return nil, nil
}

func (d *dummyHandler) DeleteObject(ctx context.Context, id string) error {
	return nil
}

func (d *dummyHandler) ListObjects(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	return nil, nil
}
