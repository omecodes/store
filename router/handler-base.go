package router

import (
	"context"
	"github.com/omecodes/store/pb"
)

type BaseHandler struct {
	next ObjectsHandler
}

func (b *BaseHandler) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	return b.next.CreateCollection(ctx, collection)
}

func (b *BaseHandler) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	return b.next.GetCollection(ctx, id)
}

func (b *BaseHandler) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	return b.next.ListCollections(ctx)
}

func (b *BaseHandler) DeleteCollection(ctx context.Context, id string) error {
	return b.next.DeleteCollection(ctx, id)
}

func (b *BaseHandler) PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.TextIndex, opts pb.PutOptions) (string, error) {
	return b.next.PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func (b *BaseHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts pb.PatchOptions) error {
	return b.next.PatchObject(ctx, collection, patch, opts)
}

func (b *BaseHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts pb.MoveOptions) error {
	return b.next.MoveObject(ctx, collection, objectID, targetCollection, accessSecurityRules, opts)
}

func (b *BaseHandler) GetObject(ctx context.Context, collection string, id string, opts pb.GetOptions) (*pb.Object, error) {
	return b.next.GetObject(ctx, collection, id, opts)
}

func (b *BaseHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error) {
	return b.next.GetObjectHeader(ctx, collection, id)
}

func (b *BaseHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	return b.next.DeleteObject(ctx, collection, id)
}

func (b *BaseHandler) ListObjects(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error) {
	return b.next.ListObjects(ctx, collection, opts)
}

func (b *BaseHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*pb.Cursor, error) {
	return b.next.SearchObjects(ctx, collection, query)
}
