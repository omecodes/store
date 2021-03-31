package objects

import (
	"context"
	se "github.com/omecodes/store/search-engine"
)

type Handler interface {
	CreateCollection(ctx context.Context, collection *Collection) error
	GetCollection(ctx context.Context, id string) (*Collection, error)
	ListCollections(ctx context.Context) ([]*Collection, error)
	DeleteCollection(ctx context.Context, id string) error

	PutObject(ctx context.Context, collection string, object *Object, accessSecurityRules *PathAccessRules, indexes []*se.TextIndex, opts PutOptions) (string, error)
	PatchObject(ctx context.Context, collection string, patch *Patch, opts PatchOptions) error
	MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *PathAccessRules, opts MoveOptions) error
	GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*Object, error)
	GetObjectHeader(ctx context.Context, collection string, id string) (*Header, error)
	DeleteObject(ctx context.Context, collection string, id string) error
	ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error)
	SearchObjects(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error)
}

func CreateCollection(ctx context.Context, collection *Collection) error {
	return GetRouterHandler(ctx).CreateCollection(ctx, collection)
}

func GetCollection(ctx context.Context, id string) (*Collection, error) {
	return GetRouterHandler(ctx).GetCollection(ctx, id)
}

func ListCollections(ctx context.Context) ([]*Collection, error) {
	return GetRouterHandler(ctx).ListCollections(ctx)
}

func DeleteCollection(ctx context.Context, id string) error {
	return GetRouterHandler(ctx).DeleteCollection(ctx, id)
}

func PutObject(ctx context.Context, collection string, object *Object, accessSecurityRules *PathAccessRules, indexes []*se.TextIndex, opts PutOptions) (string, error) {
	return GetRouterHandler(ctx).PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func PatchObject(ctx context.Context, collection string, patch *Patch, opts PatchOptions) error {
	return GetRouterHandler(ctx).PatchObject(ctx, collection, patch, opts)
}

func MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *PathAccessRules, opts MoveOptions) error {
	return GetRouterHandler(ctx).MoveObject(ctx, collection, objectID, targetCollection, accessSecurityRules, opts)
}

func GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*Object, error) {
	return GetRouterHandler(ctx).GetObject(ctx, collection, id, opts)
}

func GetObjectHeader(ctx context.Context, collection string, id string) (*Header, error) {
	return GetRouterHandler(ctx).GetObjectHeader(ctx, collection, id)
}

func DeleteObject(ctx context.Context, collection string, id string) error {
	return GetRouterHandler(ctx).DeleteObject(ctx, collection, id)
}

func ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	return GetRouterHandler(ctx).ListObjects(ctx, collection, opts)
}

func SearchObjects(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error) {
	return GetRouterHandler(ctx).SearchObjects(ctx, collection, query)
}
