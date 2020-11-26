package oms

import (
	"bytes"
	"database/sql"
)

type dataCursor struct {
	rows   *sql.Rows
	filter IDFilter
	count  int64
	index  int
	err    error
	next   *Object
}

func (c *dataCursor) Walk() (bool, error) {
	for {
		if !c.rows.Next() {
			return false, nil
		}

		if c.index > 0 && int64(c.index) == c.count {
			return false, nil
		}

		var value string
		err := c.rows.Scan(&value)
		if err != nil {
			return false, err
		}

		c.next = NewObject()
		c.next.SetContent(bytes.NewBuffer([]byte(value)), int64(len(value)))

		if c.filter != nil {
			ok, err := c.filter.Filter(c.next.ID())
			if err != nil {
				return false, err
			}
			if !ok {
				continue
			}
		}
		c.index++
		return true, err
	}
}

func (c *dataCursor) Get() *Object {
	return c.next
}

func (c *dataCursor) Close() error {
	return c.rows.Close()
}

func NewDataCursor(rows *sql.Rows, filter IDFilter, count int64) DataCursor {
	return &dataCursor{
		rows:   rows,
		filter: filter,
		count:  count,
	}
}

type DataCursor interface {
	Walk() (bool, error)
	Get() *Object
	Close() error
}
