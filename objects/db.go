package objects

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
)

type DB interface {
	CreateCollection(ctx context.Context, collection *pb.Collection) error

	GetCollection(ctx context.Context, id string) (*pb.Collection, error)

	ListCollections(ctx context.Context) ([]*pb.Collection, error)

	DeleteCollection(ctx context.Context, id string) error

	// Save saves object content as JSON in database
	Save(ctx context.Context, collection string, object *pb.Object, index ...*pb.TextIndex) error

	Patch(ctx context.Context, collection string, patch *pb.Patch) error

	// Delete removes all content associated with objectID
	Delete(ctx context.Context, collection string, objectID string) error

	// Get gets the object associated with objectID
	Get(ctx context.Context, collection string, objectID string, opts GetOptions) (*pb.Object, error)

	// Info gets header of the object associated with objectID
	Info(ctx context.Context, collection string, objectID string) (*pb.Header, error)

	// List returns a list of at most 'opts.Count' objects
	List(ctx context.Context, collection string, opts ListOptions) (*Cursor, error)

	Search(ctx context.Context, collection string, query *pb.SearchQuery) (*Cursor, error)
}
