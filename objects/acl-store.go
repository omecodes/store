package objects

import (
	"context"
)

type ACLManager interface {
	SaveRules(ctx context.Context, collection string, objectID string, rules *PathAccessRules) error
	GetRules(ctx context.Context, collection string, objectID string) (*PathAccessRules, error)
	GetForPath(ctx context.Context, collection string, objectID string, path string) (*AccessRules, error)
	Delete(ctx context.Context, collection string, objectID string) error
	DeleteForPath(ctx context.Context, collection string, objectID string, path string) error
}
