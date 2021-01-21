package files

import (
	"context"
	"github.com/omecodes/errors"
	"io"
)

type ExecHandler struct {
	BaseHandler
}

func (h *ExecHandler) CreateSource(ctx context.Context, source *Source) error {
	sourceManager := GetSourceManager(ctx)
	if sourceManager == nil {
		return errors.New("context missing source manager")
	}
	_, err := sourceManager.Save(ctx, source)
	return err
}

func (h *ExecHandler) ListSources(ctx context.Context) ([]*Source, error) {
	sourceManager := GetSourceManager(ctx)
	if sourceManager == nil {
		return nil, errors.New("context missing source manager")
	}
	return sourceManager.List(ctx)
}

func (h *ExecHandler) GetSource(ctx context.Context, sourceID string) (*Source, error) {
	sourceManager := GetSourceManager(ctx)
	if sourceManager == nil {
		return nil, errors.New("context missing source manager")
	}
	return sourceManager.Get(ctx, sourceID)
}

func (h *ExecHandler) DeleteSource(ctx context.Context, sourceID string) error {
	sourceManager := GetSourceManager(ctx)
	if sourceManager == nil {
		return errors.New("context missing source manager")
	}
	return sourceManager.Delete(ctx, sourceID)
}

func (h *ExecHandler) CreateDir(ctx context.Context, sourceID string, dirname string) error {
	return Mkdir(ctx, sourceID, dirname)
}

func (h *ExecHandler) WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	return Write(ctx, sourceID, filename, content, opts.Append)
}

func (h *ExecHandler) ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	return Ls(ctx, sourceID, dirname, opts.Offset, opts.Count)
}

func (h *ExecHandler) ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	return Read(ctx, sourceID, filename, opts.Range.Offset, opts.Range.Length)
}

func (h *ExecHandler) GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileInfoOptions) (*File, error) {
	return Info(ctx, sourceID, filename, opts.WithAttrs)
}

func (h *ExecHandler) DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error {
	return DeleteFile(ctx, sourceID, filename, opts.Recursive)
}

func (h *ExecHandler) SetFileMetaData(ctx context.Context, sourceID string, filename string, attrs Attributes) error {
	return SetAttributes(ctx, sourceID, filename, attrs)
}

func (h *ExecHandler) GetFileAttributes(ctx context.Context, sourceID string, filename string, names ...string) (Attributes, error) {
	return GetAttributes(ctx, sourceID, filename, names...)
}

func (h *ExecHandler) RenameFile(ctx context.Context, sourceID string, filename string, newName string) error {
	return Rename(ctx, sourceID, filename, newName)
}

func (h *ExecHandler) MoveFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	return Move(ctx, sourceID, filename, dirname)
}

func (h *ExecHandler) CopyFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	return Copy(ctx, sourceID, filename, dirname)
}

func (h *ExecHandler) OpenMultipartSession(ctx context.Context, sourceID string, filename string, info *MultipartSessionInfo) (string, error) {
	panic("implement me")
}

func (h *ExecHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *ContentPartInfo) error {
	panic("implement me")
}

func (h *ExecHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	panic("implement me")
}
