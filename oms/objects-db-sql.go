package oms

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/google/cel-go/cel"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/pb"
	"io/ioutil"
	"time"
)

func NewSQLObjects(db *sql.DB, dialect string) (Objects, error) {
	m, err := bome.NewJSONMap(db, dialect, "objects")
	if err != nil {
		return nil, err
	}

	h, err := bome.NewJSONMap(db, dialect, "headers")
	if err != nil {
		return nil, err
	}

	s := &mysqlStore{
		objects: m,
		headers: h,
	}
	return s, nil
}

type mysqlStore struct {
	objects *bome.JSONMap
	headers *bome.JSONMap
	cEnv    *cel.Env
}

func (ms *mysqlStore) Save(ctx context.Context, object *Object) error {
	if object.CreatedAt() == 0 {
		d := time.Now().UnixNano() / 1e6
		object.SetCreatedAt(d)
	}

	contentData, err := ioutil.ReadAll(object.GetContent())
	if err != nil {
		log.Error("Save: could not get object content", log.Err(err))
		return errors.BadInput
	}

	tx, err := ms.objects.BeginTransaction()
	if err != nil {
		log.Error("Save: could not start objects DB transaction", log.Err(err))
		return errors.Internal
	}

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

	size, err := tx.Size(object.ID())
	if err != nil {
		log.Error("Patch: failed to get object size", log.Field("id", object.ID()), log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Patch: rollback failed", log.Err(err))
		}
		return err
	}

	object.Header().Size = size
	headersData, err := json.Marshal(object.Header())
	if err != nil {
		log.Error("Save: could not get object header", log.Err(err))
		return errors.BadInput
	}

	htx := ms.headers.ContinueTransaction(tx.TX())
	err = htx.Save(&bome.MapEntry{
		Key:   object.ID(),
		Value: string(headersData),
	})
	if err != nil {
		log.Error("Save: failed to save object headers", log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Save: rollback failed", log.Err(err))
		}
		return errors.Internal
	}

	err = tx.Commit()
	if err != nil {
		log.Error("Save: operations commit failed", log.Err(err))
		return errors.Internal
	}

	log.Debug("Save: object saved", log.Field("id", object.ID()))
	return nil
}

func (ms *mysqlStore) Patch(ctx context.Context, patch *Patch) error {
	tx, err := ms.objects.BeginTransaction()
	if err != nil {
		log.Error("Patch: could not start objects DB transaction", log.Err(err))
		return errors.Internal
	}

	content, err := ioutil.ReadAll(patch.GetContent())
	if err != nil {
		log.Error("Patch: could not get patch content", log.Err(err))
		return errors.BadInput
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

func (ms *mysqlStore) Delete(ctx context.Context, objectID string) error {
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

func (ms *mysqlStore) List(ctx context.Context, before int64, count int, filter ObjectFilter) (*ObjectList, error) {
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

func (ms *mysqlStore) ListAt(ctx context.Context, partPath string, before int64, count int, filter ObjectFilter) (*ObjectList, error) {
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

func (ms *mysqlStore) Get(ctx context.Context, objectID string) (*Object, error) {
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

func (ms *mysqlStore) GetAt(ctx context.Context, objectID string, path string) (*Object, error) {
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

func (ms *mysqlStore) Info(ctx context.Context, id string) (*pb.Header, error) {
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

func (ms *mysqlStore) Clear() error {
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
