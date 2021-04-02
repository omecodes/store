package objects

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/service"
	"github.com/omecodes/store/common"
)

func AclGRPCClient(ctx context.Context) (ACLClient, error) {
	provider := ACLGrpcClientProvider(ctx)
	if provider == nil {
		return nil, errors.Internal("missing ACL provider")
	}
	return provider.GetClient(ctx)
}

func GRPCClient(ctx context.Context, serviceType uint32) (ObjectsClient, error) {
	provider := RouterGrpcClientProvider(ctx)
	if provider == nil {
		return nil, errors.ServiceUnavailable("no service available", errors.Details{Key: "type", Value: "service"}, errors.Details{Key: "service-type", Value: serviceType})
	}
	return provider.GetClient(ctx, serviceType)
}

type ClientProvider interface {
	GetClient(ctx context.Context, serviceType uint32) (ObjectsClient, error)
}

type ctxClientProvider struct{}

func WithRouterGrpcClientProvider(parent context.Context, provider ClientProvider) context.Context {
	return context.WithValue(parent, ctxClientProvider{}, provider)
}

func RouterGrpcClientProvider(ctx context.Context) ClientProvider {
	o := ctx.Value(ctxClientProvider{})
	if o == nil {
		return nil
	}
	return o.(ClientProvider)
}

type ACLClientProvider interface {
	GetClient(ctx context.Context) (ACLClient, error)
}

type ctxACLClientProvider struct{}

func WithACLGrpcClientProvider(parent context.Context, provider ACLClientProvider) context.Context {
	return context.WithValue(parent, ctxACLClientProvider{}, provider)
}

func ACLGrpcClientProvider(ctx context.Context) ACLClientProvider {
	o := ctx.Value(ctxACLClientProvider{})
	if o == nil {
		return nil
	}
	return o.(ACLClientProvider)
}

type DefaultClientProvider struct{}

func (p *DefaultClientProvider) GetClient(ctx context.Context, serviceType uint32) (ObjectsClient, error) {
	switch serviceType {
	case common.ServiceTypeObjectsHandler, common.ServiceTypeObjectsStorage:
		conn, err := service.Connect(ctx, serviceType)
		if err != nil {
			return nil, err
		}
		return NewObjectsClient(conn), nil

	default:
		return nil, errors.Unsupported("no client for this service type", errors.Details{Key: "service-type", Value: serviceType})
	}

}
