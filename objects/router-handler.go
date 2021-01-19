package objects

import (
	"context"
	se "github.com/omecodes/store/search-engine"
)

type ObjectsHandler interface {
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
