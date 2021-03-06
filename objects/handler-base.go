package objects

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
)

type BaseHandler struct {
	next Handler
}

func (b *BaseHandler) CreateCollection(ctx context.Context, collection *pb.Collection, opts CreateCollectionOptions) error {
	return b.next.CreateCollection(ctx, collection, opts)
}

func (b *BaseHandler) GetCollection(ctx context.Context, id string, opts GetCollectionOptions) (*pb.Collection, error) {
	return b.next.GetCollection(ctx, id, opts)
}

func (b *BaseHandler) ListCollections(ctx context.Context, opts ListCollectionOptions) ([]*pb.Collection, error) {
	return b.next.ListCollections(ctx, opts)
}

func (b *BaseHandler) DeleteCollection(ctx context.Context, id string, opts DeleteCollectionOptions) error {
	return b.next.DeleteCollection(ctx, id, opts)
}

func (b *BaseHandler) PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.TextIndex, opts PutOptions) (string, error) {
	return b.next.PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func (b *BaseHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts PatchOptions) error {
	return b.next.PatchObject(ctx, collection, patch, opts)
}

func (b *BaseHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts MoveOptions) error {
	return b.next.MoveObject(ctx, collection, objectID, targetCollection, accessSecurityRules, opts)
}

func (b *BaseHandler) GetObject(ctx context.Context, collection string, id string, opts GetObjectOptions) (*pb.Object, error) {
	return b.next.GetObject(ctx, collection, id, opts)
}

func (b *BaseHandler) GetObjectHeader(ctx context.Context, collection string, id string, opts GetHeaderOptions) (*pb.Header, error) {
	return b.next.GetObjectHeader(ctx, collection, id, opts)
}

func (b *BaseHandler) DeleteObject(ctx context.Context, collection string, id string, opts DeleteObjectOptions) error {
	return b.next.DeleteObject(ctx, collection, id, opts)
}

func (b *BaseHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	return b.next.ListObjects(ctx, collection, opts)
}

func (b *BaseHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery, opts SearchObjectsOptions) (*Cursor, error) {
	return b.next.SearchObjects(ctx, collection, query, opts)
}
