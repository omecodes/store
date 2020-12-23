package clients

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/omestore/pb"
)

type UnitProvider interface {
	GetClient(ctx context.Context, serviceType uint32) (pb.HandlerUnitClient, error)
}

func ACLStore(ctx context.Context) (pb.ACLClient, error) {
	return nil, errors.ServiceNotAvailable
}

func Unit(ctx context.Context, unitType uint32) (pb.HandlerUnitClient, error) {
	provider := UnitClientProvider(ctx)
	if provider == nil {
		return nil, errors.NotFound
	}
	return provider.GetClient(ctx, unitType)
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
