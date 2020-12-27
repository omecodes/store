package clients

import (
	"context"
	"github.com/omecodes/service"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/pb"
)

type DefaultACLGrpcProvider struct{}

func (d *DefaultACLGrpcProvider) GetClient(ctx context.Context) (pb.ACLClient, error) {
	conn, err := service.Connect(ctx, common.ServiceTypeACL)
	if err != nil {
		return nil, err
	}
	return pb.NewACLClient(conn), nil
}
