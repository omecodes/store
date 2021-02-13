package objects

import (
	"context"
	"github.com/google/cel-go/cel"
)

type ctxDB struct{}
type ctxACLStore struct{}
type ctxACL struct{}
type ctxSettings struct{}
type ctxCELPolicyEnv struct{}
type ctxRouterProvider struct{}

// ContextUpdater is a convenience for context enriching object
// It take a Context object and return a new one with that contains
// at least the same info as the passed one.
type ContextUpdater interface {
	UpdateContext(ctx context.Context) context.Context
}

type ContextUpdaterFunc func(ctx context.Context) context.Context

func (u ContextUpdaterFunc) UpdateContext(ctx context.Context) context.Context {
	return u(ctx)
}

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

func ContextWithACLStore(parent context.Context, store ACLManager) context.Context {
	return context.WithValue(parent, ctxACLStore{}, store)
}

func GetACLStore(ctx context.Context) ACLManager {
	o := ctx.Value(ctxACLStore{})
	if o == nil {
		return nil
	}
	return o.(ACLManager)
}

// WithObjectsStore creates a context updater that adds ACL to a context
func WithACLStoreContextUpdater(store ACLManager) ContextUpdaterFunc {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxACL{}, store)
	}
}

// WithSettings creates a context updater that adds permissions to a context
func WithSettingsContextUpdater(settings SettingsManager) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxSettings{}, settings)
	}
}

// WithRouterProvider updates context by adding a RouterProvider object in its values
func WithRouterProvider(ctx context.Context, p RouterProvider) context.Context {
	return context.WithValue(ctx, ctxRouterProvider{}, p)
}

func CELPolicyEnv(ctx context.Context) *cel.Env {
	o := ctx.Value(ctxCELPolicyEnv{})
	if o == nil {
		return nil
	}
	return o.(*cel.Env)
}

func ContextWithSettings(parent context.Context, manager SettingsManager) context.Context {
	return context.WithValue(parent, ctxSettings{}, manager)
}

func Settings(ctx context.Context) SettingsManager {
	o := ctx.Value(ctxSettings{})
	if o == nil {
		return nil
	}
	return o.(SettingsManager)
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
