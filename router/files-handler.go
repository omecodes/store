package router

import (
	"context"
	"github.com/omecodes/store/pb"
	"io"
)

type FilesHandler interface {
	CreateDir(ctx context.Context, filename string) error
	PutContent(ctx context.Context, filename string, content io.Reader, size int64, accessRules *pb.FileAccessRules, opts pb.PutFileOptions) error
	GetContent(ctx context.Context, filename string, content io.Reader, size int64, opts pb.GetFileOptions) (io.ReadCloser, int64, error)
	GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error)
	DeleteFile(ctx context.Context, filename string) error
	SetFileMetaData(ctx context.Context, filename string, name string, value string) error
	GetFileMetaData(ctx context.Context, filename string, name string) (string, error)
	RenameFile(ctx context.Context, filename string, newName string) error
	MoveFile(ctx context.Context, srcFilename string, dstFilename string) error
	CopyFile(ctx context.Context, srcFilename string, dstFilename string) error
	OpenMultipartSession(ctx context.Context, path string, partInfo *pb.MultipartSessionInfo) (string, error)
	AddContentPart(ctx context.Context, path string, content io.Reader, size int64, info *pb.ContentPartInfo) error
	CloseMultipartSession(ctx context.Context, sessionId string) error
}
