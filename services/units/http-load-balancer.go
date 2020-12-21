package units

import (
	"context"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
)

type httpBalancer struct{}

func (h *httpBalancer) SetSettings(ctx context.Context, name string, value string, opts oms.SettingsOptions) error {
	panic("implement me")
}

func (h *httpBalancer) GetSettings(ctx context.Context, name string) (string, error) {
	panic("implement me")
}

func (h *httpBalancer) DeleteSettings(ctx context.Context, name string) error {
	panic("implement me")
}

func (h *httpBalancer) ClearSettings(ctx context.Context) error {
	panic("implement me")
}

func (h *httpBalancer) PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	panic("implement me")
}

func (h *httpBalancer) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	panic("implement me")
}

func (h *httpBalancer) GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*oms.Object, error) {
	panic("implement me")
}

func (h *httpBalancer) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	panic("implement me")
}

func (h *httpBalancer) DeleteObject(ctx context.Context, id string) error {
	panic("implement me")
}

func (h *httpBalancer) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	panic("implement me")
}

func (h *httpBalancer) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	panic("implement me")
}
