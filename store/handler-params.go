package store

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/ent"
	"github.com/omecodes/omestore/pb"
	"io"
	"time"
)

type paramsHandler struct {
	base
}

func (p *paramsHandler) SetSettings(ctx context.Context, value *JSON, opts pb.SettingsOptions) error {
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

func (p *paramsHandler) GetSettings(ctx context.Context, opts pb.SettingsOptions) (*JSON, error) {
	if opts.Path != "" {
		_, allowed := settingsPathFormats[opts.Path]
		if !allowed {
			return nil, errors.BadInput
		}
	}
	return p.base.GetSettings(ctx, opts)
}

func (p *paramsHandler) CreateUser(ctx context.Context, user *ent.User) error {
	if user == nil || user.ID == "" || user.Email == "" {
		return errors.BadInput
	}
	return p.base.CreateUser(ctx, user)
}

func (p *paramsHandler) UserInfo(ctx context.Context, username string, opts pb.UserOptions) (*ent.User, error) {
	if username == "" {
		return nil, errors.BadInput
	}
	return p.base.UserInfo(ctx, username, opts)
}

func (p *paramsHandler) ValidateUser(ctx context.Context, username string, opts pb.UserOptions) error {
	if username == "" {
		return errors.BadInput
	}
	return p.base.ValidateUser(ctx, username, opts)
}

func (p *paramsHandler) RegisterUser(ctx context.Context, user *ent.User, opts pb.UserOptions) error {
	if user == nil || user.Password == "" || user.Email == "" || user.ID == "" {
		return errors.BadInput
	}
	return p.base.RegisterUser(ctx, user, opts)
}

func (p *paramsHandler) PutData(ctx context.Context, data *pb.Data, opts pb.PutDataOptions) error {
	if data == nil || data.Collection == "" || data.ID == "" || data.Size == 0 || data.Content == "" {
		return errors.BadInput
	}

	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
	s, err := route.GetSettings(ctx, pb.SettingsOptions{Path: settingsDataMaxSizePath})
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
		log.Error("could not process request. Data too big", log.Field("max", maxLength), log.Field("received", data.Size))
		return errors.BadInput
	}

	return p.next.PutData(ctx, data, opts)
}

func (p *paramsHandler) PatchData(ctx context.Context, collection string, id string, content io.Reader, size int64, opts pb.PatchOptions) error {
	if collection == "" || id == "" || size == 0 || content == nil {
		return errors.BadInput
	}

	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
	s, err := route.GetSettings(ctx, pb.SettingsOptions{Path: settingsDataMaxSizePath})
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
		log.Error("could not process request. Data too big", log.Field("max", maxLength), log.Field("received", size))
		return errors.BadInput
	}

	return p.next.PatchData(ctx, collection, id, content, size, opts)
}

func (p *paramsHandler) GetData(ctx context.Context, collection string, id string, opts pb.GetDataOptions) (*pb.Data, error) {
	if collection == "" || id == "" {
		return nil, errors.BadInput
	}
	return p.base.GetData(ctx, collection, id, opts)
}

func (p *paramsHandler) Info(ctx context.Context, collection string, id string) (*pb.Info, error) {
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

func (p *paramsHandler) List(ctx context.Context, collection string, opts pb.ListOptions) (*pb.DataList, error) {
	if collection == "" {
		return nil, errors.BadInput
	}
	return p.base.List(ctx, collection, opts)
}

func (p *paramsHandler) SaveGraft(ctx context.Context, graft *pb.Graft) (string, error) {
	if graft == nil || graft.Collection == "" || graft.DataID == "" || graft.Content == "" || graft.Size == 0 {
		return "", errors.BadInput
	}

	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
	s, err := route.GetSettings(ctx, pb.SettingsOptions{Path: settingsDataMaxSizePath})
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
		log.Error("could not process request. Data too big", log.Field("max", maxLength), log.Field("received", graft.Size))
		return "", errors.BadInput
	}

	return p.next.SaveGraft(ctx, graft)
}

func (p *paramsHandler) GetGraft(ctx context.Context, collection string, dataID string, id string) (*pb.Graft, error) {
	if collection == "" || dataID == "" || id == "" {
		return nil, errors.BadInput
	}
	return p.base.GetGraft(ctx, collection, dataID, id)
}

func (p *paramsHandler) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*pb.GraftInfo, error) {
	if collection == "" || dataID == "" || id == "" {
		return nil, errors.BadInput
	}
	return p.base.GraftInfo(ctx, collection, dataID, id)
}

func (p *paramsHandler) ListGrafts(ctx context.Context, collection string, dataID string, opts pb.ListOptions) (*pb.GraftList, error) {
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
