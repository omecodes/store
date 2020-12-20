package router

import (
	"context"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
)

type dummyHandler struct{}

func (d *dummyHandler) SetSettings(ctx context.Context, name string, value string, opts oms.SettingsOptions) error {
	return nil
}

func (d *dummyHandler) GetSettings(ctx context.Context, name string) (string, error) {
	return "", nil
}

func (d *dummyHandler) DeleteSettings(ctx context.Context, name string) error {
	return nil
}

func (d *dummyHandler) ClearSettings(ctx context.Context) error {
	return nil
}

func (d *dummyHandler) PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	return "", nil
}

func (d *dummyHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	return nil
}

func (d *dummyHandler) GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*oms.Object, error) {
	return nil, nil
}

func (d *dummyHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	return nil, nil
}

func (d *dummyHandler) DeleteObject(ctx context.Context, id string) error {
	return nil
}

func (d *dummyHandler) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	return nil, nil
}

func (d *dummyHandler) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	return nil, nil
}
