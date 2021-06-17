package files

import (
	"context"
	"github.com/omecodes/libome/logs"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"

	"github.com/omecodes/errors"
)

type ExecHandler struct {
	BaseHandler
}

func (h *ExecHandler) CreateAccess(ctx context.Context, source *pb.FSAccess, _ CreateAccessOptions) error {
	sourceManager := getAccessManager(ctx)
	if sourceManager == nil {
		return errors.Internal("context missing source manager")
	}
	_, err := sourceManager.Save(ctx, source)
	return err
}

func (h *ExecHandler) GetAccess(ctx context.Context, accessID string, _ GetAccessOptions) (*pb.FSAccess, error) {
	sourceManager := getAccessManager(ctx)
	if sourceManager == nil {
		return nil, errors.Internal("context missing source manager")
	}
	return sourceManager.Get(ctx, accessID)
}

func (h *ExecHandler) DeleteAccess(ctx context.Context, accessID string, _ DeleteAccessOptions) error {
	sourceManager := getAccessManager(ctx)
	if sourceManager == nil {
		return errors.Internal("context missing source manager")
	}
	return sourceManager.Delete(ctx, accessID)
}

func (h *ExecHandler) CreateDir(ctx context.Context, accessID string, dirname string, _ CreateDirOptions) error {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		logs.Error("could not get fs", logs.Err(err))
		return err
	}
	return fs.Mkdir(ctx, dirname)
}

func (h *ExecHandler) WriteFileContent(ctx context.Context, accessID string, filename string, content io.Reader, _ int64, opts WriteOptions) error {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		return err
	}
	return fs.Write(ctx, filename, content, opts.Append)
}

func (h *ExecHandler) ListDir(ctx context.Context, accessID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		return nil, err
	}
	return fs.Ls(ctx, dirname, opts.Offset, opts.Count)
}

func (h *ExecHandler) ReadFileContent(ctx context.Context, accessID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		return nil, 0, err
	}
	return fs.Read(ctx, filename, opts.Range.Offset, opts.Range.Length)
}

func (h *ExecHandler) GetFileInfo(ctx context.Context, accessID string, filename string, opts GetFileOptions) (*pb.File, error) {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		return nil, err
	}
	return fs.Info(ctx, filename, opts.WithAttrs)
}

func (h *ExecHandler) DeleteFile(ctx context.Context, accessID string, filename string, opts DeleteFileOptions) error {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		return err
	}
	return fs.DeleteFile(ctx, filename, opts.Recursive)
}

func (h *ExecHandler) SetFileAttributes(ctx context.Context, accessID string, filename string, attrs Attributes, _ SetFileAttributesOptions) error {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		return err
	}
	return fs.SetAttributes(ctx, filename, attrs)
}

func (h *ExecHandler) GetFileAttributes(ctx context.Context, accessID string, filename string, names []string, _ GetFileAttributesOptions) (Attributes, error) {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		return nil, err
	}
	return fs.GetAttributes(ctx, filename, names...)
}

func (h *ExecHandler) RenameFile(ctx context.Context, accessID string, filename string, newName string, _ RenameFileOptions) error {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		return err
	}
	return fs.Rename(ctx, filename, newName)
}

func (h *ExecHandler) MoveFile(ctx context.Context, accessID string, filename string, dirname string, _ MoveFileOptions) error {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		return err
	}
	return fs.Rename(ctx, filename, dirname)
}

func (h *ExecHandler) CopyFile(ctx context.Context, accessID string, filename string, dirname string, _ CopyFileOptions) error {
	fs, err := getFS(ctx, accessID)
	if err != nil {
		return err
	}
	return fs.Rename(ctx, filename, dirname)
}

func (h *ExecHandler) OpenMultipartSession(_ context.Context, _ string, _ string, _ MultipartSessionInfo, _ OpenMultipartSessionOptions) (string, error) {
	panic("implement me")
}

func (h *ExecHandler) WriteFilePart(_ context.Context, _ string, _ io.Reader, _ int64, _ ContentPartInfo, _ WriteFilePartOptions) (int64, error) {
	panic("implement me")
}

func (h *ExecHandler) CloseMultipartSession(_ context.Context, _ string, _ CloseMultipartSessionOptions) error {
	panic("implement me")
}
