package clients

import (
	"context"
	"github.com/omecodes/omestore/common"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/service"
)

type StoreProvider struct{}

func (p *StoreProvider) GetClient(ctx context.Context, serviceType uint32) (pb.HandlerUnitClient, error) {
	conn, err := service.Connect(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}
	return pb.NewHandlerUnitClient(conn), nil
}
