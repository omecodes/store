package files

import (
	"context"
	"io"
)

type BaseHandler struct {
	next Handler
}

func (h *BaseHandler) CreateDir(ctx context.Context, filename string) error {
	return h.next.CreateDir(ctx, filename)
}

func (h *BaseHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts WriteOptions) error {
	return h.next.WriteFileContent(ctx, filename, content, size, opts)
}

func (h *BaseHandler) ListDir(ctx context.Context, dirname string, opts ListDirOptions) (*DirContent, error) {
	return h.next.ListDir(ctx, dirname, opts)
}

func (h *BaseHandler) ReadFileContent(ctx context.Context, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	return h.next.ReadFileContent(ctx, filename, opts)
}

func (h *BaseHandler) GetFileInfo(ctx context.Context, filename string, opts GetFileInfoOptions) (*File, error) {
	return h.next.GetFileInfo(ctx, filename, opts)
}

func (h *BaseHandler) DeleteFile(ctx context.Context, filename string, opts DeleteFileOptions) error {
	return h.next.DeleteFile(ctx, filename, opts)
}

func (h *BaseHandler) SetFileMetaData(ctx context.Context, filename string, attrs Attributes) error {
	return h.next.SetFileMetaData(ctx, filename, attrs)
}

func (h *BaseHandler) GetFileAttributes(ctx context.Context, filename string, name ...string) (Attributes, error) {
	return h.next.GetFileAttributes(ctx, filename, name...)
}

func (h *BaseHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	return h.next.RenameFile(ctx, filename, newName)
}

func (h *BaseHandler) MoveFile(ctx context.Context, filename string, dirname string) error {
	return h.next.MoveFile(ctx, filename, dirname)
}

func (h *BaseHandler) CopyFile(ctx context.Context, filename string, dirname string) error {
	return h.next.CopyFile(ctx, filename, dirname)
}

func (h *BaseHandler) OpenMultipartSession(ctx context.Context, filename string, info *MultipartSessionInfo) (string, error) {
	return h.next.OpenMultipartSession(ctx, filename, info)
}

func (h *BaseHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *ContentPartInfo) error {
	return h.next.AddContentPart(ctx, sessionID, content, size, info)
}

func (h *BaseHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	return h.next.CloseMultipartSession(ctx, sessionId)
}
