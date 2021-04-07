package files

import (
	"context"
	"sync"

	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/service"
)

type ClientProvider interface {
	GetClient(ctx context.Context, serviceType uint32) (FilesClient, error)
}

type DefaultClientProvider struct {
	sync.RWMutex
	balanceIndex int
}

func (p *DefaultClientProvider) incrementBalanceIndex() {
	p.Lock()
	defer p.Unlock()
	p.balanceIndex++
}

func (p *DefaultClientProvider) getBalanceIndex() int {
	p.RLock()
	defer p.Unlock()
	return p.balanceIndex
}

func (p *DefaultClientProvider) GetClient(ctx context.Context, serviceType uint32) (FilesClient, error) {
	infoList, err := service.GetRegistry(ctx).GetOfType(serviceType)
	if err != nil {
		return nil, err
	}

	defer p.incrementBalanceIndex()
	balanceIndex := p.getBalanceIndex()
	lastBalanceIndex := balanceIndex % len(infoList)

	if len(infoList) == 0 {
		return nil, errors.ServiceUnavailable("could not find service ", errors.Details{Key: "type", Value: serviceType})
	}

	if len(infoList) == 1 {
		info := infoList[0]
		conn, err := service.ConnectToSpecificService(ctx, info.Id)
		if err != nil {
			return nil, err
		}
		return NewFilesClient(conn), nil
	}

	for i := balanceIndex + 1; i%len(infoList) != lastBalanceIndex; i++ {
		info := infoList[balanceIndex%len(infoList)]
		conn, err := service.ConnectToSpecificService(ctx, info.Id)
		if err != nil {
			logs.Error("could not connect to service", logs.Details("service-id", info.Id))
			continue
		}
		return NewFilesClient(conn), nil
	}
	return nil, errors.ServiceUnavailable("could not find service ", errors.Details{Key: "type", Value: serviceType})

}

// NewClient is a FilesClient constructor
func NewClient(ctx context.Context, serviceType uint32) (FilesClient, error) {
	provider := GetClientProvider(ctx)
	if provider == nil {
		return nil, errors.ServiceUnavailable("no service available", errors.Details{Key: "type", Value: "service"}, errors.Details{Key: "service-type", Value: serviceType})
	}
	return provider.GetClient(ctx, serviceType)
}
