package files

import (
	"context"
	"github.com/omecodes/store/common"
)

func NewSourcesManagerServiceClient() SourceManager {
	return &sourcesManagerServiceClient{}
}

type sourcesManagerServiceClient struct{}

func (s *sourcesManagerServiceClient) Save(ctx context.Context, source *Source) (string, error) {
	client, err := NewSourcesServiceClient(ctx, common.ServiceTypeSource)
	if err != nil {
		return "", err
	}
	_, err = client.CreateSource(ctx, &CreateSourceRequest{Source: source})
	return "", err
}

func (s *sourcesManagerServiceClient) Get(ctx context.Context, id string) (*Source, error) {
	client, err := NewSourcesServiceClient(ctx, common.ServiceTypeSource)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetSource(ctx, &GetSourceRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return rsp.Source, nil
}

func (s *sourcesManagerServiceClient) Delete(ctx context.Context, id string) error {
	client, err := NewSourcesServiceClient(ctx, common.ServiceTypeSource)
	if err != nil {
		return err
	}

	stream, err := client.DeleteSource(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = stream.CloseSend()
	}()
	return stream.Send(&DeleteSourceRequest{SourceId: id})
}

func (s *sourcesManagerServiceClient) UserSources(ctx context.Context, username string) ([]*Source, error) {
	client, err := NewSourcesServiceClient(ctx, common.ServiceTypeSource)
	if err != nil {
		return nil, err
	}

	stream, err := client.GetSources(ctx, &GetSourcesRequest{User: username})
	if err != nil {
		return nil, err
	}

	var sources []*Source
	var source *Source

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
