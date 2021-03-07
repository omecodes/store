package objects

import (
	"context"
	"github.com/omecodes/errors"
)

func ACLGrpc(ctx context.Context) (ACLClient, error) {
	provider := ACLGrpcClientProvider(ctx)
	if provider == nil {
		return nil, errors.Internal("missing ACL provider")
	}
	return provider.GetClient(ctx)
}

func RouterGrpc(ctx context.Context, unitType uint32) (HandlerUnitClient, error) {
	provider := RouterGrpcClientProvider(ctx)
	if provider == nil {
		return nil, errors.ServiceUnavailable("no service available", errors.Details{Key: "type", Value: "service"}, errors.Details{Key: "service-type", Value: unitType})
	}
	return provider.GetClient(ctx, unitType)
}
