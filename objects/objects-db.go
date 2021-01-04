package objects

import (
	"context"
	"github.com/omecodes/store/pb"
)

type Objects interface {
	// Save saves object content as JSON in database
	Save(ctx context.Context, object *pb.Object, index ...*pb.Index) error

	// Update applies patch to object with id equals patch.ID()
	Patch(ctx context.Context, patch *pb.Patch) error

	// Delete removes all content associated with objectID
	Delete(ctx context.Context, objectID string) error

	// List returns a list of at most 'opts.Count' objects
	// pass filter and
	// have CreatedAt property lower than before
	List(ctx context.Context, filter ObjectFilter, opts ListOptions) (*pb.ObjectList, error)

	// Get gets the object associated with objectID
	Get(ctx context.Context, objectID string, opts GetObjectOptions) (*pb.Object, error)

	// Info gets header of the object associated with objectID
	Info(ctx context.Context, objectID string) (*pb.Header, error)

	// Clear removes all objects store
	Clear() error
}

type ObjectFilter interface {
	Filter(o *pb.Object) (bool, error)
}

type FilterObjectFunc func(o *pb.Object) (bool, error)

func (f FilterObjectFunc) Filter(o *pb.Object) (bool, error) {
	return f(o)
}

type ObjectResolver interface {
	ResolveObject(string) (*pb.Object, error)
}

type ObjectResolveFunc func(string) (*pb.Object, error)

func (f ObjectResolveFunc) ResolveObject(id string) (*pb.Object, error) {
	return f(id)
}
