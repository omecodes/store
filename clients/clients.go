package clients

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/omestore/pb"
)

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
