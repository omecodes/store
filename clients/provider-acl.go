package clients

import (
	"context"
	"github.com/omecodes/omestore/common"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/service"
)

type DefaultACLGrpcProvider struct{}

func (d *DefaultACLGrpcProvider) GetClient(ctx context.Context) (pb.ACLClient, error) {
	conn, err := service.Connect(ctx, common.ServiceTypeACL)
	if err != nil {
		return nil, err
	}
	return pb.NewACLClient(conn), nil
}
