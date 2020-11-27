package router

import (
	"context"
	"github.com/omecodes/omestore/oms"
)

type Handler interface {
	SetSettings(ctx context.Context, value *oms.JSON, opts oms.SettingsOptions) error
	GetSettings(ctx context.Context, opts oms.SettingsOptions) (*oms.JSON, error)

	ListWorkers(ctx context.Context) ([]*oms.JSON, error)
	RegisterWorker(ctx context.Context, info *oms.JSON) error

	PutData(ctx context.Context, object *oms.Object, opts oms.PutDataOptions) (string, error)
	PatchData(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error
	GetData(ctx context.Context, id string, opts oms.GetDataOptions) (*oms.Object, error)
	Info(ctx context.Context, id string) (*oms.Info, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error)
	Search(ctx context.Context, opts oms.SearchOptions) (*oms.ObjectList, error)
}
