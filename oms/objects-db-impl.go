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
	"github.com/tidwall/gjson"
	"io/ioutil"
	"strings"
	"time"
)

func NewStore(db *sql.DB) (Objects, error) {
	m, err := bome.NewJSONMap(db, bome.MySQL, "objects")
	if err != nil {
		return nil, err
	}

	h, err := bome.NewJSONMap(db, bome.MySQL, "headers")
	if err != nil {
		return nil, err
	}

	l, err := bome.NewList(db, bome.MySQL, "dated_refs")
	if err != nil {
		return nil, err
	}

	s := &mysqlStore{
		objects:   m,
		datedRefs: l,
		headers:   h,
	}
	return s, nil
}

type mysqlStore struct {
	objects   *bome.JSONMap
	headers   *bome.JSONMap
	datedRefs *bome.List
	cEnv      *cel.Env
}

func (ms *mysqlStore) Save(ctx context.Context, object *Object) error {
	d := time.Now().Unix()
	object.SetCreatedAt(d)

	contentData, err := ioutil.ReadAll(object.Content())
	if err != nil {
		log.Error("Save: could not get object content", log.Err(err))
		return errors.BadInput
	}

	headersData, err := json.Marshal(object.Header())
	if err != nil {
		log.Error("Save: could not get object info", log.Err(err))
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

	ltx := ms.datedRefs.ContinueTransaction(tx.TX())
	err = ltx.Save(&bome.ListEntry{
		Index: d,
		Value: object.ID(),
	})
	if err != nil {
		log.Error("Save: failed to save object dated ref", log.Err(err))
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

func (ms *mysqlStore) Patch(ctx context.Context, objectID string, path string, data string) error {

	tx, err := ms.objects.BeginTransaction()
	if err != nil {
		log.Error("Patch: could not start objects DB transaction", log.Err(err))
		return errors.Internal
	}

	err = tx.EditAt(objectID, path, bome.StringExpr(data))
	if err != nil {
		log.Error("Update: object patch failed", log.Field("id", objectID), log.Err(err))
		return errors.Internal
	}

	size, err := tx.Size(objectID)
	if err != nil {
		log.Error("Patch: failed to get object size", log.Field("id", objectID), log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Patch: rollback failed", log.Err(err))
		}
		return err
	}

	info, err := ms.Info(ctx, objectID)
	if err != nil {
		return err
	}

	info.Size = int64(size)
	headersData, err := json.Marshal(info)
	if err != nil {
		log.Error("Save: could not get object info", log.Err(err))
		if err := tx.Rollback(); err != nil {
			log.Error("Patch: rollback failed", log.Err(err))
		}
		return errors.BadInput
	}

	htx := ms.headers.ContinueTransaction(tx.TX())
	err = htx.Save(&bome.MapEntry{
		Key:   objectID,
		Value: string(headersData),
	})
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

	log.Debug("Update: object updated", log.Field("id", objectID))
	return nil
}

func (ms *mysqlStore) Delete(ctx context.Context, objectID string) error {
	object, err := ms.Get(ctx, objectID)
	if err != nil {
		return err
	}

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

	ltx := ms.datedRefs.ContinueTransaction(tx.TX())
	err = ltx.Delete(object.CreatedAt())
	if err != nil && !bome.IsNotFound(err) {
		log.Error("Delete: failed to delete dated ref", log.Err(err))
		return errors.Internal
	}

	err = ltx.Commit()
	if err != nil {
		log.Error("Delete: operations commit failed", log.Err(err))
		return errors.Internal
	}

	log.Debug("Delete: object deleted", log.Field("id", object.ID()))
	return nil
}

func (ms *mysqlStore) List(ctx context.Context, before int64, count int, filter ObjectFilter) (*ObjectList, error) {
	cursor, err := ms.datedRefs.AllBefore(before)
	if err != nil {
		log.Error("List: failed to get objects",
			log.Field("created before", before),
			log.Field("count", count), log.Err(err))
		return nil, errors.Internal
	}

	var result ObjectList
	for cursor.HasNext() && len(result.Objects) < count {
		item, err := cursor.Next()
		if err != nil {
			return nil, err
		}

		id := item.(string)
		o, err := ms.Get(ctx, id)
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
		result.Objects = append(result.Objects, o)
	}
	return &result, nil
}

func (ms *mysqlStore) ListAt(ctx context.Context, partPath string, before int64, count int, filter ObjectFilter) (*ObjectList, error) {
	cursor, err := ms.datedRefs.AllBefore(before)
	if err != nil {
		log.Error("List: failed to get objects",
			log.Field("created before", before),
			log.Field("count", count), log.Err(err))
		return nil, errors.Internal
	}

	var result ObjectList
	for cursor.HasNext() && len(result.Objects) < count {
		item, err := cursor.Next()
		if err != nil {
			return nil, err
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
		result.Objects = append(result.Objects, o)
	}
	return &result, nil
}

func (ms *mysqlStore) Get(ctx context.Context, id string) (*Object, error) {
	value, err := ms.objects.Get(id)
	if err != nil {
		return nil, err
	}

	o, err := DecodeObject(value)
	if err != nil {
		log.Error("List: failed to decode item", log.Field("encoded", value), log.Err(err))
		return nil, errors.Internal
	}
	return o, nil
}

func (ms *mysqlStore) GetAt(ctx context.Context, id string, path string) (*Object, error) {
	value, err := ms.objects.Get(id)
	if err != nil {
		return nil, err
	}

	v := gjson.Get(value, strings.TrimPrefix("$.", path))
	value = v.String()

	o, err := DecodeObject(value)
	if err != nil {
		log.Error("List: failed to decode item", log.Field("encoded", value), log.Err(err))
		return nil, errors.Internal
	}

	o.SetContent(bytes.NewBuffer([]byte(value)), int64(len(value)))
	return o, nil
}

func (ms *mysqlStore) Info(ctx context.Context, id string) (*Info, error) {
	value, err := ms.objects.Get(id)
	if err != nil {
		return nil, err
	}

	var info *Info
	err = json.Unmarshal([]byte(value), info)
	if err != nil {
		log.Error("List: failed to decode object info", log.Field("encoded", value), log.Err(err))
		return nil, errors.Internal
	}
	return info, nil
}
