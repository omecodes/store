package objects

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/se"
	"github.com/omecodes/store/utime"
	"github.com/tidwall/gjson"
	"io"
	"strings"
)

func NewSQLCollection(collection *pb.Collection, db *sql.DB, dialect string, tableName string) (*sqlCollection, error) {
	objects, err := bome.NewJSONMap(db, dialect, tableName)
	if err != nil {
		return nil, err
	}

	headers, err := bome.NewJSONMap(db, dialect, tableName+"_headers")
	if err != nil {
		return nil, err
	}

	datedRefs, err := bome.NewList(db, dialect, tableName+"_dated_refs")
	if err != nil {
		return nil, err
	}

	indexes, err := bome.NewJSONMap(db, dialect, tableName+"_indexes")
	if err != nil {
		return nil, err
	}

	fk := &bome.ForeignKey{
		Name: "fk_objects_header_id",
		Table: &bome.Keys{
			Table:  headers.Table(),
			Fields: headers.Keys(),
		},
		References: &bome.Keys{
			Table:  objects.Table(),
			Fields: objects.Keys(),
		},
		OnDeleteCascade: true,
	}
	err = objects.AddForeignKey(fk)
	if err != nil {
		return nil, err
	}

	fk = &bome.ForeignKey{
		Name: "fk_objects_indexes_id",
		Table: &bome.Keys{
			Table:  indexes.Table(),
			Fields: indexes.Keys(),
		},
		References: &bome.Keys{
			Table:  objects.Table(),
			Fields: objects.Keys(),
		},
		OnDeleteCascade: true,
	}
	err = objects.AddForeignKey(fk)
	if err != nil {
		return nil, err
	}

	indexStore, err := se.NewSQLIndexStore(db, dialect, tableName)
	if err != nil {
		return nil, err
	}

	s := &sqlCollection{
		db:        db,
		dialect:   dialect,
		objects:   objects,
		headers:   headers,
		datedRefs: datedRefs,
		info:      collection,
		engine:    se.NewEngine(indexStore),
	}
	return s, nil
}

type sqlCollection struct {
	info    *pb.Collection
	dialect string
	db      *sql.DB
	engine  *se.Engine

	indexes []*pb.Index

	objects   *bome.JSONMap
	datedRefs *bome.List
	headers   *bome.JSONMap
}

func (s *sqlCollection) Objects() *bome.JSONMap {
	return s.objects
}

func (s *sqlCollection) DatedRefs() *bome.List {
	return s.datedRefs
}

