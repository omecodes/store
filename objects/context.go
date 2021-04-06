package objects

import (
	"context"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/store/common"
)

type ctxDB struct{}
type ctxACLStore struct{}
type ctxACL struct{}
type ctxCELPolicyEnv struct{}
type ctxRouterProvider struct{}

// WithObjectsStore creates a context updater that adds store to a context
func ContextWithStore(parent context.Context, db DB) context.Context {
	return context.WithValue(parent, ctxDB{}, db)
}

func Get(ctx context.Context) DB {
	o := ctx.Value(ctxDB{})
	if o == nil {
		return nil
	}
	return o.(DB)
}

func ContextWithACLManager(parent context.Context, store ACLManager) context.Context {
	return context.WithValue(parent, ctxACLStore{}, store)
}

func GetACLStore(ctx context.Context) ACLManager {
	o := ctx.Value(ctxACLStore{})
	if o == nil {
		return nil
	}
	return o.(ACLManager)
}

func WithRouterProviderContextUpdater(provider RouterProvider) ome.GrpcContextUpdater {
	return ome.GrpcContextUpdaterFunc(func(ctx context.Context) (context.Context, error) {
		return context.WithValue(ctx, ctxRouterProvider{}, provider), nil
	})
}

// WithObjectsStore creates a context updater that adds ACL to a context
func WithACLStoreContextUpdater(store ACLManager) common.ContextUpdaterFunc {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxACL{}, store)
	}
}

// ContextWithRouterProvider updates context by adding a RouterProvider object in its values
func ContextWithRouterProvider(ctx context.Context, p RouterProvider) context.Context {
	return context.WithValue(ctx, ctxRouterProvider{}, p)
}

func GetRouterHandler(ctx context.Context, opt ...RouteOption) Handler {
	o := ctx.Value(ctxRouterProvider{})
	if o != nil {
		p := o.(RouterProvider)
		router := p.GetRouter(ctx)
		return router.GetHandler(opt...)
	}
	return DefaultRouter().GetHandler(opt...)
}
