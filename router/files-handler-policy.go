package router

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/pb"
	"io"
)

type FilesPolicyHandler struct {
	FilesBaseObjectsHandler
}

func (h *FilesPolicyHandler) CreateDir(ctx context.Context, filename string) error {
	panic("not implemented")
}

func (h *FilesPolicyHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, accessRules *pb.FileAccessRules, opts pb.PutFileOptions) error {
	panic("implement me")
}

func (h *FilesPolicyHandler) ListDir(ctx context.Context, dirname string, opts pb.GetFileInfoOptions) ([]*pb.File, error) {
	if dirname == "" {
		return nil, errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "dirname",
			Details: "required",
		})
	}
	return h.next.ListDir(ctx, dirname, opts)
}

func (h *FilesPolicyHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	panic("implement me")
}

func (h *FilesPolicyHandler) GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error) {
	panic("implement me")
}

func (h *FilesPolicyHandler) DeleteFile(ctx context.Context, filename string) error {
	panic("implement me")
}

func (h *FilesPolicyHandler) SetFileMetaData(ctx context.Context, filename string, name string, value string) error {
	panic("implement me")
}

func (h *FilesPolicyHandler) GetFileMetaData(ctx context.Context, filename string, name string) (string, error) {
	panic("implement me")
}

func (h *FilesPolicyHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	panic("implement me")
}

func (h *FilesPolicyHandler) MoveFile(ctx context.Context, srcFilename string, dstFilename string) error {
	panic("implement me")
}

func (h *FilesPolicyHandler) CopyFile(ctx context.Context, srcFilename string, dstFilename string) error {
	panic("implement me")
}

func (h *FilesPolicyHandler) OpenMultipartSession(ctx context.Context, filename string, info *pb.MultipartSessionInfo) (string, error) {
	panic("implement me")
}

func (h *FilesPolicyHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error {
	panic("implement me")
}

func (h *FilesPolicyHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	panic("implement me")
}
