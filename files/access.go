package files

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
)

const (
	SchemeFS  = "files"
	SchemeRef = "ref"
)

type ctxAccessManager struct{}

type AccessManager interface {
	Save(ctx context.Context, access *pb.FSAccess) (string, error)
	Get(ctx context.Context, accessID string) (*pb.FSAccess, error)
	GetResolved(ctx context.Context, accessID string) (*pb.FSAccess, error)
	Delete(ctx context.Context, accessID string) error
}

func ContextWithAccessManager(parent context.Context, manager AccessManager) context.Context {
	return context.WithValue(parent, ctxAccessManager{}, manager)
}

func getAccessManager(ctx context.Context) AccessManager {
	o := ctx.Value(ctxAccessManager{})
	if o == nil {
		return nil
	}
	return o.(AccessManager)
}
