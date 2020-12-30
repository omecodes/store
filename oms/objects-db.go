package oms

import (
	"context"
	"github.com/omecodes/store/pb"
)

type Objects interface {
	// Save saves object content as JSON in database
	Save(ctx context.Context, object *Object, index ...*pb.Index) error

	// Update applies patch to object with id equals patch.ID()
	Patch(ctx context.Context, patch *Patch) error

	// Delete removes all content associated with objectID
	Delete(ctx context.Context, objectID string) error

	// List returns a list of objects long of at most opts.Count
	// pass filter and
	// have CreatedAt property lower than before
	List(ctx context.Context, before int64, count int, filter ObjectFilter) (*ObjectList, error)

	// ListAt returns a at most opts.Count list sized of objects value at path
	// pass opts.Filter and
	// have CreatedAt property lower than opts.Before
	ListAt(ctx context.Context, path string, before int64, count int, filter ObjectFilter) (*ObjectList, error)

	// Get gets the object associated with objectID
	Get(ctx context.Context, objectID string) (*Object, error)

	// GetPart get content at JSON-Path path in object identified by id
	GetAt(ctx context.Context, objectID string, path string) (*Object, error)

	// Info gets header of the object associated with objectID
	Info(ctx context.Context, objectID string) (*pb.Header, error)

	// Clear removes all objects store
	Clear() error
}

type GraftInfo struct {
	ID         string `json:"id"`
	DataID     string `json:"data_id"`
	CreatedBy  string `json:"created_by"`
	CreatedAt  int64  `json:"created_at"`
	Collection string `json:"collection"`
	Size       int64  `json:"size"`
}

type PutDataOptions struct {
	Indexes []*pb.Index
}

type PatchOptions struct {
	Indexes []*pb.Index
}

type DeleteOptions struct {
	Path string `json:"path"`
}

type DataOptions struct {
	Path string `json:"path"`
}

type ListOptions struct {
	Filter ObjectFilter
	Path   string `json:"path"`
	Before int64  `json:"before"`
	Count  int    `json:"count"`
}

type SearchParams struct {
	Collection        string `json:"collection"`
	MatchedExpression string `json:"matched_expression"`
}

type SearchOptions struct {
	Filter ObjectFilter
	Path   string `json:"path"`
	Before int64  `json:"before"`
	Count  int    `json:"count"`
}

type SettingsOptions struct{}

type UserOptions struct {
	WithAccessList  bool
	WithPermissions bool
	WithGroups      bool
	WithPassword    bool
}

type GetObjectOptions struct {
	Path string `json:"path"`
	Info bool   `json:"header"`
}

type ObjectList struct {
	Before  int64 `json:"before"`
	Count   int   `json:"count"`
	Objects []*Object
}

type IDFilter interface {
	Filter(id string) (bool, error)
}

type IDFilterFunc func(id string) (bool, error)

func (f IDFilterFunc) Filter(id string) (bool, error) {
	return f(id)
}

type ObjectFilter interface {
	Filter(o *Object) (bool, error)
}

type FilterObjectFunc func(o *Object) (bool, error)

func (f FilterObjectFunc) Filter(o *Object) (bool, error) {
	return f(o)
}
