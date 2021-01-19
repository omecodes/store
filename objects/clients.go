package objects

import (
	"context"
	"github.com/omecodes/common/errors"
)

func ACLGrpc(ctx context.Context) (ACLClient, error) {
	provider := ACLGrpcClientProvider(ctx)
	if provider == nil {
		return nil, errors.NotFound
	}
	return provider.GetClient(ctx)
}

func RouterGrpc(ctx context.Context, unitType uint32) (HandlerUnitClient, error) {
	provider := RouterGrpcClientProvider(ctx)
	if provider == nil {
		return nil, errors.NotFound
	}
	return provider.GetClient(ctx, unitType)
}
