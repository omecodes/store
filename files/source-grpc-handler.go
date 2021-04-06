package files

import (
	"context"
	"io"
)

func NewSourceServerHandler() SourcesServer {
	return &sourcesServerHandler{}
}

type sourcesServerHandler struct {
	UnimplementedSourcesServer
}

func (s *sourcesServerHandler) CreateSource(ctx context.Context, request *CreateSourceRequest) (*CreateSourceResponse, error) {
	return &CreateSourceResponse{}, CreateSource(ctx, request.Source)
}

func (s *sourcesServerHandler) GetSource(ctx context.Context, request *GetSourceRequest) (*GetSourceResponse, error) {
	source, err := GetSource(ctx, request.Id)
	return &GetSourceResponse{Source: source}, err
}

func (s *sourcesServerHandler) GetSources(request *GetSourcesRequest, server Sources_GetSourcesServer) error {
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

func (s *sourcesServerHandler) ResolveSource(ctx context.Context, request *ResolveSourceRequest) (*ResolveSourceResponse, error) {
	source, err := ResolveSource(ctx, request.Source)
	return &ResolveSourceResponse{ResolvedSource: source}, err
}

func (s *sourcesServerHandler) DeleteSource(server Sources_DeleteSourceServer) error {
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
