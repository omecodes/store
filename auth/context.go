package auth

import (
	"context"
	ome "github.com/omecodes/libome"
)

type ctxUser struct{}
type ctxApp struct{}
type ctxJWt struct{}
type ctxCredentialsManager struct{}
type ctxAuthenticationProviders struct{}

func ContextWithUser(parent context.Context, user *User) context.Context {
	return context.WithValue(parent, ctxUser{}, user)
}

func ContextWithApp(parent context.Context, clientApp *ClientApp) context.Context {
	return context.WithValue(parent, ctxApp{}, clientApp)
}

func Get(ctx context.Context) *User {
	o := ctx.Value(ctxUser{})
	if o == nil {
		return nil
	}
	return o.(*User)
}

func App(ctx context.Context) *ClientApp {
	o := ctx.Value(ctxApp{})
	if o == nil {
		return nil
	}
	return o.(*ClientApp)
}

func JWT(ctx context.Context) *ome.JWT {
	o := ctx.Value(ctxJWt{})
	if o == nil {
		return nil
	}
	return o.(*ome.JWT)
}

func GetCredentialsManager(ctx context.Context) CredentialsManager {
	o := ctx.Value(ctxCredentialsManager{})
	if o == nil {
		return nil
	}
	return o.(CredentialsManager)
}

func GetProviders(ctx context.Context) ProviderManager {
	o := ctx.Value(ctxAuthenticationProviders{})
	if o == nil {
		return nil
	}
	return o.(ProviderManager)
}
