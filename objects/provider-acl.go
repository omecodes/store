package objects

import (
	"context"
	"github.com/omecodes/service"
)

type DefaultACLGrpcProvider struct{}

func (d *DefaultACLGrpcProvider) GetClient(ctx context.Context) (ACLClient, error) {
	conn, err := service.Connect(ctx, ServiceTypeACL)
	if err != nil {
		return nil, err
	}
	return NewACLClient(conn), nil
}
