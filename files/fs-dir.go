package files

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/pkg/xattr"
)

type dirFS struct {
	root string
}

func (d *dirFS) Mkdir(_ context.Context, dirname string) error {
	fullDirname := filepath.Join(d.root, dirname)

	err := os.MkdirAll(UnNormalizePath(fullDirname), os.ModePerm)
	if err != nil {
		logs.Error("failed to create directory", logs.Details("file", dirname), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.Create(errors.NotFound, "file not found",
				errors.Info{Name: "file", Details: dirname},
			)
		}

		if os.IsPermission(err) {
			return errors.Create(errors.Internal, "file not found",
				errors.Info{Name: "system", Details: "Store app has no READ permissions on FS"},
				errors.Info{Name: "file", Details: dirname},
			)
		}
		return errors.Create(errors.Internal, "could not get file info", errors.Info{Name: "file", Details: dirname})
	}
	return nil
}

func (d *dirFS) Ls(_ context.Context, dirname string, offset int, count int) (*DirContent, error) {
	fullDirname := filepath.Join(d.root, dirname)
	f, err := os.Open(UnNormalizePath(fullDirname))
	if err != nil {
		logs.Error("failed to open file", logs.Details("file", dirname), logs.Err(err))

		if os.IsNotExist(err) {
			return nil, errors.Create(errors.NotFound, "file not found",
				errors.Info{Name: "file", Details: dirname},
			)
		}

		if os.IsPermission(err) {
			return nil, errors.Create(errors.Internal, "file not found",
				errors.Info{Name: "system", Details: "Store app has no READ permissions on FS"},
				errors.Info{Name: "file", Details: dirname},
			)
		}
		return nil, errors.Create(errors.Internal, "could not get file info", errors.Info{Name: "file", Details: dirname})
	}

	defer func() {
		_ = f.Close()
	}()

	names, err := f.Readdirnames(-1)
	if err != nil {
		logs.Error("failed to get children names", logs.Details("file", dirname), logs.Err(err))
		return nil, errors.Create(errors.Internal, "failed to get children names", errors.Info{Name: "file", Details: dirname})
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

func (d *dirFS) Write(_ context.Context, filename string, content io.Reader, append bool) error {
	fullFilename := filepath.Join(d.root, filename)

	flags := os.O_CREATE | os.O_WRONLY
	if append {
		flags |= os.O_APPEND
	}

	file, err := os.OpenFile(UnNormalizePath(fullFilename), flags, os.ModePerm)
	if err != nil {
		logs.Error("failed to open file", logs.Details("file", filename), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.Create(errors.NotFound, "failed to open file",
				errors.Info{Name: "file", Details: filename},
			)
		}

		if os.IsPermission(err) {
			return errors.Create(errors.Internal, "failed to open file",
				errors.Info{Name: "system", Details: "Store app has no READ permissions on FS"},
				errors.Info{Name: "file", Details: filename},
			)
		}

		return errors.Create(errors.Internal, "could not get file info", errors.Info{Name: "file", Details: filename}, errors.Info{Name: "open", Details: err.Error()})
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

func (d *dirFS) Read(_ context.Context, filename string, offset int64, length int64) (io.ReadCloser, int64, error) {
	fullFilename := filepath.Join(d.root, filename)

	f, err := os.Open(UnNormalizePath(fullFilename))
	if err != nil {
		logs.Error("failed to open file", logs.Details("file", filename), logs.Err(err))
		return nil, 0, errors.Create(errors.Internal, "failed to open file",
			errors.Info{Name: "file", Details: filename})
	}

	stats, err := f.Stat()
	if err != nil {
		logs.Error("failed to get file stats", logs.Details("file", filename), logs.Err(err))
		return nil, 0, errors.Create(errors.Internal, "could not get file info", errors.Info{Name: "file", Details: filename})
	}

	if offset > 0 {
		_, err = f.Seek(offset, io.SeekStart)
		if err != nil {
			logs.Error("failed to perform seek on file", logs.Details("file", filename), logs.Err(err))
			return nil, 0, errors.Create(errors.Internal, "seek failed on file",
				errors.Info{Name: "file", Details: filename})
		}
	}

	if length > 0 {
		return LimitReadCloser(f, length), stats.Size(), nil
	}
	return f, stats.Size(), nil
}

func (d *dirFS) Info(_ context.Context, filename string, withAttrs bool) (*File, error) {
	fullFilename := filepath.Join(d.root, filename)

	stats, err := os.Stat(UnNormalizePath(fullFilename))
	if err != nil {
		logs.Error("failed to get file stats", logs.Details("file", filename), logs.Err(err))

		if os.IsNotExist(err) {
			return nil, errors.Create(errors.NotFound, "file not found",
				errors.Info{Name: "file", Details: filename},
			)
		}

		if os.IsPermission(err) {
			return nil, errors.Create(errors.Internal, "file not found",
				errors.Info{Name: "system", Details: "Store app has no READ permissions on FS"},
				errors.Info{Name: "file", Details: filename},
			)
		}

		return nil, errors.Create(errors.Internal, "could not get file info", errors.Info{Name: "file", Details: filename})
	}

	file := &File{
		Name:    path.Base(filename),
		IsDir:   stats.IsDir(),
		Size:    stats.Size(),
		ModTime: stats.ModTime().Unix(),
	}

	if withAttrs {
		file.Attributes = Attributes{}

		attrsName, err := xattr.List(fullFilename)
		if err != nil {
			logs.Error("failed to get file attributes names", logs.Details("file", filename), logs.Err(err))
			return nil, errors.Create(errors.Internal, "could not get file attributes", errors.Info{Name: "file", Details: filename})
		}

		for _, name := range attrsName {
			if strings.HasPrefix(name, AttrPrefix) {
				attrsBytes, err := xattr.Get(fullFilename, name)
				if err != nil {
					logs.Error("failed to get file attribute", logs.Details("file", filename), logs.Details("attribute", name), logs.Err(err))
					return nil, errors.Create(errors.Internal, "could not get file attribute", errors.Info{Name: "file", Details: filename})
				}
				file.Attributes[name] = string(attrsBytes)
			}
		}
	}
	return file, nil
}

func (d *dirFS) SetAttributes(_ context.Context, filename string, attrs Attributes) error {
	fullFilename := filepath.Join(d.root, filename)
	for name, value := range attrs {
		err := xattr.Set(UnNormalizePath(fullFilename), name, []byte(value))
		if err != nil {
			logs.Error("failed to get file attributes names", logs.Details("file", filename), logs.Err(err))
			return errors.Create(errors.Internal, "failed to set file attribute",
				errors.Info{Name: name, Details: value},
				errors.Info{Name: "file", Details: filename})
		}
	}
	return nil
}

func (d *dirFS) GetAttributes(_ context.Context, filename string, names ...string) (Attributes, error) {
	fullFilename := filepath.Join(d.root, filename)
	logs.Info("Getting file attributes", logs.Details("path", fullFilename))

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
				return nil, errors.Create(errors.Internal, "could not get file attribute", errors.Info{Name: "file", Details: filename})
			}
			attributes[name] = string(attrsBytes)
		}
	}

	return attributes, nil
}

func (d *dirFS) Rename(_ context.Context, filename string, newName string) error {
	fullFilename := filepath.Join(d.root, filename)
	newPath := filepath.Join(UnNormalizePath(fullFilename), newName)
	err := os.Rename(UnNormalizePath(fullFilename), newPath)
	if err != nil {
		logs.Error("failed to rename file", logs.Details("file", filename), logs.Details("new name", newName), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.Create(errors.NotFound, "file not found",
				errors.Info{Name: "file", Details: filename},
			)
		}

		if os.IsPermission(err) {
			return errors.Create(errors.Internal, "permission denied for applicaiton",
				errors.Info{Name: "system", Details: "disk permissions denied for store application"},
				errors.Info{Name: "file", Details: filename},
			)
		}
		return errors.Create(errors.Internal, "could not rename file", errors.Info{Name: "file", Details: filename}, errors.Info{Name: "new name", Details: newName})
	}
	return nil
}

func (d *dirFS) Move(_ context.Context, filename string, dirname string) error {
	fullFilename := filepath.Join(d.root, filename)
	newPath := UnNormalizePath(filepath.Join(d.root, dirname))

	err := os.Rename(UnNormalizePath(fullFilename), newPath)
	if err != nil {
		logs.Error("failed to move file", logs.Details("file", filename), logs.Details("directory", dirname), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.Create(errors.NotFound, "file not found",
				errors.Info{Name: "file", Details: filename},
			)
		}

		if os.IsPermission(err) {
			return errors.Create(errors.Internal, "permission denied for application",
				errors.Info{Name: "system", Details: "disk permissions denied for store application"},
				errors.Info{Name: "file", Details: filename},
			)
		}
		return errors.Create(errors.Internal, "could not rename file", errors.Info{Name: "file", Details: filename}, errors.Info{Name: "directory", Details: dirname})
	}

	return nil
}

func (d *dirFS) Copy(_ context.Context, filename string, dirname string) error {
	fullFilename := filepath.Join(d.root, filename)
	newPath := UnNormalizePath(filepath.Join(d.root, dirname, filepath.Base(filename)))

	err := os.Rename(UnNormalizePath(fullFilename), newPath)
	if err != nil {
		logs.Error("failed to copy file", logs.Details("file", filename), logs.Details("directory", dirname), logs.Err(err))

		if os.IsNotExist(err) {
			return errors.Create(errors.NotFound, "file not found",
				errors.Info{Name: "file", Details: filename},
			)
		}

		if os.IsPermission(err) {
			return errors.Create(errors.Internal, "permission denied for application",
				errors.Info{Name: "system", Details: "disk permissions denied for store application"},
				errors.Info{Name: "file", Details: filename},
			)
		}
		return errors.Create(errors.Internal, "could not copy file", errors.Info{Name: "file", Details: filename}, errors.Info{Name: "directory", Details: dirname})
	}

	return nil
}

func (d *dirFS) DeleteFile(_ context.Context, filename string, recursive bool) error {
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
			return errors.Create(errors.NotFound, "file not found",
				errors.Info{Name: "file", Details: filename},
			)
		}

		if os.IsPermission(err) {
			return errors.Create(errors.Internal, "file not found",
				errors.Info{Name: "system", Details: "Store app has no READ permissions on FS"},
				errors.Info{Name: "file", Details: filename},
			)
		}

		return errors.Create(errors.Internal, "could not delete directory", errors.Info{Name: "file", Details: filename})
	}
	return nil
}
