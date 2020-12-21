package units

import (
	"context"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/omestore/router"
	"sync"
)

func NewGRPCLoadBalancer() router.Handler {
	return &gRPCBalancer{
		Mutex: sync.Mutex{},
	}
}

type gRPCBalancer struct {
	sync.Mutex
}

func (g *gRPCBalancer) nextClient() (pb.HandlerUnitClient, error) {
	return nil, nil
}

func (g *gRPCBalancer) SetSettings(ctx context.Context, name string, value string, opts oms.SettingsOptions) error {
	panic("implement me")
}

func (g *gRPCBalancer) GetSettings(ctx context.Context, name string) (string, error) {
	panic("implement me")
}

func (g *gRPCBalancer) DeleteSettings(ctx context.Context, name string) error {
	panic("implement me")
}

func (g *gRPCBalancer) ClearSettings(ctx context.Context) error {
	panic("implement me")
}

func (g *gRPCBalancer) PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	panic("implement me")
}

func (g *gRPCBalancer) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	panic("implement me")
}

func (g *gRPCBalancer) GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*oms.Object, error) {
	panic("implement me")
}

func (g *gRPCBalancer) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	panic("implement me")
}

func (g *gRPCBalancer) DeleteObject(ctx context.Context, id string) error {
	panic("implement me")
}

func (g *gRPCBalancer) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	panic("implement me")
}

func (g *gRPCBalancer) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	panic("implement me")
}
