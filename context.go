package oms

import (
	"context"
	ome "github.com/omecodes/libome"
)

type ctxRegistry struct{}

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
