package router

import (
	"context"
	"fmt"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/pb"
	"io"
	"os"
	"strings"
)

type FilesExecHandler struct {
	FilesBaseObjectsHandler
}

func (h *FilesExecHandler) CreateDir(ctx context.Context, filename string) error {
	source := files.GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	if source.Type != files.TypeFS {
		return errors.Create(errors.NotImplemented, "file source type is not supported")
	}

	prefix := fmt.Sprintf("%s://", files.SchemeFS)
	filename = strings.TrimPrefix(filename, prefix)

	return os.MkdirAll(filename, os.ModePerm)
}

func (h *FilesExecHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts pb.PutFileOptions) error {
	source := files.GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	if source.Type != files.TypeFS {
		return errors.Create(errors.NotImplemented, "file source type is not supported")
	}

	prefix := fmt.Sprintf("%s://", files.SchemeFS)
	filename = strings.TrimPrefix(filename, prefix)

	flags := os.O_CREATE | os.O_WRONLY
	if opts.Append {
		flags |= os.O_APPEND
	}

	file, err := os.OpenFile(filename, flags, os.ModePerm)
	if err != nil {
		return errors.Create(errors.Internal, "can open file", errors.Info{Name: "open", Details: err.Error()})
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			logs.Error("file descriptor close", logs.Err(err))
		}
	}()

	buf := make([]byte, 1024)
	total := 0
	done := false
	for !done {
		n, err := content.Read(buf)
		if err != nil {
			if done = err == io.EOF; !done {
				return err
			}
		}
		total += n
		n, err = file.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	if opts.Permissions != nil {
		return files.SetPermissions(filename, opts.Permissions)
	}

	return nil
}

func (h *FilesExecHandler) ListDir(ctx context.Context, dirname string, opts pb.GetFileInfoOptions) ([]*pb.File, error) {
	if dirname == "" {
		return nil, errors.Create(errors.BadRequest, "missing parameters", errors.Info{
			Name:    "dirname",
			Details: "required",
		})
	}
	return h.next.ListDir(ctx, dirname, opts)
}

func (h *FilesExecHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	panic("implement me")
}

func (h *FilesExecHandler) GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error) {
	panic("implement me")
}

func (h *FilesExecHandler) DeleteFile(ctx context.Context, filename string) error {
	panic("implement me")
}

func (h *FilesExecHandler) SetFileMetaData(ctx context.Context, filename string, name files.AttrName, value string) error {
	panic("implement me")
}

func (h *FilesExecHandler) GetFileMetaData(ctx context.Context, filename string, name files.AttrName) (string, error) {
	panic("implement me")
}

func (h *FilesExecHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	panic("implement me")
}

func (h *FilesExecHandler) MoveFile(ctx context.Context, srcFilename string, dstFilename string) error {
	panic("implement me")
}

func (h *FilesExecHandler) CopyFile(ctx context.Context, srcFilename string, dstFilename string) error {
	panic("implement me")
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
