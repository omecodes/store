package router

import (
	"context"
	"github.com/omecodes/store/pb"
)

type Handler interface {
	CreateCollection(ctx context.Context, collection *pb.Collection) error
	GetCollection(ctx context.Context, id string) (*pb.Collection, error)
	ListCollections(ctx context.Context) ([]*pb.Collection, error)
	DeleteCollection(ctx context.Context, id string) error

	PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.TextIndex, opts pb.PutOptions) (string, error)
	PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts pb.PatchOptions) error
	MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts pb.MoveOptions) error
	GetObject(ctx context.Context, collection string, id string, opts pb.GetOptions) (*pb.Object, error)
	GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error)
	DeleteObject(ctx context.Context, collection string, id string) error
	ListObjects(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error)
	SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*pb.Cursor, error)
}
