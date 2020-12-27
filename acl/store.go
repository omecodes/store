package acl

import (
	"context"
	"github.com/omecodes/store/pb"
)

type Store interface {
	SaveRules(ctx context.Context, objectID string, rules *pb.PathAccessRules) error
	GetRules(ctx context.Context, objectID string) (*pb.PathAccessRules, error)
	GetForPath(ctx context.Context, objectID string, path string) (*pb.AccessRules, error)
	Delete(ctx context.Context, objectID string) error
	DeleteForPath(ctx context.Context, objectID string, path string) error
}
