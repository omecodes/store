package files

import (
	"context"
	"github.com/omecodes/errors"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"
)

type ParamsHandler struct {
	BaseHandler
}

func (h *ParamsHandler) CreateAccess(ctx context.Context, source *pb.FSAccess, opts CreateAccessOptions) error {
	if source == nil || source.Uri == "" {
		return errors.BadRequest("invalid source value")
	}
	return h.next.CreateAccess(ctx, source, opts)
}

func (h *ParamsHandler) GetAccess(ctx context.Context, sourceID string, opts GetAccessOptions) (*pb.FSAccess, error) {
	if sourceID == "" {
		return nil, errors.BadRequest("source id is required")
	}
	return h.next.GetAccess(ctx, sourceID, opts)
}

func (h *ParamsHandler) DeleteAccess(ctx context.Context, sourceID string, opts DeleteAccessOptions) error {
	if sourceID == "" {
		return errors.BadRequest("source id is required")
	}
	return h.next.DeleteAccess(ctx, sourceID, opts)
}

func (h *ParamsHandler) CreateDir(ctx context.Context, sourceID string, dirname string, opts CreateDirOptions) error {
	if sourceID == "" {
		return errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})
	}

	if dirname == "" {
		return errors.BadRequest("missing parameters", errors.Details{
			Key:   "filename",
			Value: "required",
		})
	}

	return h.next.CreateDir(ctx, sourceID, dirname, opts)
}

func (h *ParamsHandler) WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	if sourceID == "" || filename == "" || content == nil || size == 0 {
		err := errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})

		if filename == "" {
			err.AddDetails("filename", "required")
		}

		if content == nil || size == 0 {
			err.AddDetails("content", "required")
		}

		return err
	}

	return h.next.WriteFileContent(ctx, sourceID, filename, content, size, opts)
}

func (h *ParamsHandler) ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	if sourceID == "" {
		return nil, errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})
	}

	if dirname == "" {
		return nil, errors.BadRequest("missing parameters", errors.Details{
			Key:   "dirname",
			Value: "required",
		})
	}

	return h.next.ListDir(ctx, sourceID, dirname, opts)
}

func (h *ParamsHandler) ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	if filename == "" {
		return nil, 0, errors.BadRequest("missing parameters", errors.Details{
			Key:   "filename",
			Value: "required",
		})
	}

	return h.next.ReadFileContent(ctx, sourceID, filename, opts)
}

func (h *ParamsHandler) GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileOptions) (*pb.File, error) {
	if sourceID == "" {
		return nil, errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})
	}

	if filename == "" {
		return nil, errors.BadRequest("missing parameters", errors.Details{
			Key:   "filename",
			Value: "required",
		})
	}

	return h.next.GetFileInfo(ctx, sourceID, filename, opts)
}

func (h *ParamsHandler) DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error {
	if sourceID == "" {
		return errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})
	}

	if filename == "" {
		return errors.BadRequest("missing parameters", errors.Details{
			Key:   "filename",
			Value: "required",
		})
	}

	return h.next.DeleteFile(ctx, sourceID, filename, opts)
}

func (h *ParamsHandler) SetFileAttributes(ctx context.Context, sourceID string, filename string, attrs Attributes, opts SetFileAttributesOptions) error {
	if sourceID == "" {
		return errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})
	}

	if filename == "" || len(attrs) == 0 {
		err := errors.BadRequest("missing parameters")
		if filename == "" {
			err.AddDetails("filename", "required")
		}

		if len(attrs) == 0 {
			err.AddDetails("attributes", "required")
		}
		return err
	}

	return h.next.SetFileAttributes(ctx, sourceID, filename, attrs, opts)
}

func (h *ParamsHandler) GetFileAttributes(ctx context.Context, sourceID string, filename string, names []string, opts GetFileAttributesOptions) (Attributes, error) {
	if sourceID == "" {
		return nil, errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})
	}

	if filename == "" || len(names) == 0 {
		err := errors.BadRequest("missing parameters")
		if filename == "" {
			err.AddDetails("filename", "required")
		}

		if len(names) == 0 {
			err.AddDetails("names", "required")
		}
		return nil, err
	}

	return h.next.GetFileAttributes(ctx, sourceID, filename, names, opts)
}

func (h *ParamsHandler) RenameFile(ctx context.Context, sourceID string, filename string, newName string, opts RenameFileOptions) error {
	if sourceID == "" {
		return errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})
	}

	if filename == "" || newName == "" {
		err := errors.BadRequest("missing parameters")
		if filename == "" {
			err.AddDetails("filename", "required")
		}

		if newName == "" {
			err.AddDetails("new_name", "required")
		}
		return err
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" || fPath == "" {
		return errors.BadRequest("wrong path format")
	}

	return h.next.RenameFile(ctx, sourceID, filename, newName, opts)
}

func (h *ParamsHandler) MoveFile(ctx context.Context, sourceID string, filename string, dirname string, opts MoveFileOptions) error {
	if sourceID == "" {
		return errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})
	}

	if filename == "" || dirname == "" {
		err := errors.BadRequest("missing parameters")
		if filename == "" {
			err.AddDetails("filename", "required")
		}

		if dirname == "" {
			err.AddDetails("new_filename", "required")
		}
		return err
	}

	return h.next.MoveFile(ctx, sourceID, filename, dirname, opts)
}

func (h *ParamsHandler) CopyFile(ctx context.Context, sourceID string, filename string, dirname string, opts CopyFileOptions) error {
	if sourceID == "" {
		return errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})
	}

	if filename == "" || dirname == "" {
		err := errors.BadRequest("missing parameters")
		if filename == "" {
			err.AddDetails("filename", "required")
		}

		if dirname == "" {
			err.AddDetails("copy_filename", "required")
		}
		return err
	}

	return h.next.CopyFile(ctx, sourceID, filename, dirname, opts)
}

func (h *ParamsHandler) OpenMultipartSession(ctx context.Context, sourceID string, filename string, info MultipartSessionInfo, opts OpenMultipartSessionOptions) (string, error) {
	if sourceID == "" {
		return "", errors.BadRequest("missing parameters", errors.Details{
			Key:   "source",
			Value: "required",
		})
	}

	if filename == "" {
		err := errors.BadRequest("missing parameters")
		if filename == "" {
			err.AddDetails("filename", "required")
		}
		return "", err
	}

	return h.next.OpenMultipartSession(ctx, sourceID, filename, info, opts)
}

func (h *ParamsHandler) WriteFilePart(ctx context.Context, sessionID string, content io.Reader, size int64, info ContentPartInfo, opts WriteFilePartOptions) (int64, error) {
	if sessionID == "" || content == nil || size == 0 {
		err := errors.BadRequest("missing parameters")
		if sessionID == "" {
			err.AddDetails("session_id", "required")
		}

		return 0, err
	}
	return h.next.WriteFilePart(ctx, sessionID, content, size, info, opts)
}

func (h *ParamsHandler) CloseMultipartSession(ctx context.Context, sessionID string, opts CloseMultipartSessionOptions) error {
	if sessionID == "" {
		return errors.BadRequest("missing parameters", errors.Details{
			Key:   "session_id",
			Value: "required",
		})
	}
	return h.next.CloseMultipartSession(ctx, sessionID, opts)
}
