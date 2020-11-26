package oms

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/cel-go/cel"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"strings"
	"time"
)

func NewStore(db *sql.DB) (Store, error) {
	m, err := bome.NewJSONMap(db, bome.MySQL, "objects")
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
	}
	return s, nil
}

type mysqlStore struct {
	objects   *bome.JSONMap
	datedRefs *bome.List
	cEnv      *cel.Env
}

func (ms *mysqlStore) Save(ctx context.Context, object *Object) error {
	d := time.Now().Unix()
	object.SetCreatedAt(d)

	data, err := object.Marshal()
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
		Value: string(data),
	})
	if err != nil {
		log.Error("Save: failed to save object data", log.Err(err))
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

	err = ltx.Commit()
	if err != nil {
		log.Error("Save: operations commit failed", log.Err(err))
		return errors.Internal
	}

	log.Debug("Save: object saved", log.Field("id", object.ID()))
	return nil
}

func (ms *mysqlStore) Update(ctx context.Context, patch *Patch) error {
	data, err := patch.Marshal()
	if err != nil {
		log.Error("Update: could not get patch content", log.Err(err))
		return errors.BadInput
	}

	err = ms.objects.EditAt(patch.GetObjectID(), patch.path, bome.StringExpr(string(data)))
	if err != nil {
		log.Error("Update: object patch failed", log.Field("id", patch.objectID), log.Err(err))
		return errors.Internal
	}

	log.Debug("Update: object updated", log.Field("id", patch.GetObjectID()))
	return nil
}

func (ms *mysqlStore) Delete(ctx context.Context, objectID string) error {
	object, err := ms.Get(ctx, objectID, DataOptions{})
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

func (ms *mysqlStore) List(ctx context.Context, opts ListOptions) (*ListResult, error) {
	entries, err := ms.objects.Range(opts.Offset, opts.Count)
	if err != nil {
		log.Error("List: failed to get objects",
			log.Field("offset", opts.Offset),
			log.Field("", opts.Count), log.Err(err))
		return nil, errors.Internal
	}

	var result ListResult
	for _, entry := range entries {
		o, err := DecodeObject(entry.Value)
		if err != nil {
			log.Error("List: failed to decode item", log.Field("encoded", entry.Value), log.Err(err))
			return nil, errors.Internal
		}
		result.Objects = append(result.Objects, o)
	}
	return &result, nil
}

func (ms *mysqlStore) Get(ctx context.Context, id string, opts DataOptions) (*Object, error) {
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

func (ms *mysqlStore) Search(ctx context.Context, opts SearchOptions) (*SearchResult, error) {
	cursor, err := ms.datedRefs.GetAllFromSeq(opts.Before)
	if err != nil {
		return nil, err
	}

	result := &SearchResult{}
	for cursor.HasNext() && len(result.Objects) < opts.Count {
		item, err := cursor.Next()
		if err != nil {
			return nil, err
		}

		id := item.(string)
		o, err := ms.Get(ctx, id, DataOptions{})
		if err != nil {
			return nil, err
		}

		if opts.Filter != nil {
			canRead, err := opts.Filter.Filter(o)
			if err != nil {
				return nil, err
			}

			if !canRead {
				continue
			}
		}

		result.Objects = append(result.Objects, o)
	}

	return result, nil
}

func escaped(value string) string {
	replace := map[string]string{"\\": "\\\\", "'": `\'`, "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`, "\x1a": "\\Z"}
	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}
	return value
}

func clauseFromCondition(condition *any.Any) (string, error) {
	switch strings.ToLower(condition.TypeUrl) {
	case "eval":
		return evalWhereClause(condition)
	case "or", "and":
		return operatorCondition(condition)
	case "not":
		return operatorCondition(condition)
	default:
		log.Info("operator not supported", log.Field("name", condition.TypeUrl))
		return "", errors.NotSupported
	}
}

func operatorCondition(any *any.Any) (string, error) {
	b := strings.Builder{}
	var conditions Conditions
	err := ptypes.UnmarshalAny(any, &conditions)
	if err != nil {
		return "", err
	}

	condition := conditions.Items[0]
	strCond, err := clauseFromCondition(condition)
	if err != nil {
		return "", err
	}

	b.WriteString("(")
	b.WriteString(strCond)

	for _, condition := range conditions.Items[1:] {
		strCond, err := clauseFromCondition(condition)
		if err != nil {
			return "", err
		}
		b.WriteString(" ")
		b.WriteString(any.TypeUrl)
		b.WriteString(" ")
		b.WriteString(strCond)
	}
	return b.String(), nil
}

func evalWhereClause(any *any.Any) (string, error) {
	var p OperationParams
	err := ptypes.UnmarshalAny(any, &p)
	if err != nil {
		return "", err
	}

	switch p.Func {
	case "has":
		return fmt.Sprintf("json_contains(value, '\"%s\"', '\"%s\"')", escaped(p.Path), escaped(p.Value)), nil
	case "ex":
		return fmt.Sprintf("json_contains_path(value, '\"%s\"')", escaped(p.Path)), nil
	case "eq":
		return fmt.Sprintf("json_extracts(value, '\"%s\"')=='\"%s\"'", escaped(p.Path), escaped(p.Value)), nil
	case "eqn":
		return fmt.Sprintf("json_extracts(value, '\"%s\"')==%s", escaped(p.Path), escaped(p.Value)), nil
	default:
		log.Info("function not supported", log.Field("name", p.Func))
		return "", errors.NotSupported
	}
}
