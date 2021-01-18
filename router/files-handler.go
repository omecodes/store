package router

import (
	"context"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/pb"
	"io"
)

type FilesHandler interface {
	CreateDir(ctx context.Context, filename string) error
	WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts pb.PutFileOptions) error
	ListDir(ctx context.Context, dirname string, opts pb.ListDirOptions) (*pb.DirContent, error)
	ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error)
	GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error)
	DeleteFile(ctx context.Context, filename string, opts *pb.DeleteFileOptions) error
	SetFileMetaData(ctx context.Context, filename string, attrs files.Attributes) error
	GetFileAttributes(ctx context.Context, filename string, name ...string) (files.Attributes, error)
	RenameFile(ctx context.Context, filename string, newName string) error
	MoveFile(ctx context.Context, filename string, dirname string) error
	CopyFile(ctx context.Context, filename string, dirname string) error
	OpenMultipartSession(ctx context.Context, filename string, info *pb.MultipartSessionInfo) (string, error)
	AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error
	CloseMultipartSession(ctx context.Context, sessionId string) error
}
