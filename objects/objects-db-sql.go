package objects

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/pb"
	"github.com/omecodes/store/utime"
	"github.com/tidwall/gjson"
)

func NewSQLObjects(db *sql.DB, dialect string, tableName string) (Objects, error) {
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

	col, err := bome.NewJSONMap(db, dialect, tableName+"_collections")
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

	s := &sqlStore{
		db:          db,
		dialect:     dialect,
		objects:     objects,
		headers:     headers,
		datedRefs:   datedRefs,
		indexes:     indexes,
		collections: col,
	}
	return s, nil
}

type sqlStore struct {
	db          *sql.DB
	dialect     string
	objects     *bome.JSONMap
	datedRefs   *bome.List
	headers     *bome.JSONMap
	collections *bome.JSONMap
	indexes     *bome.JSONMap
}

func (ms *sqlStore) GetOrCreateCollection(ctx context.Context, name string) (Collection, error) {
	tableName := strcase.ToSnake(name)
	col, err := NewSQLCollection(ms.db, ms.dialect, tableName)
	if err != nil {
		return nil, err
	}

	contains, err := ms.collections.Contains(name)
	if err != nil {
		return nil, err
	}

	if !contains {
		_, tx, err := ms.collections.Transaction(ctx)
		if err != nil {
			return nil, err
		}

		err = tx.Save(&bome.MapEntry{
			Key:   name,
			Value: "{\"count\": 0}",
		})
		if err != nil {
			return nil, err
		}

		datedRefs := col.DatedRefs()
		client := tx.TX()

		fk := &bome.ForeignKey{
			Name: "fk_col_objects_dated_refs_keys",
			Table: &bome.Keys{
				Table:  datedRefs.Table(),
				Fields: datedRefs.Keys(),
			},
			References: &bome.Keys{
				Table:  ms.datedRefs.Table(),
				Fields: ms.datedRefs.Keys(),
			},
			OnDeleteCascade: true,
		}
		err = client.SQLExec(fk.AlterTableAddQuery())
		if err != nil {
			return nil, err
		}

		objects := col.Objects()
		fk = &bome.ForeignKey{
			Name: "fk_col_objects_keys",
			Table: &bome.Keys{
				Table:  objects.Table(),
				Fields: objects.Keys(),
			},
			References: &bome.Keys{
				Table:  ms.objects.Table(),
				Fields: ms.objects.Keys(),
			},
			OnDeleteCascade: true,
		}
		err = client.SQLExec(fk.AlterTableAddQuery())
		if err != nil {
			return nil, err
		}
	}

	return col, nil
}

func (ms *sqlStore) GetCollection(ctx context.Context, name string) (Collection, error) {
	contains, err := ms.collections.Contains(name)
	if err != nil {
		return nil, err
	}

	if !contains {
		return nil, errors.NotFound
	}

	tableName := strcase.ToSnake(name)
	return NewSQLCollection(ms.db, ms.dialect, tableName)
}

