package objects

import (
	"io"
)

type idsListCursor struct {
	ids           []string
	getObjectFunc func(string) (*Object, error)
	pos           int
}

func (i *idsListCursor) Browse() (*Object, error) {
	if i.pos == len(i.ids) {
		return nil, io.EOF
	}

	id := i.ids[i.pos]
	i.pos++

	return i.getObjectFunc(id)
}

func (i *idsListCursor) Close() error {
	return nil
}

type Browser interface {
	Browse() (*Object, error)
}

type BrowseFunc func() (*Object, error)

func (f BrowseFunc) Browse() (*Object, error) {
	return f()
}

type Closer interface {
	Close() error
}

type CloseFunc func() error

func (f CloseFunc) Close() error {
	return f()
}

func NewCursor(browser Browser, closer Closer) *Cursor {
	return &Cursor{
		browser: browser,
		closer:  closer,
	}
}

type Cursor struct {
	browser Browser
	closer  Closer
}

func (c *Cursor) Browse() (*Object, error) {
	return c.browser.Browse()
}

func (c *Cursor) Close() error {
	return c.closer.Close()
}

func (c *Cursor) GetCloser() Closer {
	return c.closer
}

func (c *Cursor) GetBrowser() Browser {
	return c.browser
}

func (c *Cursor) SetCloser(closer Closer) {
	c.closer = closer
}

func (c *Cursor) SetBrowser(browser Browser) {
	c.browser = browser
}
