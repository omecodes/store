package accounts

import "context"

type ctxManager struct{}

func ContextWithManager(parent context.Context, man Manager) context.Context {
	return context.WithValue(parent, ctxManager{}, man)
}

func GetManager(ctx context.Context) Manager {
	o := ctx.Value(ctxManager{})
	if o == nil {
		return nil
	}
	return o.(Manager)
}
