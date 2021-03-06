package objects

import (
	"context"
	"github.com/omecodes/errors"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/service"
	pb "github.com/omecodes/store/gen/go/proto"
	"sync"
)

type LoadBalancer struct {
	sync.Mutex
	counter int
}

func (b *LoadBalancer) GetClient(ctx context.Context, serviceType uint32) (pb.ObjectsClient, error) {
	b.Lock()
	defer b.Unlock()

	registry := service.GetRegistry(ctx)
	if registry == nil {
		logs.Error("Load balancer • missing registry in context")
		return nil, errors.Internal("missing service registry")
	}

	infoList, err := registry.GetOfType(serviceType)
	if err != nil {
		return nil, err
	}

	if len(infoList) == 0 {
		return nil, errors.NotFound("no service found")
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

	return pb.NewObjectsClient(conn), nil
}
