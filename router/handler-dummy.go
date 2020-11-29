package router

import (
	"context"
	"github.com/omecodes/omestore/oms"
)

type dummyHandler struct{}

func (d *dummyHandler) SetSettings(ctx context.Context, value *oms.JSON, opts oms.SettingsOptions) error {
	return nil
}

func (d *dummyHandler) GetSettings(ctx context.Context, opts oms.SettingsOptions) (*oms.JSON, error) {
	return nil, nil
}

func (d *dummyHandler) ListWorkers(ctx context.Context) ([]*oms.JSON, error) {
	return nil, nil
}

func (d *dummyHandler) RegisterWorker(ctx context.Context, info *oms.JSON) error {
	return nil
}
func (d *dummyHandler) PutObject(ctx context.Context, object *oms.Object, security *oms.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	return "", nil
}

func (d *dummyHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	return nil
}

func (d *dummyHandler) GetObject(ctx context.Context, id string, opts oms.GetDataOptions) (*oms.Object, error) {
	return nil, nil
}

func (d *dummyHandler) GetObjectHeader(ctx context.Context, id string) (*oms.Header, error) {
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
