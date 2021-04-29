package auth

import (
	"context"
	ome "github.com/omecodes/libome"
	pb "github.com/omecodes/store/gen/go/proto"
)

type ctxUser struct{}
type ctxApp struct{}
type ctxJWt struct{}
type ctxCredentialsManager struct{}
type ctxAuthenticationProviders struct{}

func ContextWithUser(parent context.Context, user *pb.User) context.Context {
	return context.WithValue(parent, ctxUser{}, user)
}

func ContextWithApp(parent context.Context, clientApp *pb.ClientApp) context.Context {
	return context.WithValue(parent, ctxApp{}, clientApp)
}

func Get(ctx context.Context) *pb.User {
	o := ctx.Value(ctxUser{})
	if o == nil {
		return nil
	}
	return o.(*pb.User)
}

func App(ctx context.Context) *pb.ClientApp {
	o := ctx.Value(ctxApp{})
	if o == nil {
		return nil
	}
	return o.(*pb.ClientApp)
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
