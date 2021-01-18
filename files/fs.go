package files

import (
	"context"
	"io"
	"strings"

	"github.com/omecodes/errors"
	"github.com/omecodes/store/pb"
)

type FS interface {
	Mkdir(ctx context.Context, dirname string) error
	Ls(ctx context.Context, dirname string, offset int, count int) (*pb.DirContent, error)
	Write(ctx context.Context, filename string, content io.Reader, append bool) error
	Read(ctx context.Context, filename string, offset int64, count int64) (io.ReadCloser, int64, error)
	Info(ctx context.Context, filename string, withAttrs bool) (*pb.File, error)
	SetAttributes(ctx context.Context, filename string, attrs Attributes) error
	GetAttributes(ctx context.Context, filename string, names ...string) (Attributes, error)
	Rename(ctx context.Context, filename string, newName string) error
	Move(ctx context.Context, filename string, dirname string) error
	Copy(ctx context.Context, filename string, dirname string) error
	DeleteFile(ctx context.Context, filename string, recursive bool) error
}

func NewFS(sourceType SourceType, uri string) (FS, error) {
	if sourceType != TypeFS {
		return nil, errors.Create(errors.NotImplemented, "file source type is not supported")
	}
	rootDir := strings.TrimPrefix(SchemeFS, uri)
	return &dirFS{root: rootDir}, nil
}

func Mkdir(ctx context.Context, dirname string) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return err
	}
	return fs.Mkdir(ctx, dirname)
}

func Ls(ctx context.Context, dirname string, offset int, count int) (*pb.DirContent, error) {
	source := GetSource(ctx)
	if source == nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return nil, err
	}
	return fs.Ls(ctx, dirname, offset, count)
}

func Write(ctx context.Context, filename string, content io.Reader, append bool) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return err
	}
	return fs.Write(ctx, filename, content, append)
}

func Read(ctx context.Context, filename string, offset int64, count int64) (io.ReadCloser, int64, error) {
	source := GetSource(ctx)
	if source == nil {
		return nil, 0, errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return nil, 0, err
	}
	return fs.Read(ctx, filename, offset, count)
}

func Info(ctx context.Context, filename string, withAttrs bool) (*pb.File, error) {
	source := GetSource(ctx)
	if source == nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return nil, err
	}
	return fs.Info(ctx, filename, withAttrs)
}

func SetAttributes(ctx context.Context, filename string, attrs Attributes) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return err
	}
	return fs.SetAttributes(ctx, filename, attrs)
}

func GetAttributes(ctx context.Context, filename string, names ...string) (Attributes, error) {
	source := GetSource(ctx)
	if source == nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return nil, err
	}
	return fs.GetAttributes(ctx, filename, names...)
}

func Rename(ctx context.Context, filename string, newName string) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return err
	}

	return fs.Rename(ctx, filename, newName)
}

func Move(ctx context.Context, filename string, dirname string) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return err
	}

	return fs.Rename(ctx, filename, dirname)
}

func Copy(ctx context.Context, filename string, dirname string) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return err
	}

	return fs.Rename(ctx, filename, dirname)
}

func DeleteFile(ctx context.Context, filename string, recursive bool) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source.Type, source.URI)
	if err != nil {
		return err
	}

	return fs.DeleteFile(ctx, filename, recursive)
}
