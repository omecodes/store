package objects

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/common/utime"
	se "github.com/omecodes/store/search-engine"
	"github.com/tidwall/gjson"
	"io"
	"strings"
)

const objectScanner = "object"

func NewSQLCollection(collection *Collection, db *sql.DB, dialect string, tableName string) (*sqlCollection, error) {
	objectsTableName := "store_" + tableName + "_objects"
	objects, err := bome.Build().
		SetDialect(dialect).
		SetConn(db).
		SetTableName(objectsTableName).
		JSONMap()
	if err != nil {
		return nil, err
	}

	headersTableName := "store_" + tableName + "_headers"
	headers, err := bome.Build().
		SetDialect(dialect).
		SetConn(db).
		SetTableName(headersTableName).
		AddForeignKeys(&bome.ForeignKey{
			Name: "fk_objects_header_id",
			Table: &bome.Keys{
				Table:  headersTableName,
				Fields: []string{"name"},
			},
			References: &bome.Keys{
				Table:  objectsTableName,
				Fields: []string{"name"},
			},
			OnDeleteCascade: true,
		}).
		JSONMap()
	if err != nil {
		return nil, err
	}

	indexTablePrefix := "store_" + tableName + "_index"
	indexStore, err := se.NewSQLIndexStore(db, dialect, indexTablePrefix)
	if err != nil {
		return nil, err
	}

	s := &sqlCollection{
		db:      db,
		dialect: dialect,
		objects: objects,
		headers: headers,
		info:    collection,
		engine:  se.NewEngine(indexStore),
	}

	objects.RegisterScanner(objectScanner, bome.NewScannerFunc(s.scanFullObject))
	return s, nil
}

type sqlCollection struct {
	info    *Collection
	dialect string
	db      *sql.DB
	engine  *se.Engine

	indexes []*se.Index

	objects *bome.JSONMap
	headers *bome.JSONMap
}

func (s *sqlCollection) Objects() *bome.JSONMap {
	return s.objects
}

