package files

import (
	"context"
	"github.com/omecodes/store/pb"
	"io"
)

type FS interface {
	Mkdir(ctx context.Context, dirname string) error
	Ls(ctx context.Context, dirname string, offset int, count int) (*pb.DirContent, error)
	Write(ctx context.Context, filename string, content io.Reader, append bool) (string, error)
	Read(ctx context.Context, filename string, offset int64, count int64) (io.ReadCloser, int64, error)
	Info(ctx context.Context, filename string, withInfo bool) (*pb.File, error)
	SetAttributes(ctx context.Context, attrs ...*Attribute) error
	GetAttributes(ctx context.Context, attrs ...Attribute) (Attributes, error)
	Rename(ctx context.Context, filename string, newName string) (string, error)
	Move(ctx context.Context, filename string, dirname string) error
	DeleteDir(ctx context.Context, filename string, recursive bool) error
	DeleteFile(ctx context.Context, dirname string) error
}
