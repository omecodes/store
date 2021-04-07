package objects

import (
	"context"
	"github.com/omecodes/service"
)

func NewDefaultACLGRPCClientProvider(serviceType uint32) ACLClientProvider {
	return &DefaultACLGrpcProvider{serviceType: serviceType}
}

type DefaultACLGrpcProvider struct {
	serviceType uint32
}

func (d *DefaultACLGrpcProvider) GetClient(ctx context.Context) (ACLClient, error) {
	conn, err := service.Connect(ctx, d.serviceType)
	if err != nil {
		return nil, err
	}
	return NewACLClient(conn), nil
}
