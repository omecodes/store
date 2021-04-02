package files

import "context"

type ctxServiceClientProvider struct{}
type ctxSourcesServiceClientProvider struct{}
type ctxTransfersServiceClientProvider struct{}

func WithClientProvider(parent context.Context, provider ClientProvider) context.Context {
	return context.WithValue(parent, ctxServiceClientProvider{}, provider)
}
func GetClientProvider(ctx context.Context) ClientProvider {
	o := ctx.Value(ctxServiceClientProvider{})
	if o == nil {
		return nil
	}
	return o.(ClientProvider)
}

func WithSourcesServiceClientProvider(parent context.Context, provider SourcesServiceClientProvider) context.Context {
	return context.WithValue(parent, ctxSourcesServiceClientProvider{}, provider)
}
func GetSourcesServiceClientProvider(ctx context.Context) SourcesServiceClientProvider {
	o := ctx.Value(ctxSourcesServiceClientProvider{})
	if o == nil {
		return nil
	}
	return o.(SourcesServiceClientProvider)
}

func WithTransfersServiceClientProvider(parent context.Context, provider SourcesServiceClientProvider) context.Context {
	return context.WithValue(parent, ctxSourcesServiceClientProvider{}, provider)
}
func GetTransfersServiceClientProvider(ctx context.Context) TransfersServiceClientProvider {
	o := ctx.Value(ctxTransfersServiceClientProvider{})
	if o == nil {
		return nil
	}
	return o.(TransfersServiceClientProvider)
}
