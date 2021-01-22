package accounts

import "context"

type ctxManager struct{}

func GetManager(ctx context.Context) Manager {
	o := ctx.Value(ctxManager{})
	if o == nil {
		return nil
	}
	return o.(Manager)
}
