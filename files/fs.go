package files

import (
	"context"
	"io"
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

func NewFS(source *Source) (FS, error) {
	if source.Type != TypeDisk {
		return nil, errors.Create(errors.NotImplemented, "file source type is not supported")
	}

	rootDir := strings.TrimPrefix(SchemeFS, source.URI)
	return &dirFS{root: rootDir}, nil
}

func Mkdir(ctx context.Context, sourceID string, dirname string) error {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return err
	}

	return fs.Mkdir(ctx, dirname)
}

func Ls(ctx context.Context, sourceID string, dirname string, offset int, count int) (*DirContent, error) {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return nil, errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return nil, err
	}

	return fs.Ls(ctx, dirname, offset, count)
}

func Write(ctx context.Context, sourceID string, filename string, content io.Reader, append bool) error {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return err
	}

	return fs.Write(ctx, filename, content, append)
}

func Read(ctx context.Context, sourceID string, filename string, offset int64, count int64) (io.ReadCloser, int64, error) {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return nil, 0, errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return nil, 0, errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return nil, 0, err
	}

	return fs.Read(ctx, filename, offset, count)
}

func Info(ctx context.Context, sourceID string, filename string, withAttrs bool) (*File, error) {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return nil, errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return nil, err
	}

	return fs.Info(ctx, filename, withAttrs)
}

func SetAttributes(ctx context.Context, sourceID string, filename string, attrs Attributes) error {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return err
	}

	return fs.SetAttributes(ctx, filename, attrs)
}

func GetAttributes(ctx context.Context, sourceID string, filename string, names ...string) (Attributes, error) {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return nil, errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return nil, err
	}
	return fs.GetAttributes(ctx, filename, names...)
}

func Rename(ctx context.Context, sourceID string, filename string, newName string) error {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return err
	}

	return fs.Rename(ctx, filename, newName)
}

func Move(ctx context.Context, sourceID string, filename string, dirname string) error {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return err
	}

	return fs.Rename(ctx, filename, dirname)
}

func Copy(ctx context.Context, sourceID string, filename string, dirname string) error {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return err
	}

	return fs.Rename(ctx, filename, dirname)
}

func DeleteFile(ctx context.Context, sourceID string, filename string, recursive bool) error {
	sm := GetSourceManager(ctx)
	if sm == nil {
		return errors.New("context missing source manager")
	}

	source, err := sm.Get(ctx, sourceID)
	if err != nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	fs, err := NewFS(source)
	if err != nil {
		return err
	}

	return fs.DeleteFile(ctx, filename, recursive)
}
