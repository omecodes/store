package router

import (
	"context"
	"github.com/omecodes/store/oms"
	"github.com/omecodes/store/pb"
)

type BaseHandler struct {
	next Handler
}

func (b *BaseHandler) PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	return b.next.PutObject(ctx, object, security, opts)
}

func (b *BaseHandler) GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*oms.Object, error) {
	return b.next.GetObject(ctx, id, opts)
}

func (b *BaseHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	return b.next.GetObjectHeader(ctx, id)
}

func (b *BaseHandler) DeleteObject(ctx context.Context, id string) error {
	return b.next.DeleteObject(ctx, id)
}

func (b *BaseHandler) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	return b.next.ListObjects(ctx, opts)
}

func (b *BaseHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	return b.next.PatchObject(ctx, patch, opts)
}

func (b *BaseHandler) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	return b.next.SearchObjects(ctx, params, opts)
}
