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

func (b *base) PutObject(ctx context.Context, object *oms.Object, security *oms.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	if b.next == nil {
		return "", errors.New("no handler available")
	}
	return b.next.PutObject(ctx, object, security, opts)
}

func (b *base) GetObject(ctx context.Context, id string, opts oms.GetDataOptions) (*oms.Object, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.GetObject(ctx, id, opts)
}

func (b *base) GetObjectHeader(ctx context.Context, id string) (*oms.Header, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.GetObjectHeader(ctx, id)
}

func (b *base) DeleteObject(ctx context.Context, id string) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.DeleteObject(ctx, id)
}

func (b *base) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.ListObjects(ctx, opts)
}

func (b *base) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	if b.next == nil {
		return errors.New("not handler available")
	}
	return b.next.PatchObject(ctx, patch, opts)
}

func (b *base) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	return nil, nil
}
