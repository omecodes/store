package files

import (
	"context"
	"io"
)

type gRPCHandler struct {
	UnimplementedFilesServer
	UnimplementedSourcesServer
}

func (s *gRPCHandler) CreateSource(ctx context.Context, request *CreateSourceRequest) (*CreateSourceResponse, error) {
	return &CreateSourceResponse{}, CreateSource(ctx, request.Source)
}

func (s *gRPCHandler) GetSource(ctx context.Context, request *GetSourceRequest) (*GetSourceResponse, error) {
	source, err := GetSource(ctx, request.Id)
	return &GetSourceResponse{Source: source}, err
}

func (s *gRPCHandler) GetSources(request *GetSourcesRequest, server Sources_GetSourcesServer) error {
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

func (s *gRPCHandler) ResolveSource(ctx context.Context, request *ResolveSourceRequest) (*ResolveSourceResponse, error) {
	source, err := ResolveSource(ctx, request.Source)
	return &ResolveSourceResponse{ResolvedSource: source}, err
}

func (s *gRPCHandler) DeleteSource(server Sources_DeleteSourceServer) error {
	for {
		req, err := server.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		err = DeleteSource(server.Context(), req.SourceId)
		if err != nil {
			return err
		}
	}
}

func (s *gRPCHandler) CreateDir(ctx context.Context, request *CreateDirRequest) (*CreateDirResponse, error) {
	return &CreateDirResponse{}, CreateDir(ctx, request.SourceId, request.Path)
}

func (s *gRPCHandler) ListDir(ctx context.Context, request *ListDirRequest) (*ListDirResponse, error) {
	dirContent, err := ListDir(ctx, request.SourceId, request.Path, ListDirOptions{
		Offset: int(request.Offset),
		Count:  int(request.Count),
	})
	if err != nil {
		return nil, err
	}
	return &ListDirResponse{
		Files:  dirContent.Files,
		Offset: uint32(dirContent.Offset),
		Total:  uint32(dirContent.Total),
	}, nil
}

func (s *gRPCHandler) GetFile(ctx context.Context, request *GetFileRequest) (*GetFileResponse, error) {
	file, err := GetFile(ctx, request.SourceId, request.Path, GetFileOptions{WithAttrs: request.WithAttributes})
	return &GetFileResponse{File: file}, err
}

func (s *gRPCHandler) DeleteFile(ctx context.Context, request *DeleteFileRequest) (*DeleteFileResponse, error) {
	return &DeleteFileResponse{}, DeleteFile(ctx, request.SourceId, request.Path, DeleteFileOptions{})
}

func (s *gRPCHandler) SetFileAttributes(ctx context.Context, request *SetFileAttributesRequest) (*SetFileAttributesResponse, error) {
	return &SetFileAttributesResponse{}, SetFileAttributes(ctx, request.SourceId, request.Path, request.Attributes)
}

func (s *gRPCHandler) GetFileAttributes(ctx context.Context, request *GetFileAttributesRequest) (*GetFileAttributesResponse, error) {
	attrs, err := GetFileAttributes(ctx, request.SourceId, request.Path, request.Names...)
	return &GetFileAttributesResponse{Attributes: attrs}, err
}

func (s *gRPCHandler) RenameFile(ctx context.Context, request *RenameFileRequest) (*RenameFileResponse, error) {
	return &RenameFileResponse{}, RenameFile(ctx, request.SourceId, request.Path, request.NewName)
}

func (s *gRPCHandler) MoveFile(ctx context.Context, request *MoveFileRequest) (*MoveFileResponse, error) {
	return &MoveFileResponse{}, MoveFile(ctx, request.SourceId, request.Path, request.TargetDir)
}

func (s *gRPCHandler) CopyFile(ctx context.Context, request *CopyFileRequest) (*CopyFileResponse, error) {
	return &CopyFileResponse{}, CopyFile(ctx, request.SourceId, request.Path, request.TargetDir)
}
