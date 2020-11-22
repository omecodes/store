package oms

import (
	"context"
	"io"
)

type doNothing struct{}

func (d *doNothing) SetSettings(ctx context.Context, value *JSON, opts SettingsOptions) error {
	return nil
}

func (d *doNothing) GetSettings(ctx context.Context, opts SettingsOptions) (*JSON, error) {
	return nil, nil
}

func (d *doNothing) ListWorkers(ctx context.Context) ([]*JSON, error) {
	return nil, nil
}

func (d *doNothing) RegisterWorker(ctx context.Context, info *JSON) error {
	return nil
}

func (d *doNothing) GetCollections(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (d *doNothing) PutData(ctx context.Context, data *Object, opts PutDataOptions) error {
	return nil
}

func (d *doNothing) PatchData(ctx context.Context, collection string, id string, content io.Reader, size int64, opts PatchOptions) error {
	return nil
}

func (d *doNothing) GetData(ctx context.Context, collection string, id string, opts GetDataOptions) (*Object, error) {
	return nil, nil
}

func (d *doNothing) Info(ctx context.Context, collection string, id string) (*Info, error) {
	return nil, nil
}

func (d *doNothing) Delete(ctx context.Context, collection string, id string) error {
	return nil
}

func (d *doNothing) List(ctx context.Context, collection string, opts ListOptions) (*DataList, error) {
	return nil, nil
}

func (d *doNothing) SaveGraft(ctx context.Context, graft *Graft) (string, error) {
	return "", nil
}

func (d *doNothing) GetGraft(ctx context.Context, collection string, dataID string, id string) (*Graft, error) {
	return nil, nil
}

func (d *doNothing) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*GraftInfo, error) {
	return nil, nil
}

func (d *doNothing) GraftBulk(ctx context.Context, collection string, dataID string, ids ...string) (*GraftList, error) {
	return nil, nil
}

func (d *doNothing) ListGrafts(ctx context.Context, collection string, dataID string, opts ListOptions) (*GraftList, error) {
	return nil, nil
}

func (d *doNothing) DeleteGraft(ctx context.Context, collection string, dataID string, id string) error {
	return nil
}
