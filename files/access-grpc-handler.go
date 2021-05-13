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
	return &pb.CreateAccessResponse{}, CreateAccess(ctx, request.Access, CreateAccessOptions{})
}

func (s *accessServerHandler) GetAccess(ctx context.Context, request *pb.GetAccessRequest) (*pb.GetAccessResponse, error) {
	Access, err := GetAccess(ctx, request.Id, GetAccessOptions{})
	return &pb.GetAccessResponse{Access: Access}, err
}

func (s *accessServerHandler) GetAccessList(request *pb.GetAccessListRequest, server pb.AccessManager_GetAccessListServer) error {
	accesses, err := GetAccessList(server.Context(), GetAccessListOptions{})
	if err != nil {
		return err
	}

	for _, access := range accesses {
		err = server.Send(access)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *accessServerHandler) ResolveAccess(ctx context.Context, request *pb.ResolveAccessRequest) (*pb.ResolveAccessResponse, error) {
	access, err := GetAccess(ctx, request.Access.Id, GetAccessOptions{Resolved: true})
	return &pb.ResolveAccessResponse{ResolvedAccess: access}, err
}

func (s *accessServerHandler) DeleteAccess(server pb.AccessManager_DeleteAccessServer) error {
	done := false

	for !done {
		req, err := server.Recv()
		if err != nil {
			if done = err == io.EOF; !done {
				return err
			}
		}

		if req != nil {
			err = DeleteAccess(server.Context(), req.AccessId, DeleteAccessOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
