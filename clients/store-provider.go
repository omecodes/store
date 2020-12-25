package clients

import (
	"context"
	"github.com/omecodes/libome/errors"
	"github.com/omecodes/omestore/auth"
	"github.com/omecodes/omestore/common"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/service"
)

type StoreProvider struct{}

func (p *StoreProvider) GetClient(ctx context.Context, serviceType uint32) (pb.HandlerUnitClient, error) {
	switch serviceType {
	case common.ServiceTypeObjects:
		conn, err := service.Connect(auth.SetMetaWithExisting(ctx), common.ServiceTypeObjects)
		if err != nil {
			return nil, err
		}
		return pb.NewHandlerUnitClient(conn), nil

	default:
		return nil, errors.New(errors.CodeNotImplemented, "no client for this service type")
	}

}
