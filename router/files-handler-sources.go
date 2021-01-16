package router

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/pb"
	"io"
)

type FilesSourcesHandler struct {
	FilesBaseObjectsHandler
}

func (h *FilesSourcesHandler) CreateDir(ctx context.Context, filename string) error {
	panic("not implemented")
}

func (h *FilesSourcesHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, accessRules *pb.FileAccessRules, opts pb.PutFileOptions) error {
	panic("implement me")
}

func (h *FilesSourcesHandler) ListDir(ctx context.Context, dirname string, opts pb.GetFileInfoOptions) ([]*pb.File, error) {
	if dirname == "" {
		return nil, errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "dirname",
			Details: "required",
		})
	}
	return h.next.ListDir(ctx, dirname, opts)
}

func (h *FilesSourcesHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	panic("implement me")
}

func (h *FilesSourcesHandler) GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error) {
	panic("implement me")
}

func (h *FilesSourcesHandler) DeleteFile(ctx context.Context, filename string) error {
	panic("implement me")
}

func (h *FilesSourcesHandler) SetFileMetaData(ctx context.Context, filename string, name string, value string) error {
	panic("implement me")
}

func (h *FilesSourcesHandler) GetFileMetaData(ctx context.Context, filename string, name string) (string, error) {
	panic("implement me")
}

func (h *FilesSourcesHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	panic("implement me")
}

func (h *FilesSourcesHandler) MoveFile(ctx context.Context, srcFilename string, dstFilename string) error {
	panic("implement me")
}

func (h *FilesSourcesHandler) CopyFile(ctx context.Context, srcFilename string, dstFilename string) error {
	panic("implement me")
}

func (h *FilesSourcesHandler) OpenMultipartSession(ctx context.Context, filename string, info *pb.MultipartSessionInfo) (string, error) {
	panic("implement me")
}

func (h *FilesSourcesHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error {
	panic("implement me")
}

func (h *FilesSourcesHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	panic("implement me")
}