func (s *sqlCollection) Save(ctx context.Context, object *Object, indexes ...*se.TextIndex) error {
	if object.Header.CreatedAt == 0 {
		object.Header.CreatedAt = utime.Now()
	}

	var err error
	var objects *bome.JSONMap
	var headers *bome.JSONMap

	ctx, objects, err = s.objects.Transaction(ctx)
	if err != nil {
		logs.Error("Save: could not start objects DB transaction", logs.Err(err))
		return errors.Internal
	}

	// Save object
	err = objects.Save(&bome.MapEntry{
		Key:   object.Header.Id,
		Value: object.Data,
	})
	if err != nil {
		logs.Error("Save: failed to save object data", logs.Err(err))
		if err := bome.Rollback(ctx); err != nil {
			logs.Error("Save: rollback failed", logs.Err(err))
		}
		return errors.Internal
	}

	// Then retrieve saved size
	size, err := objects.Map.Size(object.Header.Id)
	if err != nil {
		logs.Error("Patch: failed to get object size", logs.Details("id", object.Header.Id), logs.Err(err))
		if err := objects.Rollback(); err != nil {
			logs.Error("Patch: rollback failed", logs.Err(err))
		}
		return err
	}

	// Update object header size
	object.Header.Size = size
	headersData, err := json.Marshal(object.Header)
	if err != nil {
		logs.Error("Save: could not get object header", logs.Err(err))
		if err2 := bome.Rollback(ctx); err2 != nil {
			logs.Error("Save: rollback failed", logs.Err(err2))
		}
		return errors.BadInput
	}

	// Save object header
	ctx, headers, err = s.headers.Transaction(ctx)
	if err != nil {
		logs.Error("Save: failed to continue transactions with headers", logs.Err(err))
		if err2 := bome.Rollback(ctx); err2 != nil {
			logs.Error("Save: rollback failed", logs.Err(err2))
		}
		return err
	}

	err = headers.Save(&bome.MapEntry{
		Key:   object.Header.Id,
		Value: string(headersData),
	})
	if err != nil {
		logs.Error("Save: failed to save object headers", logs.Err(err))
		if err2 := bome.Rollback(ctx); err2 != nil {
			logs.Error("Save: rollback failed", logs.Err(err2))
		}
		return err
	}

	textIndexes := append(s.info.TextIndexes, indexes...)
	for _, index := range textIndexes {
		result := gjson.Get(object.Data, strings.TrimPrefix(index.Path, "$."))
		if !result.Exists() {
			logs.Error("Save: Text index references path that does not exists", logs.Details("path", index.Path))
			continue
		}

		if result.Type != gjson.String {
			logs.Error("Save: Text index supports only text field", logs.Err(err))
			if err2 := headers.Rollback(); err2 != nil {
				logs.Error("Save: rollback failed", logs.Err(err2))
			}
			return errors.BadInput
		}

		mp := &se.TextMapping{
			Text:     result.Str,
			Name:     index.Alias,
			ObjectId: object.Header.Id,
		}
		err = s.engine.CreateTextMapping(mp)
		if err != nil {
			logs.Error("Save: failed to create text mapping", logs.Details("path", index.Path), logs.Details("data", object.Data), logs.Err(err))
			if err2 := bome.Rollback(ctx); err2 != nil {
				logs.Error("Save: rollback failed", logs.Err(err2))
			}
			return errors.BadInput
		}
	}

	if s.info.NumberIndex != nil {
		result := gjson.Get(object.Data, strings.TrimPrefix(s.info.NumberIndex.Path, "$."))
		if !result.Exists() {
			logs.Error("Save: Number index references path that does not exists", logs.Details("path", s.info.NumberIndex.Path))
		} else {
			if result.Type != gjson.Number {
				logs.Error("Save: Number index supports only number field", logs.Err(err))
				if err2 := bome.Rollback(ctx); err2 != nil {
					logs.Error("Save: rollback failed", logs.Err(err2))
				}
				return errors.BadInput
			}

			mp := &se.NumberMapping{
				Number:   result.Int(),
				Name:     s.info.NumberIndex.Alias,
				ObjectId: object.Header.Id,
			}
			err = s.engine.CreateNumberMapping(mp)
			if err != nil {
				logs.Error("Save: failed to create number mapping", logs.Err(err))
				if err2 := bome.Rollback(ctx); err2 != nil {
					logs.Error("Save: rollback failed", logs.Err(err2))
				}
				return errors.BadInput
			}
		}
	}

	if s.info.FieldsIndex != nil && len(s.info.FieldsIndex.Aliases) > 0 {
		props := map[string]interface{}{}
		for path, alias := range s.info.FieldsIndex.Aliases {
			result := gjson.Get(object.Data, strings.TrimPrefix(path, "$."))
			if !result.Exists() {
				logs.Error("Save: Field index references path that does not exists", logs.Details("path", path))
				continue
			}

			if result.Type == gjson.JSON {
				logs.Error("Save: Text index supports only text text, number and boolean", logs.Err(err))
				if err2 := bome.Rollback(ctx); err2 != nil {
					logs.Error("Save: rollback failed", logs.Err(err2))
				}
				return errors.BadInput
			}
			props[alias] = result.Value()
		}

		value, err := json.Marshal(props)
		if err != nil {
			logs.Error("Save: could not create properties index", logs.Err(err))
			if err2 := bome.Rollback(ctx); err2 != nil {
				logs.Error("Save: rollback failed", logs.Err(err2))
			}
			return errors.BadInput
		}

		mp := &se.PropertiesMapping{
			ObjectId: object.Header.Id,
			Json:     string(value),
		}
		err = s.engine.CreatePropertiesMapping(mp)
		if err != nil {
			logs.Error("Save: failed to create fields mapping", logs.Err(err))
			if err2 := bome.Rollback(ctx); err2 != nil {
				logs.Error("Save: rollback failed", logs.Err(err2))
			}
			return errors.BadInput
		}
	}

	err = bome.Commit(ctx)
	if err != nil {
		logs.Error("Save: operations commit failed", logs.Err(err))
		return errors.Internal
	}
	logs.Debug("Save: object saved", logs.Details("id", object.Header.Id))
	return nil
}

