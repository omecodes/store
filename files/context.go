package files

import "context"

type ctxRouterProvider struct{}
type ctxServiceClientProvider struct{}
type ctxSourcesServiceClientProvider struct{}
type ctxTransfersServiceClientProvider struct{}

func ContextWithClientProvider(parent context.Context, provider ClientProvider) context.Context {
	return context.WithValue(parent, ctxServiceClientProvider{}, provider)
}
func GetClientProvider(ctx context.Context) ClientProvider {
	o := ctx.Value(ctxServiceClientProvider{})
	if o == nil {
		return nil
	}
	return o.(ClientProvider)
}

func ContextWithSourcesServiceClientProvider(parent context.Context, provider SourcesServiceClientProvider) context.Context {
	return context.WithValue(parent, ctxSourcesServiceClientProvider{}, provider)
}
func GetSourcesServiceClientProvider(ctx context.Context) SourcesServiceClientProvider {
	o := ctx.Value(ctxSourcesServiceClientProvider{})
	if o == nil {
		return nil
	}
	return o.(SourcesServiceClientProvider)
}

func ContextWithTransfersServiceClientProvider(parent context.Context, provider TransfersServiceClientProvider) context.Context {
	return context.WithValue(parent, ctxTransfersServiceClientProvider{}, provider)
}
func GetTransfersServiceClientProvider(ctx context.Context) TransfersServiceClientProvider {
	o := ctx.Value(ctxTransfersServiceClientProvider{})
	if o == nil {
		return nil
	}
	return o.(TransfersServiceClientProvider)
}

func ContextWithRouterProvider(parent context.Context, provider RouterProvider) context.Context {
	return context.WithValue(parent, ctxRouterProvider{}, provider)
}
