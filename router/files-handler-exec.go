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
	"path/filepath"
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

func (h *FilesExecHandler) ListDir(ctx context.Context, dirname string, opts pb.ListDirOptions) (*pb.DirContent, error) {
	source := files.GetSource(ctx)
	if source == nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	if source.Type != files.TypeFS {
		return nil, errors.Create(errors.NotImplemented, "file source type is not supported")
	}

	prefix := fmt.Sprintf("%s://", files.SchemeFS)
	dirname = strings.TrimPrefix(dirname, prefix)

	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = f.Close()
	}()

	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	dirContent := &pb.DirContent{
		Total: len(names),
	}

	for ind, name := range names {
		if ind >= opts.Offset && len(dirContent.Files) < opts.Count {
			stats, err := os.Stat(filepath.Join(dirname, name))
			if err != nil {
				continue
			}

			f := &pb.File{
				Name:    name,
				IsDir:   stats.IsDir(),
				Size:    stats.Size(),
				ModTime: stats.ModTime().Unix(),
			}
			dirContent.Files = append(dirContent.Files, f)
		}
	}

	return dirContent, nil
}

func (h *FilesExecHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	source := files.GetSource(ctx)
	if source == nil {
		return nil, 0, errors.Create(errors.Internal, "missing source in context")
	}

	if source.Type != files.TypeFS {
		return nil, 0, errors.Create(errors.NotImplemented, "file source type is not supported")
	}

	prefix := fmt.Sprintf("%s://", files.SchemeFS)
	filename = strings.TrimPrefix(filename, prefix)

	f, err := os.Open(filename)
	if err != nil {
		return nil, 0, err
	}

	stats, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}

	if opts.Range.Offset > 0 {
		_, err = f.Seek(opts.Range.Offset, io.SeekStart)
		if err != nil {
			return nil, 0, err
		}

		if opts.Range.Length > 0 {
			return files.LimitReadCloser(f, opts.Range.Length), stats.Size(), nil
		}
	}

	return f, stats.Size(), nil
}

func (h *FilesExecHandler) GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error) {
	source := files.GetSource(ctx)
	if source == nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	if source.Type != files.TypeFS {
		return nil, errors.Create(errors.NotImplemented, "file source type is not supported")
	}

	prefix := fmt.Sprintf("%s://", files.SchemeFS)
	filename = strings.TrimPrefix(filename, prefix)

	return nil, nil
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
