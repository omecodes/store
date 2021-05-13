package files

import (
	"context"
	"github.com/omecodes/store/auth"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"
)

func NewFilesServerHandler() pb.FilesServer {
	return &gRPCHandler{}
}

type gRPCHandler struct {
	pb.UnimplementedFilesServer
	pb.UnimplementedAccessManagerServer
}

func (s *gRPCHandler) CreateAccess(ctx context.Context, request *pb.CreateAccessRequest) (*pb.CreateAccessResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.CreateAccessResponse{}, CreateAccess(ctx, request.Access, CreateAccessOptions{})
}

func (s *gRPCHandler) GetAccess(ctx context.Context, request *pb.GetAccessRequest) (*pb.GetAccessResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}
	access, err := GetAccess(ctx, request.Id, GetAccessOptions{})
	return &pb.GetAccessResponse{Access: access}, err
}

func (s *gRPCHandler) GetAccessList(_ *pb.GetAccessListRequest, server pb.AccessManager_GetAccessListServer) error {
	var err error
	ctx, err := auth.ParseMetaInNewContext(server.Context())
	if err != nil {
		return err
	}

	accesses, err := GetAccessList(ctx, GetAccessListOptions{})
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

func (s *gRPCHandler) ResolveAccess(ctx context.Context, request *pb.ResolveAccessRequest) (*pb.ResolveAccessResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}
	source, err := GetAccess(ctx, request.Access.Id, GetAccessOptions{Resolved: true})
	return &pb.ResolveAccessResponse{ResolvedAccess: source}, err
}

func (s *gRPCHandler) DeleteAccess(server pb.AccessManager_DeleteAccessServer) error {
	ctx, err := auth.ParseMetaInNewContext(server.Context())
	if err != nil {
		return err
	}
	for {
		req, err := server.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		err = DeleteAccess(ctx, req.AccessId, DeleteAccessOptions{})
		if err != nil {
			return err
		}
	}
}

func (s *gRPCHandler) CreateDir(ctx context.Context, request *pb.CreateDirRequest) (*pb.CreateDirResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.CreateDirResponse{}, CreateDir(ctx, request.AccessId, request.Path, CreateDirOptions{})
}

func (s *gRPCHandler) ListDir(ctx context.Context, request *pb.ListDirRequest) (*pb.ListDirResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	dirContent, err := ListDir(ctx, request.AccessId, request.Path, ListDirOptions{
		Offset: int(request.Offset),
		Count:  int(request.Count),
	})
	if err != nil {
		return nil, err
	}
	return &pb.ListDirResponse{
		Files:  dirContent.Files,
		Offset: uint32(dirContent.Offset),
		Total:  uint32(dirContent.Total),
	}, nil
}

func (s *gRPCHandler) GetFile(ctx context.Context, request *pb.GetFileRequest) (*pb.GetFileResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	file, err := GetFile(ctx, request.AccessId, request.Path, GetFileOptions{WithAttrs: request.WithAttributes})
	return &pb.GetFileResponse{File: file}, err
}

func (s *gRPCHandler) DeleteFile(ctx context.Context, request *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteFileResponse{}, DeleteFile(ctx, request.AccessId, request.Path, DeleteFileOptions{})
}

func (s *gRPCHandler) SetFileAttributes(ctx context.Context, request *pb.SetFileAttributesRequest) (*pb.SetFileAttributesResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.SetFileAttributesResponse{}, SetFileAttributes(ctx, request.AccessId, request.Path, request.Attributes, SetFileAttributesOptions{})
}

func (s *gRPCHandler) GetFileAttributes(ctx context.Context, request *pb.GetFileAttributesRequest) (*pb.GetFileAttributesResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	attrs, err := GetFileAttributes(ctx, request.AccessId, request.Path, request.Names, GetFileAttributesOptions{})
	return &pb.GetFileAttributesResponse{Attributes: attrs}, err
}

func (s *gRPCHandler) RenameFile(ctx context.Context, request *pb.RenameFileRequest) (*pb.RenameFileResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.RenameFileResponse{}, RenameFile(ctx, request.AccessId, request.Path, request.NewName, RenameFileOptions{})
}

func (s *gRPCHandler) MoveFile(ctx context.Context, request *pb.MoveFileRequest) (*pb.MoveFileResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.MoveFileResponse{}, MoveFile(ctx, request.AccessId, request.Path, request.TargetDir, MoveFileOptions{})
}

func (s *gRPCHandler) CopyFile(ctx context.Context, request *pb.CopyFileRequest) (*pb.CopyFileResponse, error) {
	var err error
	ctx, err = auth.ParseMetaInNewContext(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.CopyFileResponse{}, CopyFile(ctx, request.AccessId, request.Path, request.TargetDir, CopyFileOptions{})
}
