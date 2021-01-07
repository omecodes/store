package objects

import (
	"context"
	"github.com/omecodes/store/pb"
)

type Collection interface {

	// Save saves object content as JSON in database
	Save(ctx context.Context, object *pb.Object, index ...*pb.Index) error

	// Update applies patch to object with id equals patch.ID()
	Patch(ctx context.Context, patch *pb.Patch) error

	// Delete removes all content associated with objectID
	Delete(ctx context.Context, objectID string) error

	// List returns a list of at most 'opts.Count' objects
	List(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error)

	// Get gets the object associated with objectID
	Get(ctx context.Context, objectID string, opts pb.GetOptions) (*pb.Object, error)

	// Info gets header of the object associated with objectID
	Info(ctx context.Context, objectID string) (*pb.Header, error)

	// Clear removes all objects store
	Clear() error
}

type CollectionItem struct {
	Id   string `json:"id"`
	Date int64  `json:"date"`
	Data string `json:"data"`
}
