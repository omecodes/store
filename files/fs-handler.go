package files

import (
	"context"
	"io"
)

type Handler interface {
	CreateSource(ctx context.Context, source *Source) error
	ListSources(ctx context.Context) ([]*Source, error)
	GetSource(ctx context.Context, sourceID string) (*Source, error)
	DeleteSource(ctx context.Context, sourceID string) error
	CreateDir(ctx context.Context, sourceID string, dirname string) error
	ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error)
	WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error
	ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error)
	GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileInfoOptions) (*File, error)
	DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error
	SetFileMetaData(ctx context.Context, sourceID string, filename string, attrs Attributes) error
	GetFileAttributes(ctx context.Context, sourceID string, filename string, name ...string) (Attributes, error)
	RenameFile(ctx context.Context, sourceID string, filename string, newName string) error
	MoveFile(ctx context.Context, sourceID string, filename string, dirname string) error
	CopyFile(ctx context.Context, sourceID string, filename string, dirname string) error
	OpenMultipartSession(ctx context.Context, sourceID string, filename string, info *MultipartSessionInfo) (string, error)
	AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *ContentPartInfo) error
	CloseMultipartSession(ctx context.Context, sessionId string) error
}
