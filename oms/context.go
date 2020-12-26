package oms

import "context"

type ctxStore struct{}

// WithObjectsStore creates a context updater that adds store to a context
func ContextWithStore(parent context.Context, objects Objects) context.Context {
	return context.WithValue(parent, ctxStore{}, objects)
}

func Get(ctx context.Context) Objects {
	o := ctx.Value(ctxStore{})
	if o == nil {
		return nil
	}
	return o.(Objects)
}
