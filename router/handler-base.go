package router

import (
	"context"
	"errors"
	"github.com/omecodes/omestore/oms"
)

type base struct {
	next Handler
}

func (b *base) SetSettings(ctx context.Context, value *oms.JSON, opts oms.SettingsOptions) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.SetSettings(ctx, value, opts)
}

func (b *base) GetSettings(ctx context.Context, opts oms.SettingsOptions) (*oms.JSON, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.GetSettings(ctx, opts)
}

func (b *base) ListWorkers(ctx context.Context) ([]*oms.JSON, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.ListWorkers(ctx)
}

func (b *base) RegisterWorker(ctx context.Context, info *oms.JSON) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.RegisterWorker(ctx, info)
}

func (b *base) PutData(ctx context.Context, object *oms.Object, opts oms.PutDataOptions) (string, error) {
	if b.next == nil {
		return "", errors.New("no handler available")
	}
	return b.next.PutData(ctx, object, opts)
}

func (b *base) GetData(ctx context.Context, id string, opts oms.GetDataOptions) (*oms.Object, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.GetData(ctx, id, opts)
}

func (b *base) Info(ctx context.Context, id string) (*oms.Info, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.Info(ctx, id)
}

func (b *base) Delete(ctx context.Context, id string) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.Delete(ctx, id)
}

func (b *base) List(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.List(ctx, opts)
}

func (b *base) PatchData(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	if b.next == nil {
		return errors.New("not handler available")
	}
	return b.next.PatchData(ctx, patch, opts)
}

func (b *base) Search(ctx context.Context, opts oms.SearchOptions) (*oms.ObjectList, error) {
	return nil, nil
}
