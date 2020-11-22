package oms

import (
	"context"
	"github.com/golang/protobuf/ptypes/any"
)

type Store interface {
	Save(ctx context.Context, data *Object) error
	Update(ctx context.Context, data *Object) error
	Delete(ctx context.Context, data *Object) error
	List(ctx context.Context, model string, opts ListOptions) (*DataList, error)
	Collections(ctx context.Context) ([]string, error)
	Get(ctx context.Context, collection string, id string, opts DataOptions) (*Object, error)
	Info(ctx context.Context, collection string, id string) (*Info, error)
	Search(ctx context.Context, collection string, condition *any.Any, opts ListOptions) (*DataList, error)

	SaveGraft(ctx context.Context, graft *Graft) (string, error)
	GetGraft(ctx context.Context, collection string, dataID string, id string) (*Graft, error)
	GraftInfo(ctx context.Context, collection string, dataID string, id string) (*GraftInfo, error)
	DeleteGraft(ctx context.Context, collection string, dataID string, id string) error
	GetAllGraft(ctx context.Context, collection string, dataID string, opts ListOptions) (*GraftList, error)
}

type Info struct {
	ID         string `json:"id"`
	CreatedBy  string `json:"created_by"`
	CreatedAt  int64  `json:"created_at"`
	Collection string `json:"collection"`
	Size       int64  `json:"size"`
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

type PatchOptions struct {
	Path string `json:"path"`
}

type DeleteOptions struct {
	Path string `json:"path"`
	Id   string `json:"id"`
}

type DataOptions struct {
	Path string `json:"path"`
}

type ListOptions struct {
	IDFilter IDFilter
	Path     string `json:"path"`
	Before   int64  `json:"before"`
	Count    int64  `json:"count"`
}

type BulkOptions struct {
	Provider IDFilter
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
}

type DataList struct {
	Collection string `json:"collection"`
	Cursor     DataCursor
}

type GraftList struct {
	Collection string `json:"collection"`
	DataID     string `json:"data_id"`
	Cursor     GraftCursor
}

type IDFilter interface {
	Filter(id string) (bool, error)
}

type IDFilterFunc func(id string) (bool, error)

func (f IDFilterFunc) Filter(id string) (bool, error) {
	return f(id)
}
