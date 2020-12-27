package clients

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/store/pb"
)

func ACLGrpc(ctx context.Context) (pb.ACLClient, error) {
	provider := ACLGrpcClientProvider(ctx)
	if provider == nil {
		return nil, errors.NotFound
	}
	return provider.GetClient(ctx)
}

func RouterGrpc(ctx context.Context, unitType uint32) (pb.HandlerUnitClient, error) {
	provider := RouterGrpcClientProvider(ctx)
	if provider == nil {
		return nil, errors.NotFound
	}
	return provider.GetClient(ctx, unitType)
}
