package objects

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/service"
)

type DefaultRouterGrpcProvider struct{}

func (p *DefaultRouterGrpcProvider) GetClient(ctx context.Context, serviceType uint32) (HandlerUnitClient, error) {
	switch serviceType {
	case ServiceTypeObjects:
		conn, err := service.Connect(ctx, ServiceTypeObjects)
		if err != nil {
			return nil, err
		}
		return NewHandlerUnitClient(conn), nil

	default:
		return nil, errors.Unsupported("no client for this service type", errors.Details{Key: "service-type", Value: serviceType})
	}

}
