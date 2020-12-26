package acl

import "context"

type ctxStore struct{}

func ContextWithStore(parent context.Context, store Store) context.Context {
	return context.WithValue(parent, ctxStore{}, store)
}

func GetStore(ctx context.Context) Store {
	o := ctx.Value(ctxStore{})
	if o == nil {
		return nil
	}
	return o.(Store)
}
