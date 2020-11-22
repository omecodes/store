package oms

import (
	"context"
	"io"
	"time"

	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
)

type paramsHandler struct {
	base
}

func (p *paramsHandler) SetSettings(ctx context.Context, value *JSON, opts SettingsOptions) error {
	if value == nil {
		return errors.BadInput
	}

	if opts.Path != "" {
		_, allowed := settingsPathFormats[opts.Path]
		if !allowed {
			return errors.BadInput
		}
	}

	return p.base.SetSettings(ctx, value, opts)
}

func (p *paramsHandler) GetSettings(ctx context.Context, opts SettingsOptions) (*JSON, error) {
	if opts.Path != "" {
		_, allowed := settingsPathFormats[opts.Path]
		if !allowed {
			return nil, errors.BadInput
		}
	}
	return p.base.GetSettings(ctx, opts)
}

func (p *paramsHandler) RegisterWorker(ctx context.Context, info *JSON) error {
	if info == nil {
		return errors.BadInput
	}
	return p.base.RegisterWorker(ctx, info)
}

func (p *paramsHandler) PutData(ctx context.Context, data *Object, opts PutDataOptions) error {
	if data == nil || data.Collection == "" || data.Id == "" || data.Size == 0 || data.JsonEncoded == "" {
		return errors.BadInput
	}

	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
	s, err := route.GetSettings(ctx, SettingsOptions{Path: settingsDataMaxSizePath})
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return errors.Internal
	}

	maxLength, err := s.ToInt64()
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return errors.Internal
	}

	if data.Size > maxLength {
		log.Error("could not process request. Object too big", log.Field("max", maxLength), log.Field("received", data.Size))
		return errors.BadInput
	}

	return p.next.PutData(ctx, data, opts)
}

func (p *paramsHandler) PatchData(ctx context.Context, collection string, id string, content io.Reader, size int64, opts PatchOptions) error {
	if collection == "" || id == "" || size == 0 || content == nil {
		return errors.BadInput
	}

	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
	s, err := route.GetSettings(ctx, SettingsOptions{Path: settingsDataMaxSizePath})
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return errors.Internal
	}

	maxLength, err := s.ToInt64()
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return errors.Internal
	}

	if size > maxLength {
		log.Error("could not process request. Object too big", log.Field("max", maxLength), log.Field("received", size))
		return errors.BadInput
	}

	return p.next.PatchData(ctx, collection, id, content, size, opts)
}

func (p *paramsHandler) GetData(ctx context.Context, collection string, id string, opts GetDataOptions) (*Object, error) {
	if collection == "" || id == "" {
		return nil, errors.BadInput
	}
	return p.base.GetData(ctx, collection, id, opts)
}

func (p *paramsHandler) Info(ctx context.Context, collection string, id string) (*Info, error) {
	if collection == "" || id == "" {
		return nil, errors.BadInput
	}
	return p.base.Info(ctx, collection, id)
}

func (p *paramsHandler) Delete(ctx context.Context, collection string, id string) error {
	if collection == "" || id == "" {
		return errors.BadInput
	}
	return p.next.Delete(ctx, collection, id)
}

func (p *paramsHandler) List(ctx context.Context, collection string, opts ListOptions) (*DataList, error) {
	if collection == "" {
		return nil, errors.BadInput
	}
	return p.base.List(ctx, collection, opts)
}

func (p *paramsHandler) SaveGraft(ctx context.Context, graft *Graft) (string, error) {
	if graft == nil || graft.Collection == "" || graft.DataId == "" || graft.Content == "" || graft.Size == 0 {
		return "", errors.BadInput
	}

	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
	s, err := route.GetSettings(ctx, SettingsOptions{Path: settingsDataMaxSizePath})
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return "", errors.Internal
	}

	maxLength, err := s.ToInt64()
	if err != nil {
		log.Error("could not get data max length from settings", log.Err(err))
		return "", errors.Internal
	}

	if graft.Size > maxLength {
		log.Error("could not process request. Object too big", log.Field("max", maxLength), log.Field("received", graft.Size))
		return "", errors.BadInput
	}

	return p.next.SaveGraft(ctx, graft)
}

func (p *paramsHandler) GetGraft(ctx context.Context, collection string, dataID string, id string) (*Graft, error) {
	if collection == "" || dataID == "" || id == "" {
		return nil, errors.BadInput
	}
	return p.base.GetGraft(ctx, collection, dataID, id)
}

func (p *paramsHandler) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*GraftInfo, error) {
	if collection == "" || dataID == "" || id == "" {
		return nil, errors.BadInput
	}
	return p.base.GraftInfo(ctx, collection, dataID, id)
}

func (p *paramsHandler) ListGrafts(ctx context.Context, collection string, dataID string, opts ListOptions) (*GraftList, error) {
	if collection == "" || dataID == "" {
		return nil, errors.BadInput
	}
	opts.Count = 5
	if opts.Before == 0 {
		opts.Before = time.Now().Unix()
	}
	return p.base.ListGrafts(ctx, collection, dataID, opts)
}

func (p *paramsHandler) DeleteGraft(ctx context.Context, collection string, dataID string, id string) error {
	if collection == "" || dataID == "" || id == "" {
		return errors.BadInput
	}
	return p.base.DeleteGraft(ctx, collection, dataID, id)
}
