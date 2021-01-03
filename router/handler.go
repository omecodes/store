package router

import (
	"context"
	"github.com/omecodes/store/objects"
	"github.com/omecodes/store/pb"
)

type Handler interface {
	PutObject(ctx context.Context, object *pb.Object, security *pb.PathAccessRules, opts objects.PutDataOptions) (string, error)
	PatchObject(ctx context.Context, patch *pb.Patch, opts objects.PatchOptions) error
	GetObject(ctx context.Context, id string, opts objects.GetObjectOptions) (*pb.Object, error)
	GetObjectHeader(ctx context.Context, id string) (*pb.Header, error)
	DeleteObject(ctx context.Context, id string) error
	ListObjects(ctx context.Context, opts objects.ListOptions) (*pb.ObjectList, error)
	SearchObjects(ctx context.Context, params objects.SearchParams, opts objects.SearchOptions) (*pb.ObjectList, error)
}
