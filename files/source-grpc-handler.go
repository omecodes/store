package files

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"
)

func NewAccessServerHandler() pb.AccessManagerServer {
	return &accessServerHandler{}
}

type accessServerHandler struct {
	pb.UnimplementedAccessManagerServer
}

func (s *accessServerHandler) CreateAccess(ctx context.Context, request *pb.CreateAccessRequest) (*pb.CreateAccessResponse, error) {
	return &pb.CreateAccessResponse{}, CreateSource(ctx, request.Access)
}

func (s *accessServerHandler) GetSource(ctx context.Context, request *pb.GetAccessRequest) (*pb.GetAccessResponse, error) {
	Access, err := GetSource(ctx, request.Id)
	return &pb.GetAccessResponse{Access: Access}, err
}

func (s *accessServerHandler) GetSources(request *pb.GetAccessListRequest, server pb.AccessManager_GetAccessListServer) error {
	sources, err := ListSources(server.Context())
	if err != nil {
		return err
	}

	for _, Access := range sources {
		err = server.Send(Access)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *accessServerHandler) ResolveSource(ctx context.Context, request *pb.ResolveAccessRequest) (*pb.ResolveAccessResponse, error) {
	access, err := ResolveSource(ctx, request.Access)
	return &pb.ResolveAccessResponse{ResolvedAccess: access}, err
}

func (s *accessServerHandler) DeleteSource(server pb.AccessManager_DeleteAccessServer) error {
	done := false

	for !done {
		req, err := server.Recv()
		if err != nil {
			if done = err == io.EOF; !done {
				return err
			}
		}

		if req != nil {
			err = DeleteSource(server.Context(), req.AccessId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
