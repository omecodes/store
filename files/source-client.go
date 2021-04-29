package files

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/service"
	pb "github.com/omecodes/store/gen/go/proto"
	"sync"
)

type SourcesServiceClientProvider interface {
	GetClient(ctx context.Context, serviceType uint32) (pb.AccessManagerClient, error)
}

type DefaultSourcesServiceClientProvider struct {
	sync.RWMutex
	balanceIndex int
}

func (p *DefaultSourcesServiceClientProvider) incrementBalanceIndex() {
	p.Lock()
	defer p.Unlock()
	p.balanceIndex++
}

func (p *DefaultSourcesServiceClientProvider) getBalanceIndex() int {
	p.RLock()
	defer p.Unlock()
	return p.balanceIndex
}

func (p *DefaultSourcesServiceClientProvider) GetClient(ctx context.Context, serviceType uint32) (pb.AccessManagerClient, error) {
	conn, err := service.Connect(ctx, serviceType)
	if err != nil {
		return nil, err
	}
	return pb.NewAccessManagerClient(conn), nil
}

// NewSourcesServiceClient is a source service client constructor
func NewSourcesServiceClient(ctx context.Context, serviceType uint32) (pb.AccessManagerClient, error) {
	provider := GetSourcesServiceClientProvider(ctx)
	if provider == nil {
		return nil, errors.ServiceUnavailable("no service available", errors.Details{Key: "type", Value: "service"}, errors.Details{Key: "service-type", Value: serviceType})
	}
	return provider.GetClient(ctx, serviceType)
}
