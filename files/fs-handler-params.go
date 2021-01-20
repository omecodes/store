package files

import (
	"context"
	"github.com/omecodes/errors"
	"io"
)

type ParamsHandler struct {
	BaseHandler
}

func (h *ParamsHandler) CreateDir(ctx context.Context, filename string) error {
	if filename == "" {
		return errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "filename",
			Details: "required",
		})
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" || fPath == "" {
		return errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.CreateDir(ctx, filename)
}

func (h *ParamsHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts WriteOptions) error {
	if filename == "" || content == nil {
		err := errors.Create(errors.BadRequest, "missing parameters")
		if filename == "" {
			err.AppendDetails(errors.Info{
				Name:    "filename",
				Details: "required",
			})
		}

		if content == nil {
			err.AppendDetails(errors.Info{
				Name:    "content",
				Details: "required",
			})
		}
		return err
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" || fPath == "" {
		return errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.WriteFileContent(ctx, filename, content, size, opts)
}

func (h *ParamsHandler) ListDir(ctx context.Context, dirname string, opts ListDirOptions) (*DirContent, error) {
	if dirname == "" {
		return nil, errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "dirname",
			Details: "required",
		})
	}

	sourceID, fPath := Split(dirname)
	if sourceID == "" || fPath == "" {
		return nil, errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.ListDir(ctx, dirname, opts)
}

func (h *ParamsHandler) ReadFileContent(ctx context.Context, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	if filename == "" {
		return nil, 0, errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "filename",
			Details: "required",
		})
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" || fPath == "" {
		return nil, 0, errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.ReadFileContent(ctx, filename, opts)
}

func (h *ParamsHandler) GetFileInfo(ctx context.Context, filename string, opts GetFileInfoOptions) (*File, error) {
	if filename == "" {
		return nil, errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "filename",
			Details: "required",
		})
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" || fPath == "" {
		return nil, errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.GetFileInfo(ctx, filename, opts)
}

func (h *ParamsHandler) DeleteFile(ctx context.Context, filename string, opts DeleteFileOptions) error {
	if filename == "" {
		return errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "filename",
			Details: "required",
		})
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" || fPath == "" {
		return errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.DeleteFile(ctx, filename, opts)
}

func (h *ParamsHandler) SetFileMetaData(ctx context.Context, filename string, attrs Attributes) error {
	if filename == "" || len(attrs) == 0 {
		err := errors.Create(errors.BadRequest, "missing parameters")
		if filename == "" {
			err.AppendDetails(errors.Info{
				Name:    "filename",
				Details: "required",
			})
		}

		if len(attrs) == 0 {
			err.AppendDetails(errors.Info{
				Name:    "attributes",
				Details: "required",
			})
		}
		return err
	}

	sourceID, _ := Split(filename)
	if sourceID == "" {
		return errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.SetFileMetaData(ctx, filename, attrs)
}

func (h *ParamsHandler) GetFileAttributes(ctx context.Context, filename string, name ...string) (Attributes, error) {
	if filename == "" || len(name) == 0 {
		err := errors.Create(errors.BadRequest, "missing parameters")
		if filename == "" {
			err.AppendDetails(errors.Info{
				Name:    "filename",
				Details: "required",
			})
		}

		if len(name) == 0 {
			err.AppendDetails(errors.Info{
				Name:    "names",
				Details: "required",
			})
		}
		return nil, err
	}

	sourceID, _ := Split(filename)
	if sourceID == "" {
		return nil, errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.GetFileAttributes(ctx, filename, name...)
}

func (h *ParamsHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	if filename == "" || newName == "" {
		err := errors.Create(errors.BadRequest, "missing parameters")
		if filename == "" {
			err.AppendDetails(errors.Info{
				Name:    "filename",
				Details: "required",
			})
		}

		if newName == "" {
			err.AppendDetails(errors.Info{
				Name:    "new_name",
				Details: "required",
			})
		}
		return err
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" || fPath == "" {
		return errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.RenameFile(ctx, filename, newName)
}

func (h *ParamsHandler) MoveFile(ctx context.Context, filename string, dirname string) error {
	if filename == "" || dirname == "" {
		err := errors.Create(errors.BadRequest, "missing parameters")
		if filename == "" {
			err.AppendDetails(errors.Info{
				Name:    "filename",
				Details: "required",
			})
		}

		if dirname == "" {
			err.AppendDetails(errors.Info{
				Name:    "new_filename",
				Details: "required",
			})
		}
		return err
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" || fPath == "" {
		return errors.Create(errors.BadRequest, "wrong path format")
	}

	sourceID, fPath = Split(dirname)
	if sourceID == "" {
		return errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.MoveFile(ctx, filename, dirname)
}

func (h *ParamsHandler) CopyFile(ctx context.Context, filename string, dirname string) error {
	if filename == "" || dirname == "" {
		err := errors.Create(errors.BadRequest, "missing parameters")
		if filename == "" {
			err.AppendDetails(errors.Info{
				Name:    "filename",
				Details: "required",
			})
		}

		if dirname == "" {
			err.AppendDetails(errors.Info{
				Name:    "copy_filename",
				Details: "required",
			})
		}
		return err
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" || fPath == "" {
		return errors.Create(errors.BadRequest, "wrong path format")
	}

	sourceID, fPath = Split(dirname)
	if sourceID == "" {
		return errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.MoveFile(ctx, filename, dirname)
}

func (h *ParamsHandler) OpenMultipartSession(ctx context.Context, filename string, info *MultipartSessionInfo) (string, error) {
	if filename == "" || info == nil {
		err := errors.Create(errors.BadRequest, "missing parameters")
		if filename == "" {
			err.AppendDetails(errors.Info{
				Name:    "filename",
				Details: "required",
			})
		}

		if info == nil {
			err.AppendDetails(errors.Info{
				Name:    "info",
				Details: "required",
			})
		}
		return "", err
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" || fPath == "" {
		return "", errors.Create(errors.BadRequest, "wrong path format")
	}

	return h.next.OpenMultipartSession(ctx, filename, info)
}

func (h *ParamsHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *ContentPartInfo) error {
	if sessionID == "" || content == nil || size == 0 || info == nil {
		err := errors.Create(errors.BadRequest, "missing parameters")
		if sessionID == "" {
			err.AppendDetails(errors.Info{
				Name:    "session_id",
				Details: "required",
			})
		}

		if info == nil {
			err.AppendDetails(errors.Info{
				Name:    "info",
				Details: "required",
			})
		}
		return err
	}
	return h.next.AddContentPart(ctx, sessionID, content, size, info)
}

func (h *ParamsHandler) CloseMultipartSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "session_id",
			Details: "required",
		})
	}
	return h.next.CloseMultipartSession(ctx, sessionID)
}