func (s *sqlCollection) Patch(ctx context.Context, patch *Patch) error {

	value := sqlJSONSetValue(patch.Data)

	txCtx, objects, err := s.objects.Transaction(ctx)
	if err != nil {
		logs.Error("Patch: could not start objects DB transaction", logs.Err(err))
		return errors.Internal
	}

	err = objects.EditAt(patch.ObjectId, patch.At, bome.RawExpr(value))
	if err != nil {
		logs.Error("Update: object patch failed", logs.Details("id", patch.ObjectId), logs.Err(err))
		return errors.Internal
	}

	size, err := objects.Map.Size(patch.ObjectId)
	if err != nil {
		logs.Error("Patch: failed to get object size", logs.Details("id", patch.ObjectId), logs.Err(err))
		if err := bome.Rollback(txCtx); err != nil {
			logs.Error("Patch: rollback failed", logs.Err(err))
		}
		return err
	}

	_, headers, _ := s.headers.Transaction(txCtx)
	err = headers.EditAt(patch.ObjectId, "$.size", bome.IntExpr(size))
	if err != nil {
		logs.Error("Patch: failed to save object headers", logs.Err(err))
		if err := bome.Rollback(txCtx); err != nil {
			logs.Error("Patch: rollback failed", logs.Err(err))
		}
		return errors.Internal
	}

	err = bome.Commit(txCtx)
	if err != nil {
		logs.Error("Patch: operations commit failed", logs.Err(err))
		return errors.Internal
	}

	logs.Debug("Patch: object updated", logs.Details("id", patch.ObjectId))
	return nil
}

func (s *sqlCollection) Delete(ctx context.Context, objectID string) error {
	go func() {
		if der := s.engine.DeleteObjectMappings(objectID); der != nil {
			logs.Error("failed to delete object index mappings", logs.Err(der))
		}
	}()

	err := s.objects.Delete(objectID)
	if err != nil {
		logs.Error("Delete: object deletion failed", logs.Err(err))
		return errors.Internal
	}

	logs.Debug("Delete: object deleted", logs.Details("id", objectID))
	return nil
}