func (s *sqlCollection) Save(ctx context.Context, object *pb.Object, indexes ...*pb.Index) error {
	if object.Header.CreatedAt == 0 {
		object.Header.CreatedAt = utime.Now()
	}

	_, tx, err := s.objects.Transaction(ctx)
	if err != nil {
		log.Error("Save: could not start objects DB transaction", log.Err(err))
		return errors.Internal
	}

	// Save object
	err = tx.Save(&bome.MapEntry{
		Key:   object.Header.Id,
		Value: object.Data,
	})
	if err != nil {
		log.Error("Save: failed to save object data", log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Save: rollback failed", log.Err(err))
		}
		return errors.Internal
	}

	// Then retrieve saved size
	size, err := tx.Size(object.Header.Id)
	if err != nil {
		log.Error("Patch: failed to get object size", log.Field("id", object.Header.Id), log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Patch: rollback failed", log.Err(err))
		}
		return err
	}

	// Update object header size
	object.Header.Size = size
	headersData, err := json.Marshal(object.Header)
	if err != nil {
		log.Error("Save: could not get object header", log.Err(err))
		if err2 := tx.Rollback(); err2 != nil {
			log.Error("Save: rollback failed", log.Err(err2))
		}
		return errors.BadInput
	}

	// Save object header
	htx := s.headers.ContinueTransaction(tx.TX())
	err = htx.Save(&bome.MapEntry{
		Key:   object.Header.Id,
		Value: string(headersData),
	})
	if err != nil {
		log.Error("Save: failed to save object headers", log.Err(err))
		if err2 := tx.Rollback(); err2 != nil {
			log.Error("Save: rollback failed", log.Err(err2))
		}
		return err
	}

	dTx := s.datedRefs.ContinueTransaction(tx.TX())
	err = dTx.Save(&bome.ListEntry{
		Index: object.Header.CreatedAt,
		Value: object.Header.Id,
	})
	if err != nil {
		log.Error("Save: failed to save dated reference", log.Err(err))
		if err2 := tx.Rollback(); err2 != nil {
			log.Error("Save: rollback failed", log.Err(err2))
		}
		return err
	}

	allIndexes := append(s.info.DefaultIndexes, indexes...)
	for _, index := range allIndexes {
		result := gjson.Get(object.Data, strings.TrimPrefix(index.JsonPath, "$."))
		if !result.Exists() {
			log.Error("Save: index references path that does not exists", log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return errors.BadInput
		}

		if index.FieldType == 0 {
			mp := &pb.TextMapping{
				Text:      result.Str,
				FieldName: index.MappingName,
				ObjectId:  object.Header.Id,
			}
			err = s.engine.CreateTextMapping(mp)
			if err != nil {
				log.Error("Save: index references path that does not exists", log.Err(err))
				if err2 := tx.Rollback(); err2 != nil {
					log.Error("Save: rollback failed", log.Err(err2))
				}
				return errors.BadInput
			}
		} else if index.FieldType == 1 {
			mp := &pb.NumberMapping{
				Number:    result.Int(),
				FieldName: index.MappingName,
				ObjectId:  object.Header.Id,
			}

			err = s.engine.CreateNumberMapping(mp)
			if err != nil {
				log.Error("Save: index references path that does not exists", log.Err(err))
				if err2 := tx.Rollback(); err2 != nil {
					log.Error("Save: rollback failed", log.Err(err2))
				}
				return errors.BadInput
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error("Save: operations commit failed", log.Err(err))
		return errors.Internal
	}
	log.Debug("Save: object saved", log.Field("id", object.Header.Id))
	return nil
}

func (s *sqlCollection) Patch(ctx context.Context, patch *pb.Patch) error {

	value := sqlJSONSetValue(patch.Data)

	_, tx, err := s.objects.Transaction(ctx)
	if err != nil {
		log.Error("Patch: could not start objects DB transaction", log.Err(err))
		return errors.Internal
	}

	err = tx.EditAt(patch.ObjectId, patch.At, bome.RawExpr(value))
	if err != nil {
		log.Error("Update: object patch failed", log.Field("id", patch.ObjectId), log.Err(err))
		return errors.Internal
	}

	size, err := tx.Size(patch.ObjectId)
	if err != nil {
		log.Error("Patch: failed to get object size", log.Field("id", patch.ObjectId), log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Patch: rollback failed", log.Err(err))
		}
		return err
	}

	htx := s.headers.ContinueTransaction(tx.TX())
	err = htx.EditAt(patch.ObjectId, "$.size", bome.IntExpr(size))
	if err != nil {
		log.Error("Patch: failed to save object headers", log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Patch: rollback failed", log.Err(err))
		}
		return errors.Internal
	}

	err = tx.Commit()
	if err != nil {
		log.Error("Patch: operations commit failed", log.Err(err))
		return errors.Internal
	}

	log.Debug("Patch: object updated", log.Field("id", patch.ObjectId))
	return nil
}

func (s *sqlCollection) Delete(ctx context.Context, objectID string) error {
	info, err := s.Info(ctx, objectID)
	if err != nil {
		return err
	}

	tx, err := s.objects.BeginTransaction()
	if err != nil {
		log.Error("Delete: could not start objects DB transaction", log.Err(err))
		if err2 := tx.Rollback(); err2 != nil {
			log.Error("Delete: rollback failed", log.Err(err2))
		}
		return errors.Internal
	}

	err = tx.Delete(objectID)
	if err != nil {
		log.Error("Delete: object deletion failed", log.Err(err))
		if err2 := tx.Rollback(); err2 != nil {
			log.Error("Delete: rollback failed", log.Err(err2))
		}
		return errors.Internal
	}

	htx := s.headers.ContinueTransaction(tx.TX())
	err = htx.Delete(objectID)
	if err != nil {
		log.Error("Delete: failed to delete header", log.Err(err))
		if err2 := tx.Rollback(); err2 != nil {
			log.Error("Delete: rollback failed", log.Err(err2))
		}
		return errors.Internal
	}

	dtx := s.datedRefs.ContinueTransaction(tx.TX())
	err = dtx.Delete(info.CreatedAt)
	if err != nil {
		log.Error("Delete: failed to delete dated references", log.Err(err))
		if err2 := tx.Rollback(); err2 != nil {
			log.Error("Delete: rollback failed", log.Err(err2))
		}
		return errors.Internal
	}

	err = tx.Commit()
	if err != nil {
		log.Error("Delete: operations commit failed", log.Err(err))
		return errors.Internal
	}

	log.Debug("Delete: object deleted", log.Field("id", objectID))
	return nil
}

func (s *sqlCollection) listInRange(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	cursor, _, err := s.datedRefs.IndexInRange(opts.DateOptions.After, opts.DateOptions.Before)
	if err != nil {
		return nil, err
	}

	closer := pb.CloseFunc(func() error {
		return cursor.Close()
	})
	browser := pb.BrowseFunc(func() (*pb.Object, error) {
		if !cursor.HasNext() {
			return nil, io.EOF
		}
		next, err := cursor.Next()
		if err != nil {
			return nil, err
		}
		entry := next.(*bome.ListEntry)
		id := entry.Value
		return s.Get(context.Background(), id, pb.GetOptions{
			At: opts.At,
		})
	})

	return pb.NewCursor(browser, closer), nil
}

func (s *sqlCollection) Get(ctx context.Context, objectID string, opts pb.GetOptions) (*pb.Object, error) {
	hv, err := s.headers.Get(objectID)
	if err != nil {
		log.Error("Get: could not get object header", log.Field("id", objectID), log.Err(err))
		if bome.IsNotFound(err) {
			return nil, err
		}
		return nil, errors.Internal
	}

	o := &pb.Object{
		Header: &pb.Header{},
	}

	err = json.Unmarshal([]byte(hv), o.Header)
	if err != nil {
		log.Error("Get: could not decode object header", log.Err(err))
		return nil, errors.Internal
	}

	if opts.At != "" {
		o.Data, err = s.objects.ExtractAt(objectID, opts.At)
		if err != nil {
			return nil, err
		}
	} else {
		o.Data, err = s.objects.Get(objectID)
		if err != nil {
			return nil, err
		}
	}

	log.Debug("Get: loaded object", log.Field("id", objectID))
	return o, nil
}

func (s *sqlCollection) Info(ctx context.Context, id string) (*pb.Header, error) {
	value, err := s.headers.Get(id)
	if err != nil {
		return nil, err
	}

	var info pb.Header
	err = json.Unmarshal([]byte(value), &info)
	if err != nil {
		log.Error("List: failed to decode object header", log.Field("encoded", value), log.Err(err))
		return nil, errors.Internal
	}
	return &info, nil
}

func (s *sqlCollection) List(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	cursor, err := s.headers.List()
	if err != nil {
		return nil, err
	}

	closer := pb.CloseFunc(func() error {
		return cursor.Close()
	})
	browser := pb.BrowseFunc(func() (*pb.Object, error) {
		if !cursor.HasNext() {
			return nil, io.EOF
		}

		next, err := cursor.Next()
		if err != nil {
			return nil, err
		}

		o := &pb.Object{
			Header: &pb.Header{},
		}

		entry := next.(*bome.MapEntry)
		err = json.Unmarshal([]byte(entry.Value), o.Header)
		if err != nil {
			return nil, err
		}
		o.Data, err = s.objects.Get(entry.Key)
		return o, err
	})
	return pb.NewCursor(browser, closer), nil
}

func (s *sqlCollection) Search(ctx context.Context, expression *pb.BooleanExp) (*pb.Cursor, error) {
	return nil, errors.Unavailable
}

func (s *sqlCollection) Clear() error {
	tx, err := s.objects.BeginTransaction()
	if err != nil {
		log.Error("Clear: could not start transaction in objects DB", log.Err(err))
		return errors.Internal
	}

	err = tx.Clear()
	if err != nil {
		log.Error("Clear: could not clear objects", log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Clear: operations rollback failed", log.Err(err))
		}
		return errors.Internal
	}

	htx := s.headers.ContinueTransaction(tx.TX())
	err = htx.Clear()
	if err != nil {
		log.Error("Clear: could not clear objects headers", log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Clear: operations rollback failed", log.Err(err))
		}
		return errors.Internal
	}

	if err := tx.Commit(); err != nil {
		log.Error("Clear: operations commit failed", log.Err(err))
	}

	log.Debug("Clear: objects store has been cleared")
	return nil
}

func sqlJSONSetValue(value string) string {
	var o interface{}
	err := json.Unmarshal([]byte(value), &o)
	if err != nil {
		val := strings.Replace(value, "'", `\'`, -1)
		return "'" + val + "'"
	}

	val := strings.Replace(value, "'", `\'`, -1)
	return "CAST('" + val + "' AS JSON)"
}
