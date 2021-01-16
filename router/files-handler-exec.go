package router

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/pb"
	"io"
)

type FilesExecHandler struct {
	FilesBaseObjectsHandler
}

func (h *FilesExecHandler) CreateDir(ctx context.Context, filename string) error {
	panic("")
}

func (h *FilesExecHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, accessRules *pb.FileAccessRules, opts pb.PutFileOptions) error {
	panic("implement me")
}

func (h *FilesExecHandler) ListDir(ctx context.Context, dirname string, opts pb.GetFileInfoOptions) ([]*pb.File, error) {
	if dirname == "" {
		return nil, errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "dirname",
			Details: "required",
		})
	}
	return h.next.ListDir(ctx, dirname, opts)
}

func (h *FilesExecHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	panic("implement me")
}

func (h *FilesExecHandler) GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error) {
	panic("implement me")
}

func (h *FilesExecHandler) DeleteFile(ctx context.Context, filename string) error {
	panic("implement me")
}

func (h *FilesExecHandler) SetFileMetaData(ctx context.Context, filename string, name string, value string) error {
	panic("implement me")
}

func (h *FilesExecHandler) GetFileMetaData(ctx context.Context, filename string, name string) (string, error) {
	panic("implement me")
}

func (h *FilesExecHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	panic("implement me")
}

func (h *FilesExecHandler) MoveFile(ctx context.Context, srcFilename string, dstFilename string) error {
	panic("implement me")
}

func (h *FilesExecHandler) CopyFile(ctx context.Context, srcFilename string, dstFilename string) error {
	panic("implement me")
}

func (h *FilesExecHandler) OpenMultipartSession(ctx context.Context, filename string, info *pb.MultipartSessionInfo) (string, error) {
	panic("implement me")
}

func (h *FilesExecHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error {
	panic("implement me")
}

func (h *FilesExecHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	panic("implement me")
}
