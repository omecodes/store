package clients

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/service"
	"sync"
)

type LoadBalancer struct {
	sync.Mutex
	counter int
}

func (b *LoadBalancer) GetClient(ctx context.Context, serviceType uint32) (pb.HandlerUnitClient, error) {
	b.Lock()
	defer b.Unlock()

	registry := service.GetRegistry(ctx)
	if registry == nil {
		log.Error("Load balancer • missing registry in context")
		return nil, errors.Internal
	}

	infoList, err := registry.GetOfType(serviceType)
	if err != nil {
		return nil, err
	}

	if len(infoList) == 0 {
		return nil, errors.NotFound
	}

	var info *ome.ServiceInfo

	if len(infoList) == 0 {
		info = infoList[0]
		b.counter = 0
	} else {
		counter := b.counter % len(infoList)
		info = infoList[counter]
		b.counter = counter + 1
	}

	conn, err := service.ConnectToSpecificService(ctx, info.Id)
	if err != nil {
		return nil, err
	}

	return pb.NewHandlerUnitClient(conn), nil
}
