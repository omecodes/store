package router

import (
	"context"
	"github.com/omecodes/store/pb"
	"io"
)

type FilesEncryptionHandler struct {
	FilesBaseObjectsHandler
}

func (h *FilesEncryptionHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts pb.PutFileOptions) error {
	panic("implement me")
}

func (h *FilesEncryptionHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	panic("implement me")
}

func (h *FilesEncryptionHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error {
	panic("implement me")
}