func (s *sqlCollection) Get(ctx context.Context, objectID string, opts GetOptions) (*Object, error) {
	hv, err := s.headers.Get(objectID)
	if err != nil {
		logs.Error("Get: could not get object header", logs.Details("id", objectID), logs.Err(err))
		if bome.IsNotFound(err) {
			return nil, err
		}
		return nil, errors.Internal
	}

	o := &Object{
		Header: &Header{},
	}

	err = json.Unmarshal([]byte(hv), o.Header)
	if err != nil {
		logs.Error("Get: could not decode object header", logs.Err(err))
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

	logs.Debug("Get: loaded object", logs.Details("id", objectID))
	return o, nil
}

func (s *sqlCollection) Info(ctx context.Context, id string) (*Header, error) {
	value, err := s.headers.Get(id)
	if err != nil {
		return nil, err
	}

	var info Header
	err = json.Unmarshal([]byte(value), &info)
	if err != nil {
		logs.Error("List: failed to decode object header", logs.Details("encoded", value), logs.Err(err))
		return nil, errors.Internal
	}
	return &info, nil
}

func (s *sqlCollection) List(ctx context.Context, opts ListOptions) (*Cursor, error) {
	sqlQuery := fmt.Sprintf("select headers.value as header, objects.value as object from %s as headers, %s as objects limit ?, 50", s.headers.Table(), s.objects.Table())
	cursor, err := s.objects.Query(sqlQuery, objectScanner, opts.Offset)
	if err != nil {
		return nil, err
	}

	closer := CloseFunc(func() error {
		return cursor.Close()
	})
	browser := BrowseFunc(func() (*Object, error) {
		if !cursor.HasNext() {
			return nil, io.EOF
		}
		next, err := cursor.Next()
		if err != nil {
			return nil, err
		}
		return next.(*Object), err
	})
	return NewCursor(browser, closer), nil
}

func (s *sqlCollection) Search(ctx context.Context, query *se.SearchQuery) (*Cursor, error) {
	ids, err := s.engine.Search(query)
	if err != nil {
		return nil, err
	}

	c := &idsListCursor{
		ids: ids,
		getObjectFunc: func(id string) (*Object, error) {
			return s.Get(ctx, id, GetOptions{})
		},
	}
	return NewCursor(c, c), nil
}

func (s *sqlCollection) Clear() error {
	ctx, objects, err := s.objects.Transaction(context.Background())
	if err != nil {
		logs.Error("Clear: could not start transaction in objects DB", logs.Err(err))
		return errors.Internal
	}

	err = objects.Clear()
	if err != nil {
		logs.Error("Clear: could not clear objects", logs.Err(err))
		if err := bome.Rollback(ctx); err != nil {
			logs.Error("Clear: operations rollback failed", logs.Err(err))
		}
		return errors.Internal
	}

	_, headers, _ := s.headers.Transaction(ctx)
	err = headers.Clear()
	if err != nil {
		logs.Error("Clear: could not clear objects headers", logs.Err(err))
		if err := bome.Rollback(ctx); err != nil {
			logs.Error("Clear: operations rollback failed", logs.Err(err))
		}
		return errors.Internal
	}

	if err := bome.Commit(ctx); err != nil {
		logs.Error("Clear: operations commit failed", logs.Err(err))
	}

	logs.Debug("Clear: objects store has been cleared")
	return nil
}

func (s *sqlCollection) validateSearchQuery(query *se.SearchQuery) bool {
	switch q := query.Query.(type) {
	case *se.SearchQuery_Fields:
		return s.validatePropertiesSearchingQuery(q.Fields)
	default:
		return true
	}
}

func (s *sqlCollection) validatePropertiesSearchingQuery(query *se.FieldQuery) bool {
	var fieldQueries []*se.FieldQuery
	fieldQueries = append(fieldQueries, query)
	for len(fieldQueries) > 0 {

		q := fieldQueries[0]
		fieldQueries = fieldQueries[1:]

		switch v := q.Bool.(type) {

		case *se.FieldQuery_And:
			for _, ox := range v.And.Queries {
				fieldQueries = append(fieldQueries, ox)
			}
			continue

		case *se.FieldQuery_Or:
			for _, ox := range v.Or.Queries {
				fieldQueries = append(fieldQueries, ox)
			}
			continue

		case *se.FieldQuery_Contains:
			if !s.indexFieldExists(v.Contains.Field) {
				return false
			}

		case *se.FieldQuery_StartsWith:
			if !s.indexFieldExists(v.StartsWith.Field) {
				return false
			}

		case *se.FieldQuery_EndsWith:
			if !s.indexFieldExists(v.EndsWith.Field) {
				return false
			}

		case *se.FieldQuery_StrEqual:
			if !s.indexFieldExists(v.StrEqual.Field) {
				return false
			}

		case *se.FieldQuery_Lt:
			if !s.indexFieldExists(v.Lt.Field) {
				return false
			}

		case *se.FieldQuery_Lte:
			if !s.indexFieldExists(v.Lte.Field) {
				return false
			}

		case *se.FieldQuery_Gt:
			if !s.indexFieldExists(v.Gt.Field) {
				return false
			}

		case *se.FieldQuery_Gte:
			if !s.indexFieldExists(v.Gte.Field) {
				return false
			}

		case *se.FieldQuery_NumbEq:
			if !s.indexFieldExists(v.NumbEq.Field) {
				return false
			}
		}
	}
	return true
}

func (s *sqlCollection) indexFieldExists(name string) bool {
	if s.info == nil || s.info.FieldsIndex != nil {
		return false
	}
	for _, alias := range s.info.FieldsIndex.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}

func (s *sqlCollection) scanFullObject(row bome.Row) (interface{}, error) {
	var header, data string
	err := row.Scan(&header, &data)
	if err != nil {
		return nil, err
	}
	object := &Object{
		Data: data,
	}
	err = json.NewDecoder(bytes.NewBufferString(header)).Decode(&object.Header)
	return object, err
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
