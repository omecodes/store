package files

import (
	"context"
	"io"
)

type BaseHandler struct {
	next Handler
}

func (h *BaseHandler) CreateSource(ctx context.Context, source *Source) error {
	return h.next.CreateSource(ctx, source)
}

func (h *BaseHandler) ListSources(ctx context.Context) ([]*Source, error) {
	return h.next.ListSources(ctx)
}

func (h *BaseHandler) GetSource(ctx context.Context, sourceID string) (*Source, error) {
	return h.next.GetSource(ctx, sourceID)
}

func (h *BaseHandler) DeleteSource(ctx context.Context, sourceID string) error {
	return h.next.DeleteSource(ctx, sourceID)
}

func (h *BaseHandler) CreateDir(ctx context.Context, sourceID string, filename string) error {
	return h.next.CreateDir(ctx, sourceID, filename)
}

func (h *BaseHandler) WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	return h.next.WriteFileContent(ctx, sourceID, filename, content, size, opts)
}

func (h *BaseHandler) ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	return h.next.ListDir(ctx, sourceID, dirname, opts)
}

func (h *BaseHandler) ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	return h.next.ReadFileContent(ctx, sourceID, filename, opts)
}

func (h *BaseHandler) GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileOptions) (*File, error) {
	return h.next.GetFileInfo(ctx, sourceID, filename, opts)
}

func (h *BaseHandler) DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error {
	return h.next.DeleteFile(ctx, sourceID, filename, opts)
}

func (h *BaseHandler) SetFileAttributes(ctx context.Context, sourceID string, filename string, attrs Attributes) error {
	return h.next.SetFileAttributes(ctx, sourceID, filename, attrs)
}

func (h *BaseHandler) GetFileAttributes(ctx context.Context, sourceID string, filename string, name ...string) (Attributes, error) {
	return h.next.GetFileAttributes(ctx, sourceID, filename, name...)
}

func (h *BaseHandler) RenameFile(ctx context.Context, sourceID string, filename string, newName string) error {
	return h.next.RenameFile(ctx, sourceID, filename, newName)
}

func (h *BaseHandler) MoveFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	return h.next.MoveFile(ctx, sourceID, filename, dirname)
}

func (h *BaseHandler) CopyFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	return h.next.CopyFile(ctx, sourceID, filename, dirname)
}

func (h *BaseHandler) OpenMultipartSession(ctx context.Context, sourceID string, filename string, info MultipartSessionInfo) (string, error) {
	return h.next.OpenMultipartSession(ctx, sourceID, filename, info)
}

func (h *BaseHandler) WriteFilePart(ctx context.Context, sessionID string, content io.Reader, size int64, info ContentPartInfo) (int64, error) {
	return h.next.WriteFilePart(ctx, sessionID, content, size, info)
}

func (h *BaseHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	return h.next.CloseMultipartSession(ctx, sessionId)
}
