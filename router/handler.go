package router

import (
	"context"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
)

type Handler interface {
	SetSettings(ctx context.Context, name string, value string, opts oms.SettingsOptions) error
	GetSettings(ctx context.Context, name string) (string, error)
	DeleteSettings(ctx context.Context, name string) error
	ClearSettings(ctx context.Context) error

	ListWorkers(ctx context.Context) ([]*oms.JSON, error)
	RegisterWorker(ctx context.Context, info *oms.JSON) error

	PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error)
	PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error
	GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*oms.Object, error)
	GetObjectHeader(ctx context.Context, id string) (*pb.Header, error)
	DeleteObject(ctx context.Context, id string) error
	ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error)
	SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error)
}
