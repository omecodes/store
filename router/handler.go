package router

import (
	"context"
	"github.com/omecodes/store/pb"
)

type Handler interface {
	PutObject(ctx context.Context, object *pb.Object, security *pb.PathAccessRules, indexes []*pb.Index, opts pb.PutOptions) (string, error)
	PatchObject(ctx context.Context, patch *pb.Patch, opts pb.PatchOptions) error
	GetObject(ctx context.Context, id string, opts pb.GetOptions) (*pb.Object, error)
	GetObjectHeader(ctx context.Context, id string) (*pb.Header, error)
	DeleteObject(ctx context.Context, id string) error
	ListObjects(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error)
}
