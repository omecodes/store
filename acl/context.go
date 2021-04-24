package acl

import "context"

type ctxRelationStore struct{}
type ctxManager struct{}
type ctxNamespaceConfigStore struct{}

func relationStore(ctx context.Context) RelationStore {
	o := ctx.Value(ctxRelationStore{})
	if o == nil {
		return  nil
	}
	return o.(RelationStore)
}

func manager(ctx context.Context) Manager {
	o := ctx.Value(ctxManager{})
	if o == nil {
		return  nil
	}
	return o.(Manager)
}

func namespaceConfigStore(ctx context.Context) NamespaceConfigStore {
	o := ctx.Value(ctxNamespaceConfigStore{})
	if o == nil {
		return  nil
	}
	return o.(NamespaceConfigStore)
}