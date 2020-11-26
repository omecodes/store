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

func (d *dummyHandler) GetCollections(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (d *dummyHandler) PutData(ctx context.Context, object *oms.Object, opts oms.PutDataOptions) (string, error) {
	return "", nil
}

func (d *dummyHandler) PatchData(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	return nil
}

func (d *dummyHandler) GetData(ctx context.Context, id string, opts oms.GetDataOptions) (*oms.Object, error) {
	return nil, nil
}

func (d *dummyHandler) Info(ctx context.Context, id string) (*oms.Info, error) {
	return nil, nil
}

func (d *dummyHandler) Delete(ctx context.Context, id string) error {
	return nil
}

func (d *dummyHandler) List(ctx context.Context, opts oms.ListOptions) (*oms.ListResult, error) {
	return nil, nil
}

func (d *dummyHandler) Search(ctx context.Context, opts oms.SearchOptions) (*oms.SearchResult, error) {
	return nil, nil
}
