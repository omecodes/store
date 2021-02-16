package files

import (
	"context"
	"github.com/omecodes/libome/logs"
	"io"
	"net/url"
	"strings"

	"github.com/omecodes/errors"
)

type FS interface {
	Mkdir(ctx context.Context, dirname string) error
	Ls(ctx context.Context, dirname string, offset int, count int) (*DirContent, error)
	Write(ctx context.Context, filename string, content io.Reader, append bool) error
	Read(ctx context.Context, filename string, offset int64, count int64) (io.ReadCloser, int64, error)
	Info(ctx context.Context, filename string, withAttrs bool) (*File, error)
	SetAttributes(ctx context.Context, filename string, attrs Attributes) error
	GetAttributes(ctx context.Context, filename string, names ...string) (Attributes, error)
	Rename(ctx context.Context, filename string, newName string) error
	Move(ctx context.Context, filename string, dirname string) error
	Copy(ctx context.Context, filename string, dirname string) error
	DeleteFile(ctx context.Context, filename string, recursive bool) error
}

type ctxFsProvider struct{}

type FSProvider interface {
	GetFS(source *Source) (FS, error)
}

func getFS(ctx context.Context, sourceID string) (FS, error) {
	source, err := resolveSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	logs.Info("FS: resolved source", logs.Details("uri", source.URI))

	o := ctx.Value(ctxFsProvider{})
	if o != nil {
		provider := o.(FSProvider)
		return provider.GetFS(source)
	}

	if source.Type != TypeDisk {
		return nil, errors.Create(errors.NotImplemented, "file source type is not supported")
	}

	uri, err := url.Parse(source.URI)
	if err != nil {
		return nil, err
	}

	switch uri.Scheme {
	case SchemeFS:
		rootDir := strings.TrimPrefix(source.URI, SchemeFS+"://")
		return &diskFS{root: rootDir}, nil

	default:
		return nil, errors.Create(errors.BadRequest, "not supported scheme")
	}
}
