package files

import (
	"context"
	"io"
)

type ExecHandler struct {
	BaseHandler
}

func (h *ExecHandler) CreateDir(ctx context.Context, dirname string) error {
	return Mkdir(ctx, dirname)
}

func (h *ExecHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts WriteOptions) error {
	return Write(ctx, filename, content, opts.Append)
}

func (h *ExecHandler) ListDir(ctx context.Context, dirname string, opts ListDirOptions) (*DirContent, error) {
	return Ls(ctx, dirname, opts.Offset, opts.Count)
}

func (h *ExecHandler) ReadFileContent(ctx context.Context, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	return Read(ctx, filename, opts.Range.Offset, opts.Range.Length)
}

func (h *ExecHandler) GetFileInfo(ctx context.Context, filename string, opts GetFileInfoOptions) (*File, error) {
	return Info(ctx, filename, opts.WithAttrs)
}

func (h *ExecHandler) DeleteFile(ctx context.Context, filename string, opts DeleteFileOptions) error {
	return DeleteFile(ctx, filename, opts.Recursive)
}

func (h *ExecHandler) SetFileMetaData(ctx context.Context, filename string, attrs Attributes) error {
	return SetAttributes(ctx, filename, attrs)
}

func (h *ExecHandler) GetFileAttributes(ctx context.Context, filename string, names ...string) (Attributes, error) {
	return GetAttributes(ctx, filename, names...)
}

func (h *ExecHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	return Rename(ctx, filename, newName)
}

func (h *ExecHandler) MoveFile(ctx context.Context, filename string, dirname string) error {
	return Move(ctx, filename, dirname)
}

func (h *ExecHandler) CopyFile(ctx context.Context, filename string, dirname string) error {
	return Copy(ctx, filename, dirname)
}

func (h *ExecHandler) OpenMultipartSession(ctx context.Context, filename string, info *MultipartSessionInfo) (string, error) {
	panic("implement me")
}

func (h *ExecHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *ContentPartInfo) error {
	panic("implement me")
}

func (h *ExecHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	panic("implement me")
}
