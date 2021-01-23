package webapp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
)

var (
	globalMutex  = &sync.RWMutex{}
	watchEnabled bool
	watcher      *fsnotify.Watcher
	Dir          string
	dirs         map[string]string
)

func addAppDir(dir string) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	dirs[dir] = ""
}

func removeAppDir(dir string) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	delete(dirs, dir)
}

func appList() []string {
	globalMutex.RLock()
	defer globalMutex.RUnlock()

	var list []string
	for k, _ := range dirs {
		list = append(list, k)
	}
	return list
}

func WatchDir() {
	watchEnabled = true
	defer func() {
		logs.Info("Webapp • app directory watching stopped")
	}()

	var err error
	for watchEnabled {
		watcher, err = fsnotify.NewWatcher()
		if err != nil {
			logs.Error("Webapp • could not create watcher", logs.Err(err))
			<-time.After(time.Second * 3)
			continue
		}

		err = watcher.Add(Dir)
		if err != nil {
			logs.Error("Webapp • could not watch", logs.Details("dir", Dir), logs.Err(err))
			<-time.After(time.Second * 3)
			continue
		}

		done := false
		for !done {
			select {
			case event, open := <-watcher.Events:
				if done = !open; done {
					break
				}

				if event.Op == fsnotify.Create {
					filename := filepath.Join(Dir, event.Name)
					stats, err := os.Stat(filename)
					if err == nil {
						if stats.IsDir() {
							addAppDir(filepath.Base(filename))
						}
					}

				} else if event.Op == fsnotify.Remove {
					filename := filepath.Join(Dir, event.Name)
					stats, err := os.Stat(filename)
					if err == nil {
						if stats.IsDir() {
							removeAppDir(filepath.Base(filename))
						}
					}
				}

			case err, _ := <-watcher.Errors:
				done = true
				logs.Error("Webapp • dir watch", logs.Err(err))
				break
			}
		}
	}
}

func StopWatch() {
	watchEnabled = false
	if watcher != nil {
		if err := watcher.Close(); err != nil {
			logs.Error("close watcher", logs.Err(err))
		}
	}
}

var mimes = map[string]string{
	".js":   "text/javascript",
	".html": "text/html",
	".csh":  "text/x-script.csh",
	".css":  "text/css",
	".svg":  "image/svg+xml",
}

func ServeApps(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path
	for _, routerPath := range appList() {
		if strings.HasPrefix(filename, routerPath) {
			filename = strings.Replace(filename, routerPath, "", 1)
			break
		}
	}

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

	if !strings.HasPrefix(filename, "/") {
		filename = "/" + filename
	}

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
