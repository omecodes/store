package router

import (
	"context"
	"github.com/omecodes/store/pb"
)

type BaseHandler struct {
	next Handler
}

func (b *BaseHandler) PutObject(ctx context.Context, object *pb.Object, security *pb.PathAccessRules, indexes []*pb.Index, opts pb.PutOptions) (string, error) {
	return b.next.PutObject(ctx, object, security, indexes, opts)
}

func (b *BaseHandler) GetObject(ctx context.Context, id string, opts pb.GetOptions) (*pb.Object, error) {
	return b.next.GetObject(ctx, id, opts)
}

func (b *BaseHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	return b.next.GetObjectHeader(ctx, id)
}

func (b *BaseHandler) DeleteObject(ctx context.Context, id string) error {
	return b.next.DeleteObject(ctx, id)
}

func (b *BaseHandler) ListObjects(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	return b.next.ListObjects(ctx, opts)
}

func (b *BaseHandler) PatchObject(ctx context.Context, patch *pb.Patch, opts pb.PatchOptions) error {
	return b.next.PatchObject(ctx, patch, opts)
}
