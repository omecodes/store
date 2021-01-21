package files

import (
	"encoding/hex"
	"hash"
	"io"
)

func NewReaderHash(h hash.Hash) *readerHash {
	return &readerHash{
		result: "",
		h:      h,
	}
}

type readerHash struct {
	result string
	h      hash.Hash
	stream io.Reader
}

func (m *readerHash) Sum(b []byte) string {
	data := m.h.Sum(b)
	return hex.EncodeToString(data)
}

func (m *readerHash) Read(p []byte) (n int, err error) {
	n, err = m.stream.Read(p)
	m.h.Write(p[:n])
	return
}

func (m *readerHash) Reader(reader io.Reader) io.Reader {
	m.stream = reader
	return m
}

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
