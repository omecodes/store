package oms

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/libome/v2"
)

type execHandler struct {
	base
}

func (e *execHandler) ListWorkers(ctx context.Context) ([]*JSON, error) {
	db := getWorkerInfoDB(ctx)
	if db == nil {
		log.Info("missing worker info db in context")
		return nil, errors.Internal
	}

	c, err := db.List()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var infoList []*JSON
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
		infoList = append(infoList, &JSON{object: object})
	}
	return infoList, nil
}

func (e *execHandler) RegisterWorker(ctx context.Context, info *JSON) error {
	db := getWorkerInfoDB(ctx)
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

func (e *execHandler) PatchData(ctx context.Context, collection string, id string, content io.Reader, size int64, opts PatchOptions) error {
	panic("implement me")
}

func (e *execHandler) SetSettings(ctx context.Context, value *JSON, opts SettingsOptions) error {
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
		Key:   Settings,
		Value: string(settings),
	}
	return s.Save(entry)
}

func (e *execHandler) GetSettings(ctx context.Context, opts SettingsOptions) (*JSON, error) {
	s := settings(ctx)
	if s == nil {
		log.Info("missing settings database in context")
		return nil, errors.Internal
	}

	value, err := s.ExtractAt(Settings, opts.Path)
	if err != nil {
		return nil, err
	}

	var o interface{}
	err = json.Unmarshal([]byte(value), &o)
	return &JSON{object: o}, err
}

func (e *execHandler) GetCollections(ctx context.Context) ([]string, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing storage in context")
		return nil, errors.Internal
	}
	return storage.Collections(ctx)
}

func (e *execHandler) PutData(ctx context.Context, data *Object, opts PutDataOptions) error {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing storage in context")
		return errors.Internal
	}

	userInfo := authInfo(ctx)
	if userInfo == nil {
		log.Error("could not get db from context")
		return errors.Internal
	}

	data.CreatedBy = userInfo.Uid
	data.CreatedAt = time.Now().Unix()

	return storage.Save(ctx, data)
}

func (e *execHandler) GetData(ctx context.Context, collection string, id string, opts GetDataOptions) (*Object, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.New("wrong context")
	}
	return storage.Get(ctx, collection, id, DataOptions{Path: opts.Path})
}

func (e *execHandler) Info(ctx context.Context, collection string, id string) (*Info, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.New("wrong context")
	}
	return storage.Info(ctx, collection, id)
}

func (e *execHandler) Delete(ctx context.Context, collection string, id string) error {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return errors.New("wrong context")
	}
	return storage.Delete(ctx, &Object{
		Collection: collection,
		Id:         id,
	})
}

func (e *execHandler) List(ctx context.Context, collection string, opts ListOptions) (*DataList, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.New("wrong context")
	}
	return storage.List(ctx, collection, opts)
}

func (e *execHandler) SaveGraft(ctx context.Context, graft *Graft) (string, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Error("could not get storage from context")
		return "", errors.Internal
	}

	graft.CreatedAt = time.Now().Unix()
	cred := ome.CredentialsFromContext(ctx)
	if cred != nil {
		graft.CreatedBy = cred.Username
	}
	return storage.SaveGraft(ctx, graft)
}

func (e *execHandler) GetGraft(ctx context.Context, collection string, dataID string, id string) (*Graft, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Error("could not get storage from context")
		return nil, errors.Internal
	}
	return storage.GetGraft(ctx, collection, dataID, id)
}

func (e *execHandler) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*GraftInfo, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Error("could not get storage from context")
		return nil, errors.Internal
	}
	return storage.GraftInfo(ctx, collection, dataID, id)
}

func (e *execHandler) ListGrafts(ctx context.Context, collection string, dataID string, opts ListOptions) (*GraftList, error) {
	storage := storage(ctx)
	if storage == nil {
		log.Error("could not get storage from context")
		return nil, errors.Internal
	}
	return storage.GetAllGraft(ctx, collection, dataID, opts)
}

func (e *execHandler) DeleteGraft(ctx context.Context, collection string, dataID string, id string) error {
	storage := storage(ctx)
	if storage == nil {
		log.Error("could not get storage from context")
		return errors.Internal
	}
	return storage.DeleteGraft(ctx, collection, dataID, id)
}
