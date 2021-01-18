package router

import (
	"context"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/pb"
	"io"
)

type FilesExecHandler struct {
	FilesBaseHandler
}

func (h *FilesExecHandler) CreateDir(ctx context.Context, dirname string) error {
	return files.Mkdir(ctx, dirname)
}

func (h *FilesExecHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts pb.PutFileOptions) error {
	return files.Write(ctx, filename, content, opts.Append)
}

func (h *FilesExecHandler) ListDir(ctx context.Context, dirname string, opts pb.ListDirOptions) (*pb.DirContent, error) {
	return files.Ls(ctx, dirname, opts.Offset, opts.Count)
}

func (h *FilesExecHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	return files.Read(ctx, filename, opts.Range.Offset, opts.Range.Length)
}

func (h *FilesExecHandler) GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error) {
	return files.Info(ctx, filename, opts.WithAttrs)
}

func (h *FilesExecHandler) DeleteFile(ctx context.Context, filename string, opts *pb.DeleteFileOptions) error {
	return files.DeleteFile(ctx, filename, opts.Recursive)
}

func (h *FilesExecHandler) SetFileMetaData(ctx context.Context, filename string, attrs files.Attributes) error {
	return files.SetAttributes(ctx, filename, attrs)
}

func (h *FilesExecHandler) GetFileAttributes(ctx context.Context, filename string, names ...string) (files.Attributes, error) {
	return files.GetAttributes(ctx, filename, names...)
}

func (h *FilesExecHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	return files.Rename(ctx, filename, newName)
}

func (h *FilesExecHandler) MoveFile(ctx context.Context, filename string, dirname string) error {
	return files.Move(ctx, filename, dirname)
}

func (h *FilesExecHandler) CopyFile(ctx context.Context, filename string, dirname string) error {
	return files.Copy(ctx, filename, dirname)
}

func (h *FilesExecHandler) OpenMultipartSession(ctx context.Context, filename string, info *pb.MultipartSessionInfo) (string, error) {
	panic("implement me")
}

func (h *FilesExecHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error {
	panic("implement me")
}

func (h *FilesExecHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	panic("implement me")
}
