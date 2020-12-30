package oms

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/omecodes/bome"
	"github.com/omecodes/libome/logs"
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

func (s *sqlCollection) Save(ctx context.Context, item *CollectionItem) error {
	if bome.IsTransactionContext(ctx) {
		_, tx, err := s.objects.Transaction(ctx)
		if err != nil {
			return err
		}

		err = tx.Save(&bome.MapEntry{
			Key:   item.Id,
			Value: item.Data,
		})
		if err == nil {
			dtx := s.datedRef.ContinueTransaction(tx.TX())
			err = dtx.Save(&bome.ListEntry{
				Index: item.Date,
				Value: item.Id,
			})
		}
		return err
	}

	err := s.objects.Save(&bome.MapEntry{
		Key:   item.Id,
		Value: item.Data,
	})
	if err == nil {
		err = s.datedRef.Save(&bome.ListEntry{
			Index: item.Date,
			Value: item.Id,
		})
	}
	return err
}

func (s *sqlCollection) Select(ctx context.Context, before int64, count int, selector Selector) ([]*CollectionItem, error) {
	c, err := s.objects.List()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := c.Close(); err != nil {
			logs.Error("closing cursor caused error", logs.Err(err))
		}
	}()

	var items []*CollectionItem
	for c.HasNext() && len(items) < count {
		o, err := c.Next()
		if err != nil {
			return nil, err
		}

		itemData := JSON{}
		entry := o.(*bome.MapEntry)
		err = json.Unmarshal([]byte(entry.Value), &itemData)
		if err != nil {
			return nil, err
		}

		if selector != nil {
			selected, err := selector.Select(itemData)
			if err != nil {
				return nil, err
			}

			if !selected {
				continue
			}
		}

		item := &CollectionItem{
			Id:   entry.Key,
			Data: entry.Value,
		}
		items = append(items, item)
	}

	return items, nil
}
