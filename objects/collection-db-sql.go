package objects

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/pb"
	se "github.com/omecodes/store/search"
	"github.com/omecodes/store/utime"
	"github.com/tidwall/gjson"
	"io"
	"strings"
)

const objectScanner = "object"

func NewSQLCollection(collection *pb.Collection, db *sql.DB, dialect string, tableName string) (*sqlCollection, error) {
	objects, err := bome.NewJSONMap(db, dialect, "store_"+tableName+"collection")
	if err != nil {
		return nil, err
	}

	headers, err := bome.NewJSONMap(db, dialect, "store_"+tableName+"_col_headers")
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

	indexStore, err := se.NewSQLIndexStore(db, dialect, "store_"+tableName+"_col")
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
	info    *pb.Collection
	dialect string
	db      *sql.DB
	engine  *se.Engine

	indexes []*pb.Index

	objects *bome.JSONMap
	headers *bome.JSONMap
}

func (s *sqlCollection) Objects() *bome.JSONMap {
	return s.objects
}

func (s *sqlCollection) Save(ctx context.Context, object *pb.Object, indexes ...*pb.TextIndex) error {
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

	textIndexes := append(s.info.TextIndexes, indexes...)
	for _, index := range textIndexes {
		result := gjson.Get(object.Data, strings.TrimPrefix(index.Path, "$."))
		if !result.Exists() {
			log.Error("Save: Text index references path that does not exists", log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return errors.BadInput
		}

		if result.Type != gjson.String {
			log.Error("Save: Text index supports only text field", log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return errors.BadInput
		}

		mp := &pb.TextMapping{
			Text:     result.Str,
			Name:     index.Alias,
			ObjectId: object.Header.Id,
		}
		err = s.engine.CreateTextMapping(mp)
		if err != nil {
			log.Error("Save: failed to create text mapping", log.Field("path", index.Path), log.Field("data", object.Data), log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return errors.BadInput
		}
	}

	if s.info.NumberIndex != nil {
		result := gjson.Get(object.Data, strings.TrimPrefix(s.info.NumberIndex.Path, "$."))
		if !result.Exists() {
			log.Error("Save: Number index references path that does not exists", log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return errors.BadInput
		}

		if result.Type != gjson.Number {
			log.Error("Save: Number index supports only number field", log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return errors.BadInput
		}

		mp := &pb.NumberMapping{
			Number:   result.Int(),
			Name:     s.info.NumberIndex.Alias,
			ObjectId: object.Header.Id,
		}
		err = s.engine.CreateNumberMapping(mp)
		if err != nil {
			log.Error("Save: failed to create number mapping", log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return errors.BadInput
		}
	}

	if s.info.FieldsIndex != nil && len(s.info.FieldsIndex.Aliases) > 0 {
		props := map[string]interface{}{}
		for path, alias := range s.info.FieldsIndex.Aliases {
			result := gjson.Get(object.Data, strings.TrimPrefix(path, "$."))
			if !result.Exists() {
				log.Error("Save: Field index references path that does not exists", log.Err(err))
				if err2 := tx.Rollback(); err2 != nil {
					log.Error("Save: rollback failed", log.Err(err2))
				}
				return errors.BadInput
			}

			if result.Type == gjson.JSON {
				log.Error("Save: Text index supports only text text, number and boolean", log.Err(err))
				if err2 := tx.Rollback(); err2 != nil {
					log.Error("Save: rollback failed", log.Err(err2))
				}
				return errors.BadInput
			}
			props[alias] = result.Value()
		}

		value, err := json.Marshal(props)
		if err != nil {
			log.Error("Save: could not create properties index", log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return errors.BadInput
		}

		mp := &pb.PropertiesMapping{
			ObjectId: object.Header.Id,
			Json:     string(value),
		}
		err = s.engine.CreatePropertiesMapping(mp)
		if err != nil {
			log.Error("Save: failed to create fields mapping", log.Err(err))
			if err2 := tx.Rollback(); err2 != nil {
				log.Error("Save: rollback failed", log.Err(err2))
			}
			return errors.BadInput
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
	go func() {
		if der := s.engine.DeleteObjectMappings(objectID); der != nil {
			log.Error("failed to delete object index mappings", log.Err(der))
		}
	}()

	err := s.objects.Delete(objectID)
	if err != nil {
		log.Error("Delete: object deletion failed", log.Err(err))
		return errors.Internal
	}

	log.Debug("Delete: object deleted", log.Field("id", objectID))
	return nil
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
	sqlQuery := fmt.Sprintf("select headers.value as header, objects.value as object from %s as headers, %s as objects limit ?, 50", s.headers.Table(), s.objects.Table())
	cursor, err := s.objects.RawQuery(sqlQuery, objectScanner, opts.Offset)
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
		return next.(*pb.Object), err
	})
	return pb.NewCursor(browser, closer), nil
}

func (s *sqlCollection) Search(ctx context.Context, query *pb.SearchQuery) (*pb.Cursor, error) {
	ids, err := s.engine.Search(query)
	if err != nil {
		return nil, err
	}

	c := &idsListCursor{
		ids: ids,
		getObjectFunc: func(id string) (*pb.Object, error) {
			return s.Get(ctx, id, pb.GetOptions{})
		},
	}
	return pb.NewCursor(c, c), nil
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

func (s *sqlCollection) validateSearchQuery(query *pb.SearchQuery) bool {
	switch q := query.Query.(type) {
	case *pb.SearchQuery_Fields:
		return s.validatePropertiesSearchingQuery(q.Fields)
	default:
		return true
	}
}

func (s *sqlCollection) validatePropertiesSearchingQuery(query *pb.FieldQuery) bool {
	var fieldQueries []*pb.FieldQuery
	fieldQueries = append(fieldQueries, query)
	for len(fieldQueries) > 0 {

		q := fieldQueries[0]
		fieldQueries = fieldQueries[1:]

		switch v := q.Bool.(type) {

		case *pb.FieldQuery_And:
			for _, ox := range v.And.Queries {
				fieldQueries = append(fieldQueries, ox)
			}
			continue

		case *pb.FieldQuery_Or:
			for _, ox := range v.Or.Queries {
				fieldQueries = append(fieldQueries, ox)
			}
			continue

		case *pb.FieldQuery_Contains:
			if !s.indexFieldExists(v.Contains.Field) {
				return false
			}

		case *pb.FieldQuery_StartsWith:
			if !s.indexFieldExists(v.StartsWith.Field) {
				return false
			}

		case *pb.FieldQuery_EndsWith:
			if !s.indexFieldExists(v.EndsWith.Field) {
				return false
			}

		case *pb.FieldQuery_StrEqual:
			if !s.indexFieldExists(v.StrEqual.Field) {
				return false
			}

		case *pb.FieldQuery_Lt:
			if !s.indexFieldExists(v.Lt.Field) {
				return false
			}

		case *pb.FieldQuery_Lte:
			if !s.indexFieldExists(v.Lte.Field) {
				return false
			}

		case *pb.FieldQuery_Gt:
			if !s.indexFieldExists(v.Gt.Field) {
				return false
			}

		case *pb.FieldQuery_Gte:
			if !s.indexFieldExists(v.Gte.Field) {
				return false
			}

		case *pb.FieldQuery_NumbEq:
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
	object := &pb.Object{
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
