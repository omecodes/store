package router

import (
	"context"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
)

type base struct {
	next Handler
}

func (b *base) SetSettings(ctx context.Context, name string, value string, opts oms.SettingsOptions) error {
	return b.next.SetSettings(ctx, name, value, opts)
}

func (b *base) GetSettings(ctx context.Context, name string) (string, error) {
	return b.next.GetSettings(ctx, name)
}

func (b *base) DeleteSettings(ctx context.Context, name string) error {
	return b.next.DeleteSettings(ctx, name)
}

func (b *base) ClearSettings(ctx context.Context) error {
	return b.next.ClearSettings(ctx)
}

func (b *base) ListWorkers(ctx context.Context) ([]*oms.JSON, error) {
	return b.next.ListWorkers(ctx)
}

func (b *base) RegisterWorker(ctx context.Context, info *oms.JSON) error {
	return b.next.RegisterWorker(ctx, info)
}

func (b *base) PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	return b.next.PutObject(ctx, object, security, opts)
}

func (b *base) GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*oms.Object, error) {
	return b.next.GetObject(ctx, id, opts)
}

func (b *base) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	return b.next.GetObjectHeader(ctx, id)
}

func (b *base) DeleteObject(ctx context.Context, id string) error {
	return b.next.DeleteObject(ctx, id)
}

func (b *base) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	return b.next.ListObjects(ctx, opts)
}

func (b *base) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	return b.next.PatchObject(ctx, patch, opts)
}

func (b *base) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	return b.next.SearchObjects(ctx, params, opts)
}
