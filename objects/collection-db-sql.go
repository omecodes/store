package objects

import (
	"context"
	"database/sql"
	"github.com/omecodes/bome"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/pb"
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

func (s *sqlCollection) Select(ctx context.Context, count int, filter ObjectFilter, resolver ObjectResolver) ([]*pb.Object, uint32, error) {
	total, err := s.objects.Count()
	if err != nil {
		return nil, 0, err
	}

	c, err := s.objects.List()
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		if err := c.Close(); err != nil {
			logs.Error("closing cursor caused error", logs.Err(err))
		}
	}()

	var items []*pb.Object
	for c.HasNext() && len(items) < count {
		o, err := c.Next()
		if err != nil {
			return nil, 0, err
		}

		entry := o.(*bome.MapEntry)

		object := &pb.Object{
			Header: &pb.Header{
				Id: entry.Key,
			},
			Data: entry.Value,
		}
		if filter != nil {
			selected, err := filter.Filter(object)
			if err != nil {
				return nil, 0, err
			}
			if !selected {
				continue
			}
		}

		if resolver != nil {
			object, err = resolver.ResolveObject(entry.Key)
			if err != nil {
				return nil, 0, err
			}
		}

		items = append(items, object)
	}

	return items, uint32(total), nil
}

func (s *sqlCollection) RangeSelect(ctx context.Context, after int64, before int64, count int, filter ObjectFilter, resolver ObjectResolver) ([]*pb.Object, uint32, error) {
	c, total, err := s.datedRef.IndexRange(after, before, count)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		if err := c.Close(); err != nil {
			logs.Error("closing cursor caused error", logs.Err(err))
		}
	}()

	var items []*pb.Object
	for c.HasNext() && len(items) < count {
		o, err := c.Next()
		if err != nil {
			return nil, 0, err
		}

		listEntry := o.(*bome.ListEntry)
		value, err := s.objects.Get(listEntry.Value)
		if err != nil {
			return nil, 0, err
		}

		object := &pb.Object{
			Header: &pb.Header{
				Id:        listEntry.Value,
				CreatedAt: listEntry.Index,
			},
			Data: value,
		}
		if filter != nil {
			selected, err := filter.Filter(object)
			if err != nil {
				return nil, 0, err
			}
			if !selected {
				continue
			}
		}

		if resolver != nil {
			object, err = resolver.ResolveObject(object.Header.Id)
			if err != nil {
				return nil, 0, err
			}
		}
		items = append(items, object)
	}

	return items, uint32(total), nil
}
