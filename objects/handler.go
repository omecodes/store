package objects

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
)

type Handler interface {
	CreateCollection(ctx context.Context, collection *pb.Collection, opts CreateCollectionOptions) error
	GetCollection(ctx context.Context, id string, opts GetCollectionOptions) (*pb.Collection, error)
	ListCollections(ctx context.Context, opts ListCollectionOptions) ([]*pb.Collection, error)
	DeleteCollection(ctx context.Context, id string, opts DeleteCollectionOptions) error

	PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.TextIndex, opts PutOptions) (string, error)
	PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts PatchOptions) error
	MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts MoveOptions) error
	GetObject(ctx context.Context, collection string, id string, opts GetObjectOptions) (*pb.Object, error)
	GetObjectHeader(ctx context.Context, collection string, id string, opts GetHeaderOptions) (*pb.Header, error)
	DeleteObject(ctx context.Context, collection string, id string, opts DeleteObjectOptions) error
	ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error)
	SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery, opts SearchObjectsOptions) (*Cursor, error)
}

func CreateCollection(ctx context.Context, collection *pb.Collection, opts CreateCollectionOptions) error {
	return GetRouterHandler(ctx).CreateCollection(ctx, collection, opts)
}

func GetCollection(ctx context.Context, id string, opts GetCollectionOptions) (*pb.Collection, error) {
	return GetRouterHandler(ctx).GetCollection(ctx, id, opts)
}

func ListCollections(ctx context.Context, opts ListCollectionOptions) ([]*pb.Collection, error) {
	return GetRouterHandler(ctx).ListCollections(ctx, opts)
}

func DeleteCollection(ctx context.Context, id string, opts DeleteCollectionOptions) error {
	return GetRouterHandler(ctx).DeleteCollection(ctx, id, opts)
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

func GetObject(ctx context.Context, collection string, id string, opts GetObjectOptions) (*pb.Object, error) {
	return GetRouterHandler(ctx).GetObject(ctx, collection, id, opts)
}

func GetObjectHeader(ctx context.Context, collection string, id string, opts GetHeaderOptions) (*pb.Header, error) {
	return GetRouterHandler(ctx).GetObjectHeader(ctx, collection, id, opts)
}

func DeleteObject(ctx context.Context, collection string, id string, opts DeleteObjectOptions) error {
	return GetRouterHandler(ctx).DeleteObject(ctx, collection, id, opts)
}

func ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	return GetRouterHandler(ctx).ListObjects(ctx, collection, opts)
}

func SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery, opts SearchObjectsOptions) (*Cursor, error) {
	return GetRouterHandler(ctx).SearchObjects(ctx, collection, query, opts)
}
