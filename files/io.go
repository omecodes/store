package files

import (
	"io"
)

func LimitReadCloser(readCloser io.ReadCloser, n int64) io.ReadCloser {
	return &limitReadCloser{
		reader: io.LimitReader(readCloser, n),
		closer: readCloser,
	}
}

type limitReadCloser struct {
	reader io.Reader
	closer io.Closer
}

func (lrc *limitReadCloser) Read(buf []byte) (int, error) {
	return lrc.reader.Read(buf)
}

func (lrc *limitReadCloser) Close() error {
	return lrc.closer.Close()
}
