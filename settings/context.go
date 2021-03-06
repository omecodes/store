package settings

import (
	"context"
)

type ctxSettings struct{}

func ContextWithManager(parent context.Context, manager Manager) context.Context {
	return context.WithValue(parent, ctxSettings{}, manager)
}

func GetManager(ctx context.Context) Manager {
	o := ctx.Value(ctxSettings{})
	if o == nil {
		return nil
	}
	return o.(Manager)
}

// WithSettingsContextUpdater creates a context updater that adds permissions to a context
