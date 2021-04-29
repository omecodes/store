package objects

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
)

type Handler interface {
	CreateCollection(ctx context.Context, collection *pb.Collection) error
	GetCollection(ctx context.Context, id string) (*pb.Collection, error)
	ListCollections(ctx context.Context) ([]*pb.Collection, error)
	DeleteCollection(ctx context.Context, id string) error

	PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.TextIndex, opts PutOptions) (string, error)
	PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts PatchOptions) error
	MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts MoveOptions) error
	GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*pb.Object, error)
	GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error)
	DeleteObject(ctx context.Context, collection string, id string) error
	ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error)
	SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*Cursor, error)
}

func CreateCollection(ctx context.Context, collection *pb.Collection) error {
	return GetRouterHandler(ctx).CreateCollection(ctx, collection)
}

func GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	return GetRouterHandler(ctx).GetCollection(ctx, id)
}

func ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	return GetRouterHandler(ctx).ListCollections(ctx)
}

func DeleteCollection(ctx context.Context, id string) error {
	return GetRouterHandler(ctx).DeleteCollection(ctx, id)
}

func PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.TextIndex, opts PutOptions) (string, error) {
	return GetRouterHandler(ctx).PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts PatchOptions) error {
	return GetRouterHandler(ctx).PatchObject(ctx, collection, patch, opts)
}

func MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts MoveOptions) error {
	return GetRouterHandler(ctx).MoveObject(ctx, collection, objectID, targetCollection, accessSecurityRules, opts)
}

func GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*pb.Object, error) {
	return GetRouterHandler(ctx).GetObject(ctx, collection, id, opts)
}

func GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error) {
	return GetRouterHandler(ctx).GetObjectHeader(ctx, collection, id)
}

func DeleteObject(ctx context.Context, collection string, id string) error {
	return GetRouterHandler(ctx).DeleteObject(ctx, collection, id)
}

func ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	return GetRouterHandler(ctx).ListObjects(ctx, collection, opts)
}

func SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*Cursor, error) {
	return GetRouterHandler(ctx).SearchObjects(ctx, collection, query)
}
