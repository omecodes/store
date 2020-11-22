package oms

import (
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

		c.next = new(Object)
		err := c.rows.Scan(&c.next.Id, &c.next.CreatedBy, &c.next.CreatedAt, &c.next.Size, &c.next.JsonEncoded)
		if err != nil {
			return false, err
		}
		if c.filter != nil {
			ok, err := c.filter.Filter(c.next.Id)
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

func newDataCursor(rows *sql.Rows, filter IDFilter, count int64) DataCursor {
	return &dataCursor{
		rows:   rows,
		filter: filter,
		count:  count,
	}
}

type graftCursor struct {
	rows   *sql.Rows
	filter IDFilter
	count  int64
	index  int
	err    error
	next   *Graft
}

func (c *graftCursor) Walk() (bool, error) {
	for {
		if !c.rows.Next() {
			return false, nil
		}

		if c.index > 0 && int64(c.index) == c.count {
			return false, nil
		}

		c.next = new(Graft)
		err := c.rows.Scan(&c.next.Id, &c.next.DataId, &c.next.CreatedBy, &c.next.CreatedAt, &c.next.Size, &c.next.Content)
		if err != nil {
			return false, err
		}

		if c.filter != nil {
			ok, err := c.filter.Filter(c.next.Id)
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

func (c *graftCursor) Get() *Graft {
	return c.next
}

func (c *graftCursor) Close() error {
	return c.rows.Close()
}

func newGraftCursor(rows *sql.Rows, filter IDFilter, count int64) GraftCursor {
	return &graftCursor{
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

type GraftCursor interface {
	Walk() (bool, error)
	Get() *Graft
	Close() error
}
