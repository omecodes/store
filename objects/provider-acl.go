package objects

import (
	"context"
	"github.com/omecodes/service"
	"github.com/omecodes/store/common"
)

type DefaultACLGrpcProvider struct{}

func (d *DefaultACLGrpcProvider) GetClient(ctx context.Context) (ACLClient, error) {
	conn, err := service.Connect(ctx, common.ServiceTypeACL)
	if err != nil {
		return nil, err
	}
	return NewACLClient(conn), nil
}
