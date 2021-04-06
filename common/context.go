package common

import (
	"context"
	ome "github.com/omecodes/libome"
)

type ctxRegistry struct{}
type ctxSettings struct{}

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

func WithRegistry(parent context.Context, registry ome.Registry) context.Context {
	return context.WithValue(parent, ctxRegistry{}, registry)
}

func Registry(ctx context.Context) ome.Registry {
	o := ctx.Value(ctxRegistry{})
	if o == nil {
		return nil
	}
	return o.(ome.Registry)
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

// WithSettings creates a context updater that adds permissions to a context
func WithSettingsContextUpdater(settings SettingsManager) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxSettings{}, settings)
	}
}
