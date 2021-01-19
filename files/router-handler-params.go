package files

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/pb"
	"io"
	"path"
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
	if sourceID == "" {
		return errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return err
	}

	filename = path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)

	return h.next.CreateDir(ctx, filename)
}

func (h *ParamsHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts pb.PutFileOptions) error {
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
	if sourceID == "" {
		return errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return err
	}

	resolvedPath := path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)

	return h.next.WriteFileContent(ctx, resolvedPath, content, size, opts)
}

func (h *ParamsHandler) ListDir(ctx context.Context, dirname string, opts pb.ListDirOptions) (*pb.DirContent, error) {
	if dirname == "" {
		return nil, errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "dirname",
			Details: "required",
		})
	}

	sourceID, fPath := Split(dirname)
	if sourceID != "" {
		source, err := ResolveSource(ctx, sourceID)
		if err != nil {
			return nil, err
		}
		ctx = ContextWithSource(ctx, source)
		dirname = path.Join(source.URI, fPath)
	}

	return h.next.ListDir(ctx, dirname, opts)
}

func (h *ParamsHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	if filename == "" {
		return nil, 0, errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "filename",
			Details: "required",
		})
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" {
		return nil, 0, errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return nil, 0, err
	}

	resolvedPath := path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)

	return h.next.ReadFileContent(ctx, resolvedPath, opts)
}

func (h *ParamsHandler) GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error) {
	if filename == "" {
		return nil, errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "filename",
			Details: "required",
		})
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" {
		return nil, errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	resolvedPath := path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)

	return h.next.GetFileInfo(ctx, resolvedPath, opts)
}

func (h *ParamsHandler) DeleteFile(ctx context.Context, filename string, opts *pb.DeleteFileOptions) error {
	if filename == "" {
		return errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "filename",
			Details: "required",
		})
	}

	sourceID, fPath := Split(filename)
	if sourceID == "" {
		return errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return err
	}

	filename = path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)
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

	sourceID, fPath := Split(filename)
	if sourceID == "" {
		return errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return err
	}

	filename = path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)

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

	sourceID, fPath := Split(filename)
	if sourceID == "" {
		return nil, errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	filename = path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)

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
	if sourceID == "" {
		return errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return err
	}

	filename = path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)

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
	if sourceID == "" {
		return errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return err
	}

	filename = path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)

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
	if sourceID == "" {
		return errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return err
	}

	filename = path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)

	return h.next.MoveFile(ctx, filename, dirname)
}

func (h *ParamsHandler) OpenMultipartSession(ctx context.Context, filename string, info *pb.MultipartSessionInfo) (string, error) {
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
	if sourceID == "" {
		return "", errors.Create(errors.BadRequest, "missing source reference")
	}

	source, err := ResolveSource(ctx, sourceID)
	if err != nil {
		return "", err
	}

	filename = path.Join(source.URI, fPath)
	ctx = ContextWithSource(ctx, source)

	return h.next.OpenMultipartSession(ctx, filename, info)
}

func (h *ParamsHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error {
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

	source, err := ResolveSource(ctx, sessionID)
	if err != nil {
		return err
	}
	ctx = ContextWithSource(ctx, source)

	return h.next.AddContentPart(ctx, sessionID, content, size, info)
}

func (h *ParamsHandler) CloseMultipartSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "session_id",
			Details: "required",
		})
	}

	source, err := ResolveSource(ctx, sessionID)
	if err != nil {
		return err
	}
	ctx = ContextWithSource(ctx, source)

	return h.next.CloseMultipartSession(ctx, sessionID)
}
