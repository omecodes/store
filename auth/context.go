package auth

import (
	"context"
	ome "github.com/omecodes/libome"
)

type ctxAuthentication struct{}
type ctxJWt struct{}
type ctxManager struct{}
type ctxProviders struct{}

func Get(ctx context.Context) *Auth {
	o := ctx.Value(ctxAuthentication{})
	if o == nil {
		return nil
	}
	return o.(*Auth)
}

func JWT(ctx context.Context) *ome.JWT {
	o := ctx.Value(ctxJWt{})
	if o == nil {
		return nil
	}
	return o.(*ome.JWT)
}

func GetCredentialsManager(ctx context.Context) CredentialsManager {
	o := ctx.Value(ctxManager{})
	if o == nil {
		return nil
	}
	return o.(CredentialsManager)
}

func GetProviders(ctx context.Context) ProviderManager {
	o := ctx.Value(ctxProviders{})
	if o == nil {
		return nil
	}
	return o.(ProviderManager)
}

func ContextWithAuth(parent context.Context, a *Auth) context.Context {
	return context.WithValue(parent, ctxAuthentication{}, a)
}

func ContextWithProviders(parent context.Context, providers ProviderManager) context.Context {
	return context.WithValue(parent, ctxProviders{}, providers)
}

func ContextWithCredentialsManager(parent context.Context, manager CredentialsManager) context.Context {
	return context.WithValue(parent, ctxManager{}, manager)
}
