package files

import (
	"context"
	"io"
)

func NewHandlerServiceClient(clientType uint32) Handler {
	return &ServiceClientHandler{
		clientType: clientType,
	}
}

type ServiceClientHandler struct {
	BaseHandler
	clientType uint32
}

func (h *ServiceClientHandler) CreateSource(ctx context.Context, source *Source) error {
	return h.next.CreateSource(ctx, source)
}

func (h *ServiceClientHandler) ListSources(ctx context.Context) ([]*Source, error) {
	return h.next.ListSources(ctx)
}

func (h *ServiceClientHandler) GetSource(ctx context.Context, sourceID string) (*Source, error) {
	return h.next.GetSource(ctx, sourceID)
}

func (h *ServiceClientHandler) DeleteSource(ctx context.Context, sourceID string) error {
	return h.next.DeleteSource(ctx, sourceID)
}

func (h *ServiceClientHandler) CreateDir(ctx context.Context, sourceID string, filename string) error {
	return h.next.CreateDir(ctx, sourceID, filename)
}

func (h *ServiceClientHandler) WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	return h.next.WriteFileContent(ctx, sourceID, filename, content, size, opts)
}

func (h *ServiceClientHandler) ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	return h.next.ListDir(ctx, sourceID, dirname, opts)
}

func (h *ServiceClientHandler) ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	return h.next.ReadFileContent(ctx, sourceID, filename, opts)
}

func (h *ServiceClientHandler) GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileOptions) (*File, error) {
	return h.next.GetFileInfo(ctx, sourceID, filename, opts)
}

func (h *ServiceClientHandler) DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error {
	return h.next.DeleteFile(ctx, sourceID, filename, opts)
}

func (h *ServiceClientHandler) SetFileMetaData(ctx context.Context, sourceID string, filename string, attrs Attributes) error {
	return h.next.SetFileMetaData(ctx, sourceID, filename, attrs)
}

func (h *ServiceClientHandler) GetFileAttributes(ctx context.Context, sourceID string, filename string, name ...string) (Attributes, error) {
	return h.next.GetFileAttributes(ctx, sourceID, filename, name...)
}

func (h *ServiceClientHandler) RenameFile(ctx context.Context, sourceID string, filename string, newName string) error {
	return h.next.RenameFile(ctx, sourceID, filename, newName)
}

func (h *ServiceClientHandler) MoveFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	return h.next.MoveFile(ctx, sourceID, filename, dirname)
}

func (h *ServiceClientHandler) CopyFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	return h.next.CopyFile(ctx, sourceID, filename, dirname)
}

func (h *ServiceClientHandler) OpenMultipartSession(ctx context.Context, sourceID string, filename string, info *MultipartSessionInfo) (string, error) {
	return h.next.OpenMultipartSession(ctx, sourceID, filename, info)
}

func (h *ServiceClientHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *ContentPartInfo) error {
	return h.next.AddContentPart(ctx, sessionID, content, size, info)
}

func (h *ServiceClientHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	return h.next.CloseMultipartSession(ctx, sessionId)
}
