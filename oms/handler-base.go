package oms

import (
	"context"
	"errors"
	"io"
)

type base struct {
	next Handler
}

func (b *base) SetSettings(ctx context.Context, value *JSON, opts SettingsOptions) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.SetSettings(ctx, value, opts)
}

func (b *base) GetSettings(ctx context.Context, opts SettingsOptions) (*JSON, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.GetSettings(ctx, opts)
}

func (b *base) ListWorkers(ctx context.Context) ([]*JSON, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.ListWorkers(ctx)
}

func (b *base) RegisterWorker(ctx context.Context, info *JSON) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.RegisterWorker(ctx, info)
}

func (b *base) GetCollections(ctx context.Context) ([]string, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.GetCollections(ctx)
}

func (b *base) PutData(ctx context.Context, data *Object, opts PutDataOptions) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.PutData(ctx, data, opts)
}

func (b *base) GetData(ctx context.Context, collection string, id string, opts GetDataOptions) (*Object, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.GetData(ctx, collection, id, opts)
}

func (b *base) Info(ctx context.Context, collection string, id string) (*Info, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.Info(ctx, collection, id)
}

func (b *base) Delete(ctx context.Context, collection string, id string) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.Delete(ctx, collection, id)
}

func (b *base) List(ctx context.Context, collection string, opts ListOptions) (*DataList, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.List(ctx, collection, opts)
}

func (b *base) PatchData(ctx context.Context, collection string, id string, content io.Reader, size int64, opts PatchOptions) error {
	if b.next == nil {
		return errors.New("not handler available")
	}
	return b.next.PatchData(ctx, collection, id, content, size, opts)
}

func (b *base) SaveGraft(ctx context.Context, graft *Graft) (string, error) {
	if b.next == nil {
		return "", errors.New("no handler available")
	}
	return b.next.SaveGraft(ctx, graft)
}

func (b *base) GetGraft(ctx context.Context, collection string, dataID string, id string) (*Graft, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.GetGraft(ctx, collection, dataID, id)
}

func (b *base) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*GraftInfo, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.GraftInfo(ctx, collection, dataID, id)
}

func (b *base) ListGrafts(ctx context.Context, collection string, dataID string, opts ListOptions) (*GraftList, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.ListGrafts(ctx, collection, dataID, opts)
}

func (b *base) DeleteGraft(ctx context.Context, collection string, dataID string, id string) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.DeleteGraft(ctx, collection, dataID, id)
}
