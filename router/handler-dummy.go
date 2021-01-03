package router

import (
	"context"
	"github.com/omecodes/store/oms"
	"github.com/omecodes/store/pb"
)

type dummyHandler struct{}

func (d *dummyHandler) PutObject(ctx context.Context, object *pb.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	return "", nil
}

func (d *dummyHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	return nil
}

func (d *dummyHandler) GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*pb.Object, error) {
	return nil, nil
}

func (d *dummyHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	return nil, nil
}

func (d *dummyHandler) DeleteObject(ctx context.Context, id string) error {
	return nil
}

func (d *dummyHandler) ListObjects(ctx context.Context, opts oms.ListOptions) (*pb.ObjectList, error) {
	return nil, nil
}

func (d *dummyHandler) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*pb.ObjectList, error) {
	return nil, nil
}