func (ms *sqlStore) Save(ctx context.Context, object *pb.Object, indexes ...*pb.Index) error {
	if object.Header.CreatedAt == 0 {
		object.Header.CreatedAt = utime.Now()
	}

	txCtx, tx, err := ms.objects.Transaction(ctx)
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
	htx := ms.headers.ContinueTransaction(tx.TX())
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

	dTx := ms.datedRefs.ContinueTransaction(tx.TX())
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

	// Creating index
	if len(indexes) > 0 {
		for _, ind := range indexes {
			col, err := ms.GetOrCreateCollection(txCtx, ind.Collection)
			if err != nil {
				log.Error("Save: failed to get collection", log.Err(err))
				if err2 := tx.Rollback(); err2 != nil {
					log.Error("Save: rollback failed", log.Err(err2))
				}
				return err
			}

			js := map[string]interface{}{}
			for _, item := range ind.Fields {
				result := gjson.Get(object.Data, strings.TrimPrefix(item.Path, "$."))
				if !result.Exists() {
					log.Error("Save: index references path that does not exists", log.Err(err))
					if err2 := tx.Rollback(); err2 != nil {
						log.Error("Save: rollback failed", log.Err(err2))
					}
					return errors.BadInput
				}
				js[item.Name] = result.Value()
			}

			jsonEncoded, err := json.Marshal(js)
			if err != nil {
				log.Error("Save: failed to create index", log.Err(err))
				if err2 := tx.Rollback(); err2 != nil {
					log.Error("Save: rollback failed", log.Err(err2))
				}
				return err
			}

			err = col.Save(txCtx, object.Header.CreatedAt, object.Header.Id, string(jsonEncoded))
			if err != nil {
				log.Error("Save: failed to save object headers", log.Err(err))
				if err2 := tx.Rollback(); err2 != nil {
					log.Error("Save: rollback failed", log.Err(err2))
				}
				return err
			}
		}

		encodedIndexes, err := json.Marshal(indexes)
		if err != nil {
			log.Error("Save: failed to encode object indexes", log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return err
		}

		indexesTx := ms.indexes.ContinueTransaction(tx.TX())
		entry := &bome.MapEntry{
			Key:   object.Header.Id,
			Value: string(encodedIndexes),
		}
		err = indexesTx.Save(entry)
		if err != nil {
			log.Error("Save: failed to save object indexes", log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return err
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

func (ms *sqlStore) Patch(ctx context.Context, patch *pb.Patch) error {

	value := sqlJSONSetValue(patch.Data)

	_, tx, err := ms.objects.Transaction(ctx)
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

	htx := ms.headers.ContinueTransaction(tx.TX())
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

func (ms *sqlStore) Delete(ctx context.Context, objectID string) error {
	info, err := ms.Info(ctx, objectID)
	if err != nil {
		return err
	}

	tx, err := ms.objects.BeginTransaction()
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

	htx := ms.headers.ContinueTransaction(tx.TX())
	err = htx.Delete(objectID)
	if err != nil {
		log.Error("Delete: failed to delete header", log.Err(err))
		if err2 := tx.Rollback(); err2 != nil {
			log.Error("Delete: rollback failed", log.Err(err2))
		}
		return errors.Internal
	}

	dtx := ms.datedRefs.ContinueTransaction(tx.TX())
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

func (ms *sqlStore) List(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	if opts.CollectionOptions.Name != "" {
		return ms.collectionList(ctx, opts)
	}

	if opts.DateOptions.After > 0 && opts.DateOptions.Before > 0 {
		return ms.listInRange(ctx, opts)
	}

	cursor, err := ms.headers.List()
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
		o.Data, err = ms.objects.Get(entry.Key)
		return o, err
	})
	return pb.NewCursor(browser, closer), nil
}

func (ms *sqlStore) listInRange(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	cursor, _, err := ms.datedRefs.AllInRange(opts.DateOptions.After, opts.DateOptions.Before, 100)
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
		return ms.Get(context.Background(), id, pb.GetOptions{
			At: opts.At,
		})
	})

	return pb.NewCursor(browser, closer), nil
}

func (ms *sqlStore) collectionList(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	col, err := ms.GetCollection(ctx, opts.CollectionOptions.Name)
	if err != nil {
		return nil, err
	}

	var dataResolver DataResolver
	if opts.CollectionOptions.FullObject {
		dataResolver = ms
	} else if opts.At != "" {
		dataResolver = ms.ResolveAtFunc(opts.At)
	}

	if opts.DateOptions.Before > 0 && opts.DateOptions.After > 0 {
		return col.RangeSelect(ctx, opts.DateOptions.After, opts.DateOptions.Before, ResolverHeaderFunc(ms.ResolveHeader), dataResolver)
	} else {
		return col.Select(ctx, ResolverHeaderFunc(ms.ResolveHeader), dataResolver)
	}
}

func (ms *sqlStore) Get(ctx context.Context, objectID string, opts pb.GetOptions) (*pb.Object, error) {
	hv, err := ms.headers.Get(objectID)
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
		o.Data, err = ms.objects.ExtractAt(objectID, opts.At)
		if err != nil {
			return nil, err
		}
	} else {
		o.Data, err = ms.objects.Get(objectID)
		if err != nil {
			return nil, err
		}
	}

	log.Debug("Get: loaded object", log.Field("id", objectID))
	return o, nil
}

func (ms *sqlStore) Info(ctx context.Context, id string) (*pb.Header, error) {
	value, err := ms.headers.Get(id)
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

func (ms *sqlStore) Clear() error {
	tx, err := ms.objects.BeginTransaction()
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

	htx := ms.headers.ContinueTransaction(tx.TX())
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

func (ms *sqlStore) ResolveData(id string) (string, error) {
	return ms.objects.Get(id)
}

func (ms *sqlStore) ResolveAtFunc(at string) DataResolver {
	return ResolveDataFunc(func(objectID string) (string, error) {
		return ms.objects.ExtractAt(objectID, at)
	})
}

func (ms *sqlStore) ResolveHeader(id string) (*pb.Header, error) {
	return ms.Info(context.Background(), id)
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
