package oms

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/pb"
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

func (ms *sqlStore) GetCollectionForWrite(ctx context.Context, name string) (Collection, error) {
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

func (ms *sqlStore) GetCollectionForRead(ctx context.Context, name string) (Collection, error) {
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

func (ms *sqlStore) Save(ctx context.Context, object *Object, indexes ...*pb.Index) error {
	if object.CreatedAt() == 0 {
		d := time.Now().UnixNano()
		object.SetCreatedAt(d)
	}

	contentData, err := ioutil.ReadAll(object.GetContent())
	if err != nil {
		log.Error("Save: could not get object content", log.Err(err))
		return errors.BadInput
	}

	txCtx, tx, err := ms.objects.Transaction(ctx)
	if err != nil {
		log.Error("Save: could not start objects DB transaction", log.Err(err))
		return errors.Internal
	}

	// Save object
	err = tx.Save(&bome.MapEntry{
		Key:   object.ID(),
		Value: string(contentData),
	})
	if err != nil {
		log.Error("Save: failed to save object data", log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Save: rollback failed", log.Err(err))
		}
		return errors.Internal
	}

	// Then retrieve saved size
	size, err := tx.Size(object.ID())
	if err != nil {
		log.Error("Patch: failed to get object size", log.Field("id", object.ID()), log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Patch: rollback failed", log.Err(err))
		}
		return err
	}

	// Update object header size
	object.Header().Size = size
	headersData, err := json.Marshal(object.Header())
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
		Key:   object.ID(),
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
		Index: object.CreatedAt(),
		Value: object.ID(),
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
			col, err := ms.GetCollectionForWrite(txCtx, ind.Collection)
			if err != nil {
				log.Error("Save: failed to get collection", log.Err(err))
				if err2 := tx.Rollback(); err2 != nil {
					log.Error("Save: rollback failed", log.Err(err2))
				}
				return err
			}

			js := map[string]interface{}{}
			for _, item := range ind.Fields {
				result := gjson.Get(string(contentData), strings.TrimPrefix(item.Path, "$."))
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

			err = col.Save(txCtx, &CollectionItem{
				Id:   object.ID(),
				Date: object.header.CreatedAt,
				Data: string(jsonEncoded),
			})
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
			Key:   object.ID(),
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
	log.Debug("Save: object saved", log.Field("id", object.ID()))
	return nil
}

func (ms *sqlStore) Patch(ctx context.Context, patch *Patch) error {
	content, err := ioutil.ReadAll(patch.GetContent())
	if err != nil {
		log.Error("Patch: could not get patch content", log.Err(err))
		return errors.BadInput
	}

	_, tx, err := ms.objects.Transaction(ctx)
	if err != nil {
		log.Error("Patch: could not start objects DB transaction", log.Err(err))
		return errors.Internal
	}

	err = tx.EditAt(patch.GetObjectID(), patch.Path(), bome.StringExpr(string(content)))
	if err != nil {
		log.Error("Update: object patch failed", log.Field("id", patch.GetObjectID()), log.Err(err))
		return errors.Internal
	}

	size, err := tx.Size(patch.GetObjectID())
	if err != nil {
		log.Error("Patch: failed to get object size", log.Field("id", patch.GetObjectID()), log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Patch: rollback failed", log.Err(err))
		}
		return err
	}

	htx := ms.headers.ContinueTransaction(tx.TX())
	err = htx.EditAt(patch.GetObjectID(), "$.size", bome.IntExpr(size))
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

	log.Debug("Patch: object updated", log.Field("id", patch.GetObjectID()))
	return nil
}

func (ms *sqlStore) Delete(ctx context.Context, objectID string) error {
	tx, err := ms.objects.BeginTransaction()
	if err != nil {
		log.Error("Delete: could not start objects DB transaction", log.Err(err))
		return errors.Internal
	}

	err = tx.Delete(objectID)
	if err != nil {
		log.Error("Delete: object deletion failed", log.Err(err))
		return errors.Internal
	}

	htx := ms.headers.ContinueTransaction(tx.TX())
	err = htx.Delete(objectID)
	if err != nil {
		log.Error("Delete: failed to delete header", log.Err(err))
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

func (ms *sqlStore) List(ctx context.Context, before int64, count int, filter ObjectFilter) (*ObjectList, error) {
	cursor, err := ms.headers.List()
	if err != nil {
		log.Error("List: failed to get headers",
			log.Field("created before", before),
			log.Field("count", count), log.Err(err))
		return nil, errors.Internal
	}

	defer func() {
		if err := cursor.Close(); err != nil {
			log.Error("List: cursor close failed", log.Err(err))
		}
	}()

	var result ObjectList
	for cursor.HasNext() && len(result.Objects) < count {
		item, err := cursor.Next()
		if err != nil {
			return nil, err
		}

		entry := item.(*bome.MapEntry)
		var header pb.Header
		err = json.Unmarshal([]byte(entry.Value), &header)
		if err != nil {
			log.Error("List: failed to decode object", log.Err(err))
			return nil, errors.Internal
		}

		value, err := ms.objects.Get(header.Id)
		if err != nil {
			return nil, err
		}

		o := NewObject()
		o.SetHeader(&header)
		if filter != nil {
			o.SetContent(bytes.NewBufferString(value))
			allowed, err := filter.Filter(o)
			if err != nil {
				if err == errors.Unauthorized || err == errors.Forbidden {
					continue
				}
				return nil, err
			}

			if !allowed {
				continue
			}
		}
		result.Count++
		o.SetContent(bytes.NewBufferString(value))
		result.Objects = append(result.Objects, o)
	}

	result.Before = before
	result.Count = len(result.Objects)
	return &result, nil
}

func (ms *sqlStore) ListAt(ctx context.Context, partPath string, before int64, count int, filter ObjectFilter) (*ObjectList, error) {
	cursor, err := ms.headers.ExtractAll("$.id", bome.JsonAtLe("$.created_at", bome.IntExpr(before)), bome.StringScanner)
	if err != nil {
		log.Error("ListAt: failed to get objects",
			log.Field("created before", before),
			log.Field("count", count), log.Err(err))
		return nil, errors.Internal
	}

	defer func() {
		if err := cursor.Close(); err != nil {
			log.Error("ListAt: cursor close failed", log.Err(err))
		}
	}()

	var result ObjectList
	for cursor.HasNext() && len(result.Objects) < count {
		item, err := cursor.Next()
		if err != nil {
			log.Error("ListAt: failed to get object from curosr", log.Err(err))
			return nil, errors.Internal
		}

		id := item.(string)
		o, err := ms.GetAt(ctx, id, partPath)
		if err != nil {
			return nil, err
		}

		if filter != nil {
			allowed, err := filter.Filter(o)
			if err != nil {
				return nil, err
			}

			if !allowed {
				continue
			}
		}
		result.Count++
		result.Objects = append(result.Objects, o)
	}
	return &result, nil
}

func (ms *sqlStore) Get(ctx context.Context, objectID string) (*Object, error) {
	hv, err := ms.headers.Get(objectID)
	if err != nil {
		log.Error("Get: could not get object header", log.Field("id", objectID), log.Err(err))
		if bome.IsNotFound(err) {
			return nil, err
		}
		return nil, errors.Internal
	}

	var info pb.Header
	err = json.Unmarshal([]byte(hv), &info)
	if err != nil {
		log.Error("Get: could not decode object header", log.Err(err))
		return nil, errors.Internal
	}

	value, err := ms.objects.Get(objectID)
	if err != nil {
		return nil, err
	}

	o, err := DecodeObject(value)
	if err != nil {
		log.Error("List: failed to decode item", log.Field("encoded", value), log.Err(err))
		return nil, errors.Internal
	}

	o.header = &info
	log.Debug("Get: loaded object", log.Field("id", objectID))
	return o, nil
}

func (ms *sqlStore) GetAt(ctx context.Context, objectID string, path string) (*Object, error) {
	hv, err := ms.headers.Get(objectID)
	if err != nil {
		log.Error("GetAt: could not get object header", log.Field("id", objectID), log.Err(err))
		if bome.IsNotFound(err) {
			return nil, err
		}
		return nil, errors.Internal
	}

	var info pb.Header
	err = json.Unmarshal([]byte(hv), &info)
	if err != nil {
		log.Error("Get: could not decode object header", log.Err(err))
		return nil, errors.Internal
	}

	value, err := ms.objects.ExtractAt(objectID, path)
	if err != nil {
		return nil, err
	}

	o := new(Object)
	o.SetHeader(&info)
	o.SetContent(bytes.NewBuffer([]byte(value)))
	log.Debug("GetAt: content loaded", log.Field("id", objectID), log.Field("at", path))
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
