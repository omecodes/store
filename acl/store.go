package acl

import (
	"context"
	"github.com/omecodes/store/pb"
)

type Store interface {
	SaveRules(ctx context.Context, collection string, objectID string, rules *pb.PathAccessRules) error
	GetRules(ctx context.Context, collection string, objectID string) (*pb.PathAccessRules, error)
	GetForPath(ctx context.Context, collection string, objectID string, path string) (*pb.AccessRules, error)
	Delete(ctx context.Context, collection string, objectID string) error
	DeleteForPath(ctx context.Context, collection string, objectID string, path string) error
}
