package files

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
)

func NewSourcesManagerServiceClient(serviceType uint32) AccessManager {
	return &accessManagerServiceClient{
		serviceType: serviceType,
	}
}

type accessManagerServiceClient struct {
	serviceType uint32
}

func (s *accessManagerServiceClient) Save(ctx context.Context, source *pb.FSAccess) (string, error) {
	client, err := NewSourcesServiceClient(ctx, s.serviceType)
	if err != nil {
		return "", err
	}
	_, err = client.CreateAccess(ctx, &pb.CreateAccessRequest{Access: source})
	return "", err
}

func (s *accessManagerServiceClient) Get(ctx context.Context, id string) (*pb.FSAccess, error) {
	client, err := NewSourcesServiceClient(ctx, s.serviceType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetAccess(ctx, &pb.GetAccessRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return rsp.Access, nil
}

func (s *accessManagerServiceClient) Delete(ctx context.Context, id string) error {
	client, err := NewSourcesServiceClient(ctx, s.serviceType)
	if err != nil {
		return err
	}

	stream, err := client.DeleteAccess(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = stream.CloseSend()
	}()
	return stream.Send(&pb.DeleteAccessRequest{AccessId: id})
}

func (s *accessManagerServiceClient) UserSources(ctx context.Context, username string) ([]*pb.FSAccess, error) {
	client, err := NewSourcesServiceClient(ctx, s.serviceType)
	if err != nil {
		return nil, err
	}

	stream, err := client.GetAccessList(ctx, &pb.GetAccessListRequest{User: username})
	if err != nil {
		return nil, err
	}

	var sources []*pb.FSAccess
	var source *pb.FSAccess

	for {
		source, err = stream.Recv()
		if source != nil {
			sources = append(sources, source)
		}
		if err != nil {
			break
		}
	}
	return sources, err
}
