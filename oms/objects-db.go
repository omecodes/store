package oms

import (
	"context"
	"github.com/omecodes/store/pb"
)

type Objects interface {
	// Save saves object content as JSON in database
	Save(ctx context.Context, object *pb.Object, index ...*pb.Index) error

	// Update applies patch to object with id equals patch.ID()
	Patch(ctx context.Context, patch *Patch) error

	// Delete removes all content associated with objectID
	Delete(ctx context.Context, objectID string) error

	// List returns a list of at most 'opts.Count' objects
	// pass filter and
	// have CreatedAt property lower than before
	List(ctx context.Context, filter ObjectFilter, opts ListOptions) (*pb.ObjectList, error)

	// ListAt returns a list of at most 'opts.Count' objects
	// Each object is the value extracted at json 'path' from original json Object
	ListAt(ctx context.Context, path string, filter ObjectFilter, opts ListOptions) (*pb.ObjectList, error)

	// Get gets the object associated with objectID
	Get(ctx context.Context, objectID string) (*pb.Object, error)

	// GetPart get content at JSON-Path path in object identified by id
	GetAt(ctx context.Context, objectID string, path string) (*pb.Object, error)

	// Info gets header of the object associated with objectID
	Info(ctx context.Context, objectID string) (*pb.Header, error)

	// Clear removes all objects store
	Clear() error
}

type IDFilter interface {
	Filter(id string) (bool, error)
}

type IDFilterFunc func(id string) (bool, error)

func (f IDFilterFunc) Filter(id string) (bool, error) {
	return f(id)
}

type ObjectFilter interface {
	Filter(o *pb.Object) (bool, error)
}

type FilterObjectFunc func(o *pb.Object) (bool, error)

func (f FilterObjectFunc) Filter(o *pb.Object) (bool, error) {
	return f(o)
}
