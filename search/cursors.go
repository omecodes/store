package se

import (
	"github.com/omecodes/bome"
	"io"
	"strings"
)

type aggregatedStrIdsCursor struct {
	cursor      bome.Cursor
	currentList []string
}

func (c *aggregatedStrIdsCursor) Next() (string, error) {
	for {
		if len(c.currentList) == 0 {
			if !c.cursor.HasNext() {
				return "", io.EOF
			}

			o, err := c.cursor.Next()
			if err != nil {
				return "", err
			}

			c.currentList = strings.Split(o.(string), "<>")
		}

		next := strings.Trim(c.currentList[0], " ")
		if next == "" {
			continue
		}

		c.currentList = c.currentList[1:]
		return next, nil
	}
}

func (c *aggregatedStrIdsCursor) Close() error {
	return nil
}

type dbStringCursorWrapper struct {
	cursor bome.Cursor
}

func (c *dbStringCursorWrapper) Next() (string, error) {
	if c.cursor.HasNext() {
		o, err := c.cursor.Next()
		if err == nil {
			return o.(string), nil
		}
		return "", err
	}
	return "", io.EOF
}

func (c *dbStringCursorWrapper) Close() error {
	return c.cursor.Close()
}

type dbMapEntryCursorWrapper struct {
	cursor bome.Cursor
}

func (c *dbMapEntryCursorWrapper) Next() (string, error) {
	if c.cursor.HasNext() {
		o, err := c.cursor.Next()
		if err == nil {
			entry := o.(*bome.MapEntry)
			return entry.Value, nil
		}
		return "", err
	}
	return "", io.EOF
}

func (c *dbMapEntryCursorWrapper) Close() error {
	return c.cursor.Close()
}

type idListCursor struct {
	ids []string
	pos int
}

func (c *idListCursor) Next() (string, error) {
	if c.pos < len(c.ids) {
		id := c.ids[c.pos]
		c.pos++
		return id, nil
	}
	return "", io.EOF
}

func (c *idListCursor) Close() error {
	return nil
}
