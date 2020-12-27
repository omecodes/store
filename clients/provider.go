package clients

import (
	"context"
	"github.com/omecodes/store/pb"
)

type GRPCRouterProvider interface {
	GetClient(ctx context.Context, serviceType uint32) (pb.HandlerUnitClient, error)
}

type ctxGrpcRouterClientProvider struct{}

func WithRouterGrpcClientProvider(parent context.Context, provider GRPCRouterProvider) context.Context {
	return context.WithValue(parent, ctxGrpcRouterClientProvider{}, provider)
}

func RouterGrpcClientProvider(ctx context.Context) GRPCRouterProvider {
	o := ctx.Value(ctxGrpcRouterClientProvider{})
	if o == nil {
		return nil
	}
	return o.(GRPCRouterProvider)
}

type GRPCACLProvider interface {
	GetClient(ctx context.Context) (pb.ACLClient, error)
}

type ctxGrpcACLClientProvider struct{}

func WithACLGrpcClientProvider(parent context.Context, provider GRPCACLProvider) context.Context {
	return context.WithValue(parent, ctxGrpcACLClientProvider{}, provider)
}

func ACLGrpcClientProvider(ctx context.Context) GRPCACLProvider {
	o := ctx.Value(ctxGrpcACLClientProvider{})
	if o == nil {
		return nil
	}
	return o.(GRPCACLProvider)
}
