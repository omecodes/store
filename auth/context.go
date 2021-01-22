package auth

import (
	"context"
	ome "github.com/omecodes/libome"
)

type ctxUser struct{}
type ctxJWt struct{}
type ctxCredentialsManager struct{}
type ctxAuthenticationProviders struct{}

func Get(ctx context.Context) *User {
	o := ctx.Value(ctxUser{})
	if o == nil {
		return nil
	}
	return o.(*User)
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
