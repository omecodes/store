package clients

import (
	"context"
	"github.com/omecodes/omestore/pb"
)

type UnitProvider interface {
	GetClient(ctx context.Context, serviceType uint32) (pb.HandlerUnitClient, error)
}

type ctxUnitClientProvider struct{}

func WithUnitClientProvider(parent context.Context, provider UnitProvider) context.Context {
	return context.WithValue(parent, ctxUnitClientProvider{}, provider)
}

func UnitClientProvider(ctx context.Context) UnitProvider {
	o := ctx.Value(ctxUnitClientProvider{})
	if o == nil {
		return nil
	}
	return o.(UnitProvider)
}
