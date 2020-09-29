package store

import (
	"database/sql"
	"github.com/omecodes/omestore/pb"
)

type dataCursor struct {
	rows   *sql.Rows
	filter pb.IDFilter
	count  int64
	index  int
	err    error
	next   *pb.Data
}

func (c *dataCursor) Walk() (bool, error) {
	for {
		if !c.rows.Next() {
			return false, nil
		}

		if c.index > 0 && int64(c.index) == c.count {
			return false, nil
		}

		c.next = new(pb.Data)
		err := c.rows.Scan(&c.next.ID, &c.next.CreatedBy, &c.next.CreatedAt, &c.next.Size, &c.next.Content)
		if err != nil {
			return false, err
		}
		if c.filter != nil {
			ok, err := c.filter.Filter(c.next.ID)
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

func (c *dataCursor) Get() *pb.Data {
	return c.next
}

func (c *dataCursor) Close() error {
	return c.rows.Close()
}

func newDataCursor(rows *sql.Rows, filter pb.IDFilter, count int64) pb.DataCursor {
	return &dataCursor{
		rows:   rows,
		filter: filter,
		count:  count,
	}
}

type graftCursor struct {
	rows   *sql.Rows
	filter pb.IDFilter
	count  int64
	index  int
	err    error
	next   *pb.Graft
}

func (c *graftCursor) Walk() (bool, error) {
	for {
		if !c.rows.Next() {
			return false, nil
		}

		if c.index > 0 && int64(c.index) == c.count {
			return false, nil
		}

		c.next = new(pb.Graft)
		err := c.rows.Scan(&c.next.ID, &c.next.DataID, &c.next.CreatedBy, &c.next.CreatedAt, &c.next.Size, &c.next.Content)
		if err != nil {
			return false, err
		}

		if c.filter != nil {
			ok, err := c.filter.Filter(c.next.ID)
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

func (c *graftCursor) Get() *pb.Graft {
	return c.next
}

func (c *graftCursor) Close() error {
	return c.rows.Close()
}

func newGraftCursor(rows *sql.Rows, filter pb.IDFilter, count int64) pb.GraftCursor {
	return &graftCursor{
		rows:   rows,
		filter: filter,
		count:  count,
	}
}
