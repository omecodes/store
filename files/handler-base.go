package files

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"
)

type BaseHandler struct {
	next Handler
}

func (h *BaseHandler) CreateAccess(ctx context.Context, source *pb.FSAccess, opts CreateAccessOptions) error {
	return h.next.CreateAccess(ctx, source, opts)
}

func (h *BaseHandler) GetAccessList(ctx context.Context, opts GetAccessListOptions) ([]*pb.FSAccess, error) {
	return h.next.GetAccessList(ctx, opts)
}

func (h *BaseHandler) GetAccess(ctx context.Context, sourceID string, opts GetAccessOptions) (*pb.FSAccess, error) {
	return h.next.GetAccess(ctx, sourceID, opts)
}

func (h *BaseHandler) DeleteAccess(ctx context.Context, sourceID string, opts DeleteAccessOptions) error {
	return h.next.DeleteAccess(ctx, sourceID, opts)
}

func (h *BaseHandler) CreateDir(ctx context.Context, sourceID string, dirname string, opts CreateDirOptions) error {
	return h.next.CreateDir(ctx, sourceID, dirname, opts)
}

func (h *BaseHandler) WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	return h.next.WriteFileContent(ctx, sourceID, filename, content, size, opts)
}

func (h *BaseHandler) Share(ctx context.Context, shares []*pb.ShareInfo, opts ShareOptions) error {
	return h.next.Share(ctx, shares, opts)
}

func (h *BaseHandler) GetShares(ctx context.Context, accessID string, opts GetSharesOptions) ([]*pb.UserRole, error) {
	return h.next.GetShares(ctx, accessID, opts)
}

func (h *BaseHandler) DeleteShares(ctx context.Context, shares []*pb.ShareInfo, opts DeleteSharesOptions) error {
	return h.next.DeleteShares(ctx, shares, opts)
}

func (h *BaseHandler) ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	return h.next.ListDir(ctx, sourceID, dirname, opts)
}

func (h *BaseHandler) ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	return h.next.ReadFileContent(ctx, sourceID, filename, opts)
}

func (h *BaseHandler) GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileOptions) (*pb.File, error) {
	return h.next.GetFileInfo(ctx, sourceID, filename, opts)
}

func (h *BaseHandler) DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error {
	return h.next.DeleteFile(ctx, sourceID, filename, opts)
}

func (h *BaseHandler) SetFileAttributes(ctx context.Context, sourceID string, filename string, attrs Attributes, opts SetFileAttributesOptions) error {
	return h.next.SetFileAttributes(ctx, sourceID, filename, attrs, opts)
}

func (h *BaseHandler) GetFileAttributes(ctx context.Context, sourceID string, filename string, names []string, opts GetFileAttributesOptions) (Attributes, error) {
	return h.next.GetFileAttributes(ctx, sourceID, filename, names, opts)
}

func (h *BaseHandler) RenameFile(ctx context.Context, sourceID string, filename string, newName string, opts RenameFileOptions) error {
	return h.next.RenameFile(ctx, sourceID, filename, newName, opts)
}

func (h *BaseHandler) MoveFile(ctx context.Context, sourceID string, filename string, dirname string, opts MoveFileOptions) error {
	return h.next.MoveFile(ctx, sourceID, filename, dirname, opts)
}

func (h *BaseHandler) CopyFile(ctx context.Context, sourceID string, filename string, dirname string, opts CopyFileOptions) error {
	return h.next.CopyFile(ctx, sourceID, filename, dirname, opts)
}

func (h *BaseHandler) OpenMultipartSession(ctx context.Context, sourceID string, filename string, info MultipartSessionInfo, opts OpenMultipartSessionOptions) (string, error) {
	return h.next.OpenMultipartSession(ctx, sourceID, filename, info, opts)
}

func (h *BaseHandler) WriteFilePart(ctx context.Context, sessionID string, content io.Reader, size int64, info ContentPartInfo, opts WriteFilePartOptions) (int64, error) {
	return h.next.WriteFilePart(ctx, sessionID, content, size, info, opts)
}

func (h *BaseHandler) CloseMultipartSession(ctx context.Context, sessionId string, opts CloseMultipartSessionOptions) error {
	return h.next.CloseMultipartSession(ctx, sessionId, opts)
}
