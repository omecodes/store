package files

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"
)

type Handler interface {
	CreateAccess(ctx context.Context, access *pb.FSAccess, opts CreateAccessOptions) error
	GetAccessList(ctx context.Context, opts GetAccessListOptions) ([]*pb.FSAccess, error)
	GetAccess(ctx context.Context, accessID string, opts GetAccessOptions) (*pb.FSAccess, error)
	DeleteAccess(ctx context.Context, accessID string, opts DeleteAccessOptions) error
	CreateDir(ctx context.Context, accessID string, dirname string, opts CreateDirOptions) error
	ListDir(ctx context.Context, accessID string, dirname string, opts ListDirOptions) (*DirContent, error)
	WriteFileContent(ctx context.Context, accessID string, filename string, content io.Reader, size int64, opts WriteOptions) error
	ReadFileContent(ctx context.Context, accessID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error)
	GetFileInfo(ctx context.Context, accessID string, filename string, opts GetFileOptions) (*pb.File, error)
	DeleteFile(ctx context.Context, accessID string, filename string, opts DeleteFileOptions) error
	SetFileAttributes(ctx context.Context, accessID string, filename string, attrs Attributes, opts SetFileAttributesOptions) error
	GetFileAttributes(ctx context.Context, accessID string, filename string, names []string, opts GetFileAttributesOptions) (Attributes, error)
	RenameFile(ctx context.Context, accessID string, filename string, newName string, opts RenameFileOptions) error
	MoveFile(ctx context.Context, accessID string, filename string, dirname string, opts MoveFileOptions) error
	CopyFile(ctx context.Context, accessID string, filename string, dirname string, opts CopyFileOptions) error
	OpenMultipartSession(ctx context.Context, accessID string, filename string, info MultipartSessionInfo, opts OpenMultipartSessionOptions) (string, error)
	WriteFilePart(ctx context.Context, accessID string, content io.Reader, size int64, info ContentPartInfo, opts WriteFilePartOptions) (int64, error)
	CloseMultipartSession(ctx context.Context, sessionID string, opts CloseMultipartSessionOptions) error
}

func CreateAccess(ctx context.Context, source *pb.FSAccess, opts CreateAccessOptions) error {
	return GetRouteHandler(ctx).CreateAccess(ctx, source, opts)
}

func GetAccessList(ctx context.Context, opts GetAccessListOptions) ([]*pb.FSAccess, error) {
	return GetRouteHandler(ctx).GetAccessList(ctx, opts)
}

func GetAccess(ctx context.Context, sourceID string, opts GetAccessOptions) (*pb.FSAccess, error) {
	return GetRouteHandler(ctx).GetAccess(ctx, sourceID, opts)
}

func DeleteAccess(ctx context.Context, sourceID string, opts DeleteAccessOptions) error {
	return GetRouteHandler(ctx).DeleteAccess(ctx, sourceID, opts)
}

func CreateDir(ctx context.Context, sourceID string, dirname string, opts CreateDirOptions) error {
	return GetRouteHandler(ctx).CreateDir(ctx, sourceID, dirname, opts)
}

func ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	return GetRouteHandler(ctx).ListDir(ctx, sourceID, dirname, opts)
}

func WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	return GetRouteHandler(ctx).WriteFileContent(ctx, sourceID, filename, content, size, opts)
}

func ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	return GetRouteHandler(ctx).ReadFileContent(ctx, sourceID, filename, opts)
}

func GetFile(ctx context.Context, sourceID string, filename string, opts GetFileOptions) (*pb.File, error) {
	return GetRouteHandler(ctx).GetFileInfo(ctx, sourceID, filename, opts)
}

func DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error {
	return GetRouteHandler(ctx).DeleteFile(ctx, sourceID, filename, opts)
}

func SetFileAttributes(ctx context.Context, sourceID string, filename string, attrs Attributes, opts SetFileAttributesOptions) error {
	return GetRouteHandler(ctx).SetFileAttributes(ctx, sourceID, filename, attrs, opts)
}

func GetFileAttributes(ctx context.Context, sourceID string, filename string, name []string, opts GetFileAttributesOptions) (Attributes, error) {
	return GetRouteHandler(ctx).GetFileAttributes(ctx, sourceID, filename, name, opts)
}

func RenameFile(ctx context.Context, sourceID string, filename string, newName string, opts RenameFileOptions) error {
	return GetRouteHandler(ctx).RenameFile(ctx, sourceID, filename, newName, opts)
}

func MoveFile(ctx context.Context, sourceID string, filename string, dirname string, opts MoveFileOptions) error {
	return GetRouteHandler(ctx).MoveFile(ctx, sourceID, filename, dirname, opts)
}

func CopyFile(ctx context.Context, sourceID string, filename string, dirname string, opts CopyFileOptions) error {
	return GetRouteHandler(ctx).CopyFile(ctx, sourceID, filename, dirname, opts)
}

func OpenMultipartSession(ctx context.Context, sourceID string, filename string, info MultipartSessionInfo, opts OpenMultipartSessionOptions) (string, error) {
	return GetRouteHandler(ctx).OpenMultipartSession(ctx, sourceID, filename, info, opts)
}

func AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info ContentPartInfo, opts WriteFilePartOptions) (int64, error) {
	return GetRouteHandler(ctx).WriteFilePart(ctx, sessionID, content, size, info, opts)
}

func CloseMultipartSession(ctx context.Context, sessionId string, opts CloseMultipartSessionOptions) error {
	return GetRouteHandler(ctx).CloseMultipartSession(ctx, sessionId, opts)
}
