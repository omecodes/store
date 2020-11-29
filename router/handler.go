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

	PutObject(ctx context.Context, object *oms.Object, security *oms.PathAccessRules, opts oms.PutDataOptions) (string, error)
	PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error
	GetObject(ctx context.Context, id string, opts oms.GetDataOptions) (*oms.Object, error)
	GetObjectHeader(ctx context.Context, id string) (*oms.Header, error)
	DeleteObject(ctx context.Context, id string) error
	ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error)
	SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error)
}
