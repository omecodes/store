package objects

import (
	"context"
	se "github.com/omecodes/store/search-engine"
)

type Objects interface {
	CreateCollection(ctx context.Context, collection *Collection) error

	GetCollection(ctx context.Context, id string) (*Collection, error)

	ListCollections(ctx context.Context) ([]*Collection, error)

	DeleteCollection(ctx context.Context, id string) error

	// Save saves object content as JSON in database
	Save(ctx context.Context, collection string, object *Object, index ...*se.TextIndex) error

	// Update applies patch to object with id equals patch.ID()
	Patch(ctx context.Context, collection string, patch *Patch) error

	// Delete removes all content associated with objectID
	Delete(ctx context.Context, collection string, objectID string) error

	// Get gets the object associated with objectID
	Get(ctx context.Context, collection string, objectID string, opts GetOptions) (*Object, error)

	// Info gets header of the object associated with objectID
	Info(ctx context.Context, collection string, objectID string) (*Header, error)

	// List returns a list of at most 'opts.Count' objects
	List(ctx context.Context, collection string, opts ListOptions) (*Cursor, error)

	Search(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error)
}
