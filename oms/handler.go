package oms

import (
	"context"
	"io"
)

type Handler interface {
	SetSettings(ctx context.Context, value *JSON, opts SettingsOptions) error
	GetSettings(ctx context.Context, opts SettingsOptions) (*JSON, error)

	ListWorkers(ctx context.Context) ([]*JSON, error)
	RegisterWorker(ctx context.Context, info *JSON) error

	GetCollections(ctx context.Context) ([]string, error)

	PutData(ctx context.Context, data *Object, opts PutDataOptions) error
	PatchData(ctx context.Context, collection string, id string, content io.Reader, size int64, opts PatchOptions) error
	GetData(ctx context.Context, collection string, id string, opts GetDataOptions) (*Object, error)
	Info(ctx context.Context, collection string, id string) (*Info, error)
	Delete(ctx context.Context, collection string, id string) error
	List(ctx context.Context, collection string, opts ListOptions) (*DataList, error)

	SaveGraft(ctx context.Context, graft *Graft) (string, error)
	GetGraft(ctx context.Context, collection string, dataID string, id string) (*Graft, error)
	GraftInfo(ctx context.Context, collection string, dataID string, id string) (*GraftInfo, error)
	ListGrafts(ctx context.Context, collection string, dataID string, opts ListOptions) (*GraftList, error)
	DeleteGraft(ctx context.Context, collection string, dataID string, id string) error
}
