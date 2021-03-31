package files

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"io"
	"net/url"
	"strings"
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
	GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileOptions) (*File, error)
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

func CreateSource(ctx context.Context, source *Source) error {
	return GetRouteHandler(ctx).CreateSource(ctx, source)
}

func ListSources(ctx context.Context) ([]*Source, error) {
	return GetRouteHandler(ctx).ListSources(ctx)
}

func GetSource(ctx context.Context, sourceID string) (*Source, error) {
	return GetRouteHandler(ctx).GetSource(ctx, sourceID)
}

func ResolveSource(ctx context.Context, source *Source) (*Source, error) {
	source, err := GetSource(ctx, source.Id)
	if err != nil {
		return nil, err
	}

	resolvedSource := source
	sourceChain := []string{source.Id}
	for resolvedSource.Type == SourceType_Reference {
		u, err := url.Parse(source.Uri)
		if err != nil {
			return nil, errors.Internal("could not resolve source uri", errors.Details{Key: "source uri", Value: err})
		}

		refSourceID := u.Host
		resolvedSource, err = GetSource(ctx, refSourceID)
		if err != nil {
			logs.Error("could not load source", logs.Details("source", refSourceID), logs.Err(err))
			return nil, err
		}

		for _, src := range sourceChain {
			if src == refSourceID {
				return nil, errors.Internal("source cycle references")
			}
		}
		sourceChain = append(sourceChain, refSourceID)
		resolvedSource.Uri = strings.TrimSuffix(resolvedSource.Uri, "/") + u.Path

		logs.Info("resolved source", logs.Details("uri", resolvedSource.Uri))
	}

	return resolvedSource, nil
}

func DeleteSource(ctx context.Context, sourceID string) error {
	return GetRouteHandler(ctx).DeleteSource(ctx, sourceID)
}

func CreateDir(ctx context.Context, sourceID string, dirname string) error {
	return GetRouteHandler(ctx).CreateDir(ctx, sourceID, dirname)
}

func ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	return GetRouteHandler(ctx).ListDir(ctx, sourceID, dirname, opts)
}

func WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	return GetRouteHandler(ctx).WriteFileContent(ctx, sourceID, filename, content, size, opts)
}

func ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	return GetRouteHandler(ctx).ReadFileContent(ctx, sourceID, filename, opts)
}

func GetFile(ctx context.Context, sourceID string, filename string, opts GetFileOptions) (*File, error) {
	return GetRouteHandler(ctx).GetFileInfo(ctx, sourceID, filename, opts)
}

func DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error {
	return GetRouteHandler(ctx).DeleteFile(ctx, sourceID, filename, opts)
}

func SetFileAttributes(ctx context.Context, sourceID string, filename string, attrs Attributes) error {
	return GetRouteHandler(ctx).SetFileMetaData(ctx, sourceID, filename, attrs)
}

func GetFileAttributes(ctx context.Context, sourceID string, filename string, name ...string) (Attributes, error) {
	return GetRouteHandler(ctx).GetFileAttributes(ctx, sourceID, filename, name...)
}

func RenameFile(ctx context.Context, sourceID string, filename string, newName string) error {
	return GetRouteHandler(ctx).RenameFile(ctx, sourceID, filename, newName)
}

func MoveFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	return GetRouteHandler(ctx).MoveFile(ctx, sourceID, filename, dirname)
}

func CopyFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	return GetRouteHandler(ctx).CopyFile(ctx, sourceID, filename, dirname)
}

func OpenMultipartSession(ctx context.Context, sourceID string, filename string, info *MultipartSessionInfo) (string, error) {
	return GetRouteHandler(ctx).OpenMultipartSession(ctx, sourceID, filename, info)
}

func AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *ContentPartInfo) error {
	return GetRouteHandler(ctx).AddContentPart(ctx, sessionID, content, size, info)
}

func CloseMultipartSession(ctx context.Context, sessionId string) error {
	return GetRouteHandler(ctx).CloseMultipartSession(ctx, sessionId)
}
