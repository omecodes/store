package acl

import "context"

type ctxTupleStore struct{}
type ctxManager struct{}
type ctxNamespaceConfigStore struct{}
type ctxCommitTime struct{}

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
