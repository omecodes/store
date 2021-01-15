package router

import (
	"context"
	"github.com/omecodes/store/pb"
	"io"
)

type ObjectsHandler interface {
	CreateCollection(ctx context.Context, collection *pb.Collection) error
	GetCollection(ctx context.Context, id string) (*pb.Collection, error)
	ListCollections(ctx context.Context) ([]*pb.Collection, error)
	DeleteCollection(ctx context.Context, id string) error

	PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.TextIndex, opts pb.PutOptions) (string, error)
	PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts pb.PatchOptions) error
	MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts pb.MoveOptions) error
	GetObject(ctx context.Context, collection string, id string, opts pb.GetOptions) (*pb.Object, error)
	GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error)
	DeleteObject(ctx context.Context, collection string, id string) error
	ListObjects(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error)
	SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*pb.Cursor, error)
}

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
