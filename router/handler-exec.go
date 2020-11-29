package router

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/oms"
)

type execHandler struct {
	base
}

func (e *execHandler) ListWorkers(ctx context.Context) ([]*oms.JSON, error) {
	db := workersDB(ctx)
	if db == nil {
		log.Info("missing worker info db in context")
		return nil, errors.Internal
	}

	c, err := db.List()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Error("Workers: failed to close cursor", log.Err(err))
		}
	}()

	var infoList []*oms.JSON
	for c.HasNext() {
		o, err := c.Next()
		if err != nil {
			return nil, err
		}

		entry := o.(*bome.MapEntry)
		var object interface{}
		err = json.Unmarshal([]byte(entry.Value), object)
		if err != nil {
			return nil, err
		}
		infoList = append(infoList, oms.NewJSON(object))
	}
	return infoList, nil
}

func (e *execHandler) RegisterWorker(ctx context.Context, info *oms.JSON) error {
	db := workersDB(ctx)
	if db == nil {
		log.Info("missing worker info db in context")
		return errors.Internal
	}

	name, err := info.StringAt("$.name")
	if err != nil {
		return err
	}

	entry := &bome.MapEntry{
		Key:   name,
		Value: info.String(),
	}
	return db.Save(entry)
}

func (e *execHandler) SetSettings(ctx context.Context, value *oms.JSON, opts oms.SettingsOptions) error {
	s := settings(ctx)
	if s == nil {
		log.Info("missing settings database in context")
		return errors.Internal
	}

	settings, err := json.Marshal(value)
	if err != nil {
		return err
	}

	entry := &bome.MapEntry{
		Key:   oms.Settings,
		Value: string(settings),
	}
	return s.Save(entry)
}

func (e *execHandler) GetSettings(ctx context.Context, opts oms.SettingsOptions) (*oms.JSON, error) {
	s := settings(ctx)
	if s == nil {
		log.Info("missing settings database in context")
		return nil, errors.Internal
	}

	value, err := s.ExtractAt(oms.Settings, opts.Path)
	if err != nil {
		return nil, err
	}

	var o interface{}
	err = json.Unmarshal([]byte(value), &o)
	return oms.NewJSON(o), err
}

func (e *execHandler) PutObject(ctx context.Context, object *oms.Object, security *oms.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing storage in context")
		return "", errors.Internal
	}

	accessStore := accessStore(ctx)
	if accessStore == nil {
		log.Info("missing access store in context")
		return "", errors.Internal
	}

	id := uuid.New().String()

	err := accessStore.SaveRules(id, security)
	if err != nil {
		log.Error("PutObject-exec: failed to save object access security rules", log.Err(err))
		return "", errors.Internal
	}

	object.SetID(id)
	return id, storage.Save(ctx, object)
}

func (e *execHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	panic("implement me")
}

func (e *execHandler) GetObject(ctx context.Context, id string, opts oms.GetDataOptions) (*oms.Object, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.New("wrong context")
	}

	if opts.Path == "" {
		return storage.Get(ctx, id)
	} else {
		return storage.GetAt(ctx, id, opts.Path)
	}
}

func (e *execHandler) GetObjectHeader(ctx context.Context, id string) (*oms.Header, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.New("wrong context")
	}
	return storage.Info(ctx, id)
}

func (e *execHandler) DeleteObject(ctx context.Context, id string) error {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return errors.New("wrong context")
	}
	return storage.Delete(ctx, id)
}

func (e *execHandler) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.New("wrong context")
	}
	return storage.List(ctx, opts.Before, opts.Count, opts.Filter)
}
