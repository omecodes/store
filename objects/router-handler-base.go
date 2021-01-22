package objects

import (
	"context"
	se "github.com/omecodes/store/search-engine"
)

type BaseHandler struct {
	next Handler
}

func (b *BaseHandler) CreateCollection(ctx context.Context, collection *Collection) error {
	return b.next.CreateCollection(ctx, collection)
}

func (b *BaseHandler) GetCollection(ctx context.Context, id string) (*Collection, error) {
	return b.next.GetCollection(ctx, id)
}

func (b *BaseHandler) ListCollections(ctx context.Context) ([]*Collection, error) {
	return b.next.ListCollections(ctx)
}

func (b *BaseHandler) DeleteCollection(ctx context.Context, id string) error {
	return b.next.DeleteCollection(ctx, id)
}

func (b *BaseHandler) PutObject(ctx context.Context, collection string, object *Object, accessSecurityRules *PathAccessRules, indexes []*se.TextIndex, opts PutOptions) (string, error) {
	return b.next.PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func (b *BaseHandler) PatchObject(ctx context.Context, collection string, patch *Patch, opts PatchOptions) error {
	return b.next.PatchObject(ctx, collection, patch, opts)
}

func (b *BaseHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *PathAccessRules, opts MoveOptions) error {
	return b.next.MoveObject(ctx, collection, objectID, targetCollection, accessSecurityRules, opts)
}

func (b *BaseHandler) GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*Object, error) {
	return b.next.GetObject(ctx, collection, id, opts)
}

func (b *BaseHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*Header, error) {
	return b.next.GetObjectHeader(ctx, collection, id)
}

func (b *BaseHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	return b.next.DeleteObject(ctx, collection, id)
}

func (b *BaseHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	return b.next.ListObjects(ctx, collection, opts)
}

func (b *BaseHandler) SearchObjects(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error) {
	return b.next.SearchObjects(ctx, collection, query)
}
