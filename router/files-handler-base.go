package router

import (
	"context"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/pb"
	"io"
)

type FilesBaseObjectsHandler struct {
	next FilesHandler
}

func (h *FilesBaseObjectsHandler) CreateDir(ctx context.Context, filename string) error {
	return h.next.CreateDir(ctx, filename)
}

func (h *FilesBaseObjectsHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts pb.PutFileOptions) error {
	return h.next.WriteFileContent(ctx, filename, content, size, opts)
}

func (h *FilesBaseObjectsHandler) ListDir(ctx context.Context, dirname string, opts pb.ListDirOptions) (*pb.DirContent, error) {
	return h.next.ListDir(ctx, dirname, opts)
}

func (h *FilesBaseObjectsHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	return h.next.ReadFileContent(ctx, filename, opts)
}

func (h *FilesBaseObjectsHandler) GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error) {
	return h.next.GetFileInfo(ctx, filename, opts)
}

func (h *FilesBaseObjectsHandler) DeleteFile(ctx context.Context, filename string, opts *pb.DeleteFileOptions) error {
	return h.next.DeleteFile(ctx, filename, opts)
}

func (h *FilesBaseObjectsHandler) SetFileMetaData(ctx context.Context, filename string, attrs files.Attributes) error {
	return h.next.SetFileMetaData(ctx, filename, attrs)
}

func (h *FilesBaseObjectsHandler) GetFileAttributes(ctx context.Context, filename string, name ...string) (files.Attributes, error) {
	return h.next.GetFileAttributes(ctx, filename, name...)
}

func (h *FilesBaseObjectsHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	return h.next.RenameFile(ctx, filename, newName)
}

func (h *FilesBaseObjectsHandler) MoveFile(ctx context.Context, filename string, dirname string) error {
	return h.next.MoveFile(ctx, filename, dirname)
}

func (h *FilesBaseObjectsHandler) CopyFile(ctx context.Context, filename string, dirname string) error {
	return h.next.CopyFile(ctx, filename, dirname)
}

func (h *FilesBaseObjectsHandler) OpenMultipartSession(ctx context.Context, filename string, info *pb.MultipartSessionInfo) (string, error) {
	return h.next.OpenMultipartSession(ctx, filename, info)
}

func (h *FilesBaseObjectsHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error {
	return h.next.AddContentPart(ctx, sessionID, content, size, info)
}

func (h *FilesBaseObjectsHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	return h.next.CloseMultipartSession(ctx, sessionId)
}
