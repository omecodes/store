package objects

import (
	"context"
	"github.com/omecodes/errors"
	pb "github.com/omecodes/store/gen/go/proto"
)

func grpcClient(ctx context.Context, serviceType uint32) (pb.ObjectsClient, error) {
	provider := RouterGrpcClientProvider(ctx)
	if provider == nil {
		return nil, errors.ServiceUnavailable("no service available", errors.Details{Key: "type", Value: "service"}, errors.Details{Key: "service-type", Value: serviceType})
	}
	return provider.GetClient(ctx, serviceType)
}

type ClientProvider interface {
	GetClient(ctx context.Context, serviceType uint32) (pb.ObjectsClient, error)
}

type ctxClientProvider struct{}

func WithObjectsGrpcClientProvider(parent context.Context, provider ClientProvider) context.Context {
	return context.WithValue(parent, ctxClientProvider{}, provider)
}

func RouterGrpcClientProvider(ctx context.Context) ClientProvider {
	o := ctx.Value(ctxClientProvider{})
	if o == nil {
		return nil
	}
	return o.(ClientProvider)
}
