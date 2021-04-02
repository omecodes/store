package files

import (
	"context"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/pkg/xattr"
)

type FS interface {
	Mkdir(ctx context.Context, dirname string) error
	Ls(ctx context.Context, dirname string, offset int, count int) (*DirContent, error)
	Write(ctx context.Context, filename string, content io.Reader, append bool) error
	Read(ctx context.Context, filename string, offset int64, count int64) (io.ReadCloser, int64, error)
	Info(ctx context.Context, filename string, withAttrs bool) (*File, error)
	SetAttributes(ctx context.Context, filename string, attrs Attributes) error
	GetAttributes(ctx context.Context, filename string, names ...string) (Attributes, error)
	Rename(ctx context.Context, filename string, newName string) error
	Move(ctx context.Context, filename string, dirname string) error
	Copy(ctx context.Context, filename string, dirname string) error
	DeleteFile(ctx context.Context, filename string, recursive bool) error
}

type ctxFsProvider struct{}

type FSProvider interface {
	GetFS(source *Source) (FS, error)
}

func getFS(ctx context.Context, sourceID string) (FS, error) {
	sourcesManager := getSourceManager(ctx)
	if sourcesManager == nil {
		return nil, errors.Internal("no source manager in context")
	}

	source, err := sourcesManager.Get(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	logs.Info("FS: resolved source", logs.Details("uri", source.Uri))

	o := ctx.Value(ctxFsProvider{})
	if o != nil {
		provider := o.(FSProvider)
		return provider.GetFS(source)
	}

	if source.Type != SourceType_Default {
		return nil, errors.Unsupported("file source type is not supported")
	}

	uri, err := url.Parse(source.Uri)
	if err != nil {
		return nil, err
	}

	switch uri.Scheme {
	case SchemeFS:
		rootDir := strings.TrimPrefix(source.Uri, SchemeFS+"://")
		return &diskFS{root: rootDir}, nil

	default:
		return nil, errors.BadRequest("not supported scheme")
	}
}

type diskFS struct {
	root string
}

func (d *diskFS) Mkdir(_ context.Context, dirname string) error {
	fullDirname := filepath.Join(d.root, dirname)
	logs.Info("FS: Create Dir", logs.Details("path", fullDirname))

	denormalizedPath := UnNormalizePath(fullDirname)
	logs.Info("FS: Denormalized path", logs.Details("path", denormalizedPath))

	err := os.Mkdir(denormalizedPath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		logs.Error("failed to create directory", logs.Details("file", dirname), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.NotFound("file not found",
				errors.Details{Key: "file", Value: dirname},
			)
		}

		if os.IsPermission(err) {
			return errors.Internal("file not found",
				errors.Details{Key: "system", Value: "Store app has no READ permissions on FS"},
				errors.Details{Key: "file", Value: dirname},
			)
		}
		return errors.Internal("could not create directory")
	}
	return nil
}

func (d *diskFS) Ls(_ context.Context, dirname string, offset int, count int) (*DirContent, error) {
	fullDirname := filepath.Join(d.root, dirname)
	logs.Info("FS: List Dir", logs.Details("path", fullDirname))

	f, err := os.Open(UnNormalizePath(fullDirname))
	if err != nil {
		logs.Error("failed to open file", logs.Details("file", dirname), logs.Err(err))

		if os.IsNotExist(err) {
			return nil, errors.NotFound("file not found",
				errors.Details{Key: "file", Value: dirname},
			)
		}

		if os.IsPermission(err) {
			return nil, errors.Internal("file not found",
				errors.Details{Key: "system", Value: "Store app has no READ permissions on FS"},
				errors.Details{Key: "file", Value: dirname},
			)
		}
		return nil, errors.Internal("could not get file info", errors.Details{Key: "file", Value: dirname})
	}

	defer func() {
		_ = f.Close()
	}()

	names, err := f.Readdirnames(-1)
	if err != nil {
		logs.Error("failed to get children names", logs.Details("file", dirname), logs.Err(err))
		return nil, errors.Internal("failed to get children names", errors.Details{Key: "file", Value: dirname})
	}

	dirContent := &DirContent{
		Total: len(names),
	}

	for ind, name := range names {
		if ind >= offset && len(dirContent.Files) < count {
			stats, err := os.Stat(filepath.Join(dirname, name))
			if err != nil {
				logs.Error("failed to get file stats", logs.Details("file", name), logs.Err(err))
				continue
			}

			f := &File{
				Name:     name,
				IsDir:    stats.IsDir(),
				Size:     stats.Size(),
				EditTime: stats.ModTime().Unix(),
			}
			dirContent.Files = append(dirContent.Files, f)
		}
	}

	return dirContent, nil
}

func (d *diskFS) Write(_ context.Context, filename string, content io.Reader, append bool) error {
	fullFilename := filepath.Join(d.root, filename)

	flags := os.O_CREATE | os.O_WRONLY
	if append {
		flags |= os.O_APPEND
	}

	file, err := os.OpenFile(UnNormalizePath(fullFilename), flags, os.ModePerm)
	if err != nil {
		logs.Error("failed to open file", logs.Details("file", filename), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.NotFound("failed to open file",
				errors.Details{Key: "file", Value: filename},
			)
		}

		if os.IsPermission(err) {
			return errors.Internal("failed to open file",
				errors.Details{Key: "system", Value: "Store app has no READ permissions on FS"},
				errors.Details{Key: "file", Value: filename},
			)
		}

		return errors.Internal("could not get file info", errors.Details{Key: "file", Value: filename}, errors.Details{Key: "open", Value: err.Error()})
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
	return nil
}

func (d *diskFS) Read(_ context.Context, filename string, offset int64, length int64) (io.ReadCloser, int64, error) {
	fullFilename := filepath.Join(d.root, filename)

	f, err := os.Open(UnNormalizePath(fullFilename))
	if err != nil {
		logs.Error("failed to open file", logs.Details("file", filename), logs.Err(err))
		return nil, 0, errors.Internal("failed to open file",
			errors.Details{Key: "file", Value: filename})
	}

	stats, err := f.Stat()
	if err != nil {
		logs.Error("failed to get file stats", logs.Details("file", filename), logs.Err(err))
		return nil, 0, errors.Internal("could not get file info", errors.Details{Key: "file", Value: filename})
	}

	if offset > 0 {
		_, err = f.Seek(offset, io.SeekStart)
		if err != nil {
			logs.Error("failed to perform seek on file", logs.Details("file", filename), logs.Err(err))
			return nil, 0, errors.Internal("seek failed on file",
				errors.Details{Key: "file", Value: filename})
		}
	}

	if length > 0 {
		return LimitReadCloser(f, length), stats.Size(), nil
	}
	return f, stats.Size(), nil
}

func (d *diskFS) Info(_ context.Context, filename string, withAttrs bool) (*File, error) {
	fullFilename := filepath.Join(d.root, filename)

	stats, err := os.Stat(UnNormalizePath(fullFilename))
	if err != nil {
		logs.Error("failed to get file stats", logs.Details("file", filename), logs.Err(err))

		if os.IsNotExist(err) {
			return nil, errors.NotFound("file not found",
				errors.Details{Key: "file", Value: filename},
			)
		}

		if os.IsPermission(err) {
			return nil, errors.Internal("file not found",
				errors.Details{Key: "system", Value: "Store app has no READ permissions on FS"},
				errors.Details{Key: "file", Value: filename},
			)
		}

		return nil, errors.Internal("could not get file info", errors.Details{Key: "file", Value: filename})
	}

	file := &File{
		Name:     path.Base(filename),
		IsDir:    stats.IsDir(),
		Size:     stats.Size(),
		EditTime: stats.ModTime().Unix(),
	}

	if withAttrs {
		file.Attributes = Attributes{}

		attrsName, err := xattr.List(fullFilename)
		if err != nil {
			logs.Error("failed to get file attributes names", logs.Details("file", filename), logs.Err(err))
			return nil, errors.Internal("could not get file attributes", errors.Details{Key: "file", Value: filename})
		}

		for _, name := range attrsName {
			if strings.HasPrefix(name, AttrPrefix) {
				attrsBytes, err := xattr.Get(fullFilename, name)
				if err != nil {
					logs.Error("failed to get file attribute", logs.Details("file", filename), logs.Details("attribute", name), logs.Err(err))
					return nil, errors.Internal("could not get file attribute", errors.Details{Key: "file", Value: filename})
				}
				file.Attributes[name] = string(attrsBytes)
			}
		}
	}
	return file, nil
}

func (d *diskFS) SetAttributes(_ context.Context, filename string, attrs Attributes) error {
	fullFilename := filepath.Join(d.root, filename)
	for name, value := range attrs {
		err := xattr.Set(UnNormalizePath(fullFilename), name, []byte(value))
		if err != nil {
			logs.Error("failed to get file attributes names", logs.Details("file", filename), logs.Err(err))
			return errors.Internal("failed to set file attribute",
				errors.Details{Key: name, Value: value},
				errors.Details{Key: "file", Value: filename})
		}
	}
	return nil
}

func (d *diskFS) GetAttributes(_ context.Context, filename string, names ...string) (Attributes, error) {
	fullFilename := filepath.Join(d.root, filename)
	logs.Info("FS: Get file attributes", logs.Details("path", fullFilename))

	resolvedFilename := UnNormalizePath(fullFilename)

	attributeNames, err := xattr.List(resolvedFilename)
	if err != nil {
		return nil, err
	}

	var intersection []string
	for _, attrName := range attributeNames {
		for _, name := range names {
			if name == attrName {
				intersection = append(intersection, name)
			}
		}
	}

	attributes := Attributes{}
	for _, name := range intersection {
		if strings.HasPrefix(name, AttrPrefix) {
			attrsBytes, err := xattr.Get(filename, name)
			if err != nil {
				logs.Error("failed to get file attribute", logs.Details("file", resolvedFilename), logs.Details("attribute", name), logs.Err(err))
				return nil, errors.Internal("could not get file attribute", errors.Details{Key: "file", Value: filename})
			}
			attributes[name] = string(attrsBytes)
		}
	}

	return attributes, nil
}

func (d *diskFS) Rename(_ context.Context, filename string, newName string) error {
	fullFilename := filepath.Join(d.root, filename)
	newPath := filepath.Join(UnNormalizePath(fullFilename), newName)
	err := os.Rename(UnNormalizePath(fullFilename), newPath)
	if err != nil {
		logs.Error("failed to rename file", logs.Details("file", filename), logs.Details("new name", newName), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.NotFound("file not found",
				errors.Details{Key: "file", Value: filename},
			)
		}

		if os.IsPermission(err) {
			return errors.Internal("permission denied for applicaiton",
				errors.Details{Key: "system", Value: "disk permissions denied for store application"},
				errors.Details{Key: "file", Value: filename},
			)
		}
		return errors.Internal("could not rename file", errors.Details{Key: "file", Value: filename}, errors.Details{Key: "new name", Value: newName})
	}
	return nil
}

func (d *diskFS) Move(_ context.Context, filename string, dirname string) error {
	fullFilename := filepath.Join(d.root, filename)
	newPath := UnNormalizePath(filepath.Join(d.root, dirname))

	err := os.Rename(UnNormalizePath(fullFilename), newPath)
	if err != nil {
		logs.Error("failed to move file", logs.Details("file", filename), logs.Details("directory", dirname), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.NotFound("file not found",
				errors.Details{Key: "file", Value: filename},
			)
		}

		if os.IsPermission(err) {
			return errors.Internal("permission denied for application",
				errors.Details{Key: "system", Value: "disk permissions denied for store application"},
				errors.Details{Key: "file", Value: filename},
			)
		}
		return errors.Internal("could not rename file", errors.Details{Key: "file", Value: filename}, errors.Details{Key: "directory", Value: dirname})
	}

	return nil
}

func (d *diskFS) Copy(_ context.Context, filename string, dirname string) error {
	fullFilename := filepath.Join(d.root, filename)
	newPath := UnNormalizePath(filepath.Join(d.root, dirname, filepath.Base(filename)))

	err := os.Rename(UnNormalizePath(fullFilename), newPath)
	if err != nil {
		logs.Error("failed to copy file", logs.Details("file", filename), logs.Details("directory", dirname), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.NotFound("file not found",
				errors.Details{Key: "file", Value: filename},
			)
		}

		if os.IsPermission(err) {
			return errors.Internal("permission denied for application",
				errors.Details{Key: "system", Value: "disk permissions denied for store application"},
				errors.Details{Key: "file", Value: filename},
			)
		}
		return errors.Internal("could not copy file", errors.Details{Key: "file", Value: filename}, errors.Details{Key: "directory", Value: dirname})
	}

	return nil
}

func (d *diskFS) DeleteFile(_ context.Context, filename string, recursive bool) error {
	fullDirname := filepath.Join(d.root, filename)

	var err error

	if recursive {
		err = os.RemoveAll(UnNormalizePath(fullDirname))
	} else {
		err = os.Remove(UnNormalizePath(fullDirname))
	}

	if err != nil {
		logs.Error("failed to delete file", logs.Details("file", filename), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.NotFound("file not found",
				errors.Details{Key: "file", Value: filename},
			)
		}

		if os.IsPermission(err) {
			return errors.Internal("file not found",
				errors.Details{Key: "system", Value: "Store app has no READ permissions on FS"},
				errors.Details{Key: "file", Value: filename},
			)
		}

		return errors.Internal("could not delete directory", errors.Details{Key: "file", Value: filename})
	}
	return nil
}
