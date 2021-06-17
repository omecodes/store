package acl

import (
	"context"
)

type ctxTupleStore struct{}
type ctxManager struct{}
type ctxNamespaceConfigStore struct{}
type ctxStateMinAge struct{}
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

func getStateMinAge(ctx context.Context) int64 {
	o := ctx.Value(ctxStateMinAge{})
	if o == nil {
		return 0
	}
	return o.(int64)
}

func ContextWithManager(parent context.Context, man Manager) context.Context {
	return context.WithValue(parent, ctxManager{}, man)
}

func ContextWithTupleStore(parent context.Context, store TupleStore) context.Context {
	return context.WithValue(parent, ctxTupleStore{}, store)
}

func ContextWithNamespaceConfigStore(parent context.Context, store NamespaceConfigStore) context.Context {
	return context.WithValue(parent, ctxNamespaceConfigStore{}, store)
}

// WithRouterContextUpdaterFunc return a function that updates context with router provider

// ContextWithRouterProvider creates a new context that contains ctx values in addition of the passed router provider

func GetHandler(ctx context.Context, opt ...RouteOption) Handler {
	o := ctx.Value(ctxRouterProvider{})
	if o != nil {
		p := o.(RouterProvider)
		router := p.GetRouter(ctx)
		return router.GetHandler(opt...)
	}
	return DefaultRouter().GetHandler(opt...)
}
