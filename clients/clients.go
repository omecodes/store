package clients

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/omestore/pb"
)

func ACLStore(ctx context.Context) (pb.ACLClient, error) {
	return nil, errors.ServiceNotAvailable
}

func Handler(ctx context.Context) (pb.HandlerUnitClient, error) {
	return nil, errors.ServiceNotAvailable
}

func Objects(ctx context.Context) (pb.HandlerUnitClient, error) {
	return nil, errors.ServiceNotAvailable
}
