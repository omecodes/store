package objects

import (
	"context"
	"database/sql"
	"github.com/omecodes/bome"
	"github.com/omecodes/store/pb"
	"io"
)

func NewSQLCollection(db *sql.DB, dialect string, tableName string) (*sqlCollection, error) {
	var err error
	c := &sqlCollection{}
	c.objects, err = bome.NewJSONMap(db, dialect, tableName)
	if err != nil {
		return nil, err
	}

	c.datedRef, err = bome.NewList(db, dialect, tableName+"_dated_refs")
	return c, err
}

type sqlCollection struct {
	objects  *bome.JSONMap
	datedRef *bome.List
}

func (s *sqlCollection) Objects() *bome.JSONMap {
	return s.objects
}

func (s *sqlCollection) DatedRefs() *bome.List {
	return s.datedRef
}

func (s *sqlCollection) Save(ctx context.Context, createdAt int64, id string, data string) error {
	if bome.IsTransactionContext(ctx) {
		_, tx, err := s.objects.Transaction(ctx)
		if err != nil {
			return err
		}

		err = tx.Save(&bome.MapEntry{
			Key:   id,
			Value: data,
		})
		if err == nil {
			dtx := s.datedRef.ContinueTransaction(tx.TX())
			err = dtx.Save(&bome.ListEntry{
				Index: createdAt,
				Value: id,
			})
		}
		return err
	}

	err := s.objects.Save(&bome.MapEntry{
		Key:   id,
		Value: data,
	})
	if err == nil {
		err = s.datedRef.Save(&bome.ListEntry{
			Index: createdAt,
			Value: id,
		})
	}
	return err
}

func (s *sqlCollection) Select(ctx context.Context, headerResolver HeaderResolver, dataResolver DataResolver) (*pb.Cursor, error) {
	c, err := s.objects.List()
	if err != nil {
		return nil, err
	}

	closer := pb.CloseFunc(func() error {
		return c.Close()
	})

	browser := pb.BrowseFunc(func() (*pb.Object, error) {
		if !c.HasNext() {
			return nil, io.EOF
		}

		next, err := c.Next()
		if err != nil {
			return nil, err
		}

		entry := next.(*bome.MapEntry)
		o := &pb.Object{
			Header: &pb.Header{Id: entry.Key},
			Data:   entry.Value,
		}

		o.Header, err = headerResolver.ResolveHeader(entry.Key)
		if err != nil {
			return nil, err
		}

		if dataResolver != nil {
			o.Data, err = dataResolver.ResolveData(entry.Key)
		}
		return o, nil
	})
	return pb.NewCursor(browser, closer), nil
}

func (s *sqlCollection) RangeSelect(ctx context.Context, after int64, before int64, headerResolver HeaderResolver, dataResolver DataResolver) (*pb.Cursor, error) {
	c, _, err := s.datedRef.IndexInRange(after, before)
	if err != nil {
		return nil, err
	}

	closer := pb.CloseFunc(func() error {
		return c.Close()
	})

	browser := pb.BrowseFunc(func() (*pb.Object, error) {
		if !c.HasNext() {
			return nil, io.EOF
		}

		next, err := c.Next()
		if err != nil {
			return nil, err
		}

		entry := next.(*bome.ListEntry)
		id := entry.Value

		o := &pb.Object{Header: &pb.Header{}}

		o.Header, err = headerResolver.ResolveHeader(id)
		if err != nil {
			return nil, err
		}

		if dataResolver != nil {
			o.Data, err = dataResolver.ResolveData(o.Header.Id)
		} else {
			o.Data, err = s.objects.Get(entry.Value)
			if err != nil {
				return nil, err
			}
		}
		return o, nil
	})
	return pb.NewCursor(browser, closer), nil
}
