package files

import (
	"context"
	"io"
)

type SourceService struct {
	UnimplementedSourcesServer
}

func (s *SourceService) CreateSource(ctx context.Context, request *CreateSourceRequest) (*CreateSourceResponse, error) {
	return &CreateSourceResponse{}, CreateSource(ctx, request.Source)
}

func (s *SourceService) GetSource(ctx context.Context, request *GetSourceRequest) (*GetSourceResponse, error) {
	source, err := GetSource(ctx, request.Id)
	return &GetSourceResponse{Source: source}, err
}

func (s *SourceService) GetSources(request *GetSourcesRequest, server Sources_GetSourcesServer) error {
	sources, err := ListSources(server.Context())
	if err != nil {
		return err
	}

	for _, source := range sources {
		err = server.Send(source)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SourceService) ResolveSource(ctx context.Context, request *ResolveSourceRequest) (*ResolveSourceResponse, error) {
	source, err := ResolveSource(ctx, request.Source)
	return &ResolveSourceResponse{ResolvedSource: source}, err
}

func (s *SourceService) DeleteSource(server Sources_DeleteSourceServer) error {
	done := false

	for !done {
		req, err := server.Recv()
		if err != nil {
			if done = err == io.EOF; !done {
				return err
			}
		}

		if req != nil {
			err = DeleteSource(server.Context(), req.SourceId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
