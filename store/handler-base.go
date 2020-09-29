package store

import (
	"context"
	"errors"
	"github.com/omecodes/omestore/ent"
	"github.com/omecodes/omestore/pb"
	"io"
)

type base struct {
	next Handler
}

func (b *base) SetSettings(ctx context.Context, value *JSON, opts pb.SettingsOptions) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.SetSettings(ctx, value, opts)
}

func (b *base) GetSettings(ctx context.Context, opts pb.SettingsOptions) (*JSON, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.GetSettings(ctx, opts)
}

func (b *base) RegisterUser(ctx context.Context, user *ent.User, opts pb.UserOptions) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.RegisterUser(ctx, user, opts)
}

func (b *base) ListUsers(ctx context.Context, opts pb.UserOptions) ([]*ent.User, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.ListUsers(ctx, opts)
}

func (b *base) UserInfo(ctx context.Context, username string, opts pb.UserOptions) (*ent.User, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.UserInfo(ctx, username, opts)
}

func (b *base) CreateUser(ctx context.Context, user *ent.User) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.CreateUser(ctx, user)
}

func (b *base) ValidateUser(ctx context.Context, username string, opts pb.UserOptions) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.ValidateUser(ctx, username, opts)
}

func (b *base) GetCollections(ctx context.Context) ([]string, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.GetCollections(ctx)
}

func (b *base) PutData(ctx context.Context, data *pb.Data, opts pb.PutDataOptions) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.PutData(ctx, data, opts)
}

func (b *base) GetData(ctx context.Context, collection string, id string, opts pb.GetDataOptions) (*pb.Data, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.GetData(ctx, collection, id, opts)
}

func (b *base) Info(ctx context.Context, collection string, id string) (*pb.Info, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.Info(ctx, collection, id)
}

func (b *base) Delete(ctx context.Context, collection string, id string) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.Delete(ctx, collection, id)
}

func (b *base) List(ctx context.Context, collection string, opts pb.ListOptions) (*pb.DataList, error) {
	if b.next == nil {
		return nil, errors.New("not handler available")
	}
	return b.next.List(ctx, collection, opts)
}

func (b *base) PatchData(ctx context.Context, collection string, id string, content io.Reader, size int64, opts pb.PatchOptions) error {
	if b.next == nil {
		return errors.New("not handler available")
	}
	return b.next.PatchData(ctx, collection, id, content, size, opts)
}

func (b *base) SaveGraft(ctx context.Context, graft *pb.Graft) (string, error) {
	if b.next == nil {
		return "", errors.New("no handler available")
	}
	return b.next.SaveGraft(ctx, graft)
}

func (b *base) GetGraft(ctx context.Context, collection string, dataID string, id string) (*pb.Graft, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.GetGraft(ctx, collection, dataID, id)
}

func (b *base) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*pb.GraftInfo, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.GraftInfo(ctx, collection, dataID, id)
}

func (b *base) ListGrafts(ctx context.Context, collection string, dataID string, opts pb.ListOptions) (*pb.GraftList, error) {
	if b.next == nil {
		return nil, errors.New("no handler available")
	}
	return b.next.ListGrafts(ctx, collection, dataID, opts)
}

func (b *base) DeleteGraft(ctx context.Context, collection string, dataID string, id string) error {
	if b.next == nil {
		return errors.New("no handler available")
	}
	return b.next.DeleteGraft(ctx, collection, dataID, id)
}
