package router

import (
	"context"
	"github.com/omecodes/store/objects"
	"github.com/omecodes/store/pb"
)

type dummyHandler struct{}

func (d *dummyHandler) PutObject(ctx context.Context, object *pb.Object, security *pb.PathAccessRules, opts objects.PutDataOptions) (string, error) {
	return "", nil
}

func (d *dummyHandler) PatchObject(ctx context.Context, patch *pb.Patch, opts objects.PatchOptions) error {
	return nil
}

func (d *dummyHandler) GetObject(ctx context.Context, id string, opts objects.GetObjectOptions) (*pb.Object, error) {
	return nil, nil
}

func (d *dummyHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	return nil, nil
}

func (d *dummyHandler) DeleteObject(ctx context.Context, id string) error {
	return nil
}

func (d *dummyHandler) ListObjects(ctx context.Context, opts objects.ListOptions) (*pb.ObjectList, error) {
	return nil, nil
}

func (d *dummyHandler) SearchObjects(ctx context.Context, params objects.SearchParams, opts objects.SearchOptions) (*pb.ObjectList, error) {
	return nil, nil
}
