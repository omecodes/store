package acl

import (
	"context"
	ome "github.com/omecodes/libome"
)

type ctxTupleStore struct{}
type ctxManager struct{}
type ctxNamespaceConfigStore struct{}
type ctxCommitTime struct{}
type ctxRouterProvider struct{}

func getTupleStore(ctx context.Context) TupleStore {
	o := ctx.Value(ctxTupleStore{})
	if o == nil {
		return nil
	}
	return o.(TupleStore)
}

func GetManager(ctx context.Context) Manager {
	o := ctx.Value(ctxManager{})
	if o == nil {
		return nil
	}
	return o.(Manager)
}

func getNamespaceConfigStore(ctx context.Context) NamespaceConfigStore {
	o := ctx.Value(ctxNamespaceConfigStore{})
	if o == nil {
		return nil
	}
	return o.(NamespaceConfigStore)
}

func getCommitTime(ctx context.Context) int64 {
	o := ctx.Value(ctxCommitTime{})
	if o == nil {
		return 0
	}
	return o.(int64)
}

func ContextWithManager(parent context.Context, man Manager) context.Context {
	return context.WithValue(parent, ctxManager{}, man)
}

// WithRouterContextUpdaterFunc return a function that updates context with router provider
func WithRouterContextUpdaterFunc(provider RouterProvider) ome.GrpcContextUpdater {
	return ome.GrpcContextUpdaterFunc(func(ctx context.Context) (context.Context, error) {
		return context.WithValue(ctx, ctxRouterProvider{}, provider), nil
	})
}

// ContextWithRouterProvider creates a new context that contains ctx values in addition of the passed router provider
func ContextWithRouterProvider(ctx context.Context, p RouterProvider) context.Context {
	return context.WithValue(ctx, ctxRouterProvider{}, p)
}

func GetHandler(ctx context.Context, opt ...RouteOption) Handler {
	o := ctx.Value(ctxRouterProvider{})
	if o != nil {
		p := o.(RouterProvider)
		router := p.GetRouter(ctx)
		return router.GetHandler(opt...)
	}
	return DefaultRouter().GetHandler(opt...)
}
