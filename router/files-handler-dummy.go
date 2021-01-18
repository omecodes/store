package router

import (
	"context"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/pb"
	"io"
)

type filesDummyHandler struct {
	next FilesHandler
}

func (h *filesDummyHandler) CreateDir(ctx context.Context, filename string) error {
	return nil
}

func (h *filesDummyHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts pb.PutFileOptions) error {
	return nil
}

func (h *filesDummyHandler) ListDir(ctx context.Context, dirname string, opts pb.ListDirOptions) (*pb.DirContent, error) {
	return nil, nil
}

func (h *filesDummyHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	return nil, 0, nil
}

func (h *filesDummyHandler) GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error) {
	return nil, nil
}

func (h *filesDummyHandler) DeleteFile(ctx context.Context, filename string, opts *pb.DeleteFileOptions) error {
	return h.next.DeleteFile(ctx, filename, opts)
}

func (h *filesDummyHandler) SetFileMetaData(ctx context.Context, filename string, attrs files.Attributes) error {
	return nil
}

func (h *filesDummyHandler) GetFileAttributes(ctx context.Context, filename string, name ...string) (files.Attributes, error) {
	return nil, nil
}

func (h *filesDummyHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	return nil
}

func (h *filesDummyHandler) MoveFile(ctx context.Context, filename string, dirname string) error {
	return nil
}

func (h *filesDummyHandler) CopyFile(ctx context.Context, filename string, dirname string) error {
	return nil
}

func (h *filesDummyHandler) OpenMultipartSession(ctx context.Context, filename string, info *pb.MultipartSessionInfo) (string, error) {
	return "", nil
}

func (h *filesDummyHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error {
	return nil
}

func (h *filesDummyHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	return nil
}
