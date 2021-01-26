package webapp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
)

var Dir string

var mimes = map[string]string{
	".js":   "text/javascript",
	".html": "text/html",
	".csh":  "text/x-script.csh",
	".css":  "text/css",
	".svg":  "image/svg+xml",
}

func ServeApps(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path

	contentType, f, size, err := getFileContent(r.Context(), filename)
	if err != nil {
		w.WriteHeader(errors.HTTPStatus(err))
		return
	}

	defer func() {
		if err := f.Close(); err != nil {
			logs.Error("file close", logs.Err(err))
		}
	}()

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.WriteHeader(http.StatusOK)

	done := false
	buffer := make([]byte, 1024)
	for !done {
		n, err := f.Read(buffer)
		if err != nil {
			if done = err == io.EOF; !done {
				return
			}
		}

		_, err = w.Write(buffer[:n])
		if err != nil {
			return
		}
	}
}

func getFileContent(_ context.Context, filename string) (string, io.ReadCloser, int64, error) {

	if path.Ext(filename) == "" {
		filename = path.Join(filename, "index.html")
	}

	filename = filepath.Join(Dir, filename)

	size := int64(0)
	contentType := ""
	extension := filepath.Ext(filename)
	if extension == "" {
		filename = path.Join(filename, "index.html")
		extension = "html"
	}

	f, err := os.Open(filename)
	if err != nil {
		if err == os.ErrNotExist {
			return "", f, size, errors.Create(errors.NotFound, err.Error())
		}
		return "", f, size, err
	}

	stat, err := f.Stat()
	if err != nil {
		return "", f, size, err
	}

	size = stat.Size()
	contentType, ok := mimes[extension]
	if !ok {
		extension = mimeFromFilename(filename)
	}

	return contentType, f, size, err
}

func mimeFromFilename(filename string) (contentType string) {
	contentType = "text/plain"
	fullFilename := filepath.Join(Dir, filename)

	f, err := os.Open(fullFilename)
	if err != nil {
		return
	}

	defer func() {
		if err := f.Close(); err != nil {
			logs.Error("file close", logs.Err(err))
		}
	}()

	buffer := make([]byte, 512)

	_, err = f.Read(buffer)
	if err != nil {
		return
	}

	contentType = http.DetectContentType(buffer)
	return
}
