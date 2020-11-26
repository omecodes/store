package oms

import (
	"context"
)

type Store interface {
	Save(ctx context.Context, object *Object) error
	Update(ctx context.Context, patch *Patch) error
	Delete(ctx context.Context, objectID string) error
	List(ctx context.Context, opts ListOptions) (*ListResult, error)
	Get(ctx context.Context, objectID string, opts DataOptions) (*Object, error)
	Info(ctx context.Context, objectID string) (*Info, error)
	Search(ctx context.Context, opts SearchOptions) (*SearchResult, error)
}

type GraftInfo struct {
	ID         string `json:"id"`
	DataID     string `json:"data_id"`
	CreatedBy  string `json:"created_by"`
	CreatedAt  int64  `json:"created_at"`
	Collection string `json:"collection"`
	Size       int64  `json:"size"`
}

type PutDataOptions struct{}

type PatchOptions struct{}

type DeleteOptions struct {
	Path string `json:"path"`
	Id   string `json:"id"`
}

type DataOptions struct {
	Path string `json:"path"`
}

type ListOptions struct {
	Filter ObjectFilter
	Path   string `json:"path"`
	Before int64  `json:"before"`
	Offset int    `json:"offset"`
	Count  int    `json:"count"`
}

type SearchOptions struct {
	Filter     ObjectFilter
	Before     int64  `json:"before"`
	Offset     int    `json:"offset"`
	Count      int    `json:"count"`
	Expression string `json:"expr"`
}

type SettingsOptions struct {
	Path string `json:"path"`
}

type UserOptions struct {
	WithAccessList  bool
	WithPermissions bool
	WithGroups      bool
	WithPassword    bool
}

type GetDataOptions struct {
	Path string `json:"path"`
	Info bool   `json:"info"`
}

type ListResult struct {
	Before  int64 `json:"before"`
	Offset  int   `json:"offset"`
	Count   int   `json:"count"`
	Objects []*Object
}

type SearchResult struct {
	Before  int64 `json:"before"`
	Offset  int   `json:"offset"`
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
