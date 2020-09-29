package store

import (
	"context"
	"github.com/omecodes/omestore/ent"
	"github.com/omecodes/omestore/pb"
	"io"
)

type doNothing struct{}

func (d *doNothing) SetSettings(ctx context.Context, value *JSON, opts pb.SettingsOptions) error {
	return nil
}

func (d *doNothing) GetSettings(ctx context.Context, opts pb.SettingsOptions) (*JSON, error) {
	return nil, nil
}

func (d *doNothing) RegisterUser(ctx context.Context, user *ent.User, opts pb.UserOptions) error {
	return nil
}

func (d *doNothing) ListUsers(ctx context.Context, opts pb.UserOptions) ([]*ent.User, error) {
	return nil, nil
}

func (d *doNothing) CreateUser(ctx context.Context, user *ent.User) error {
	return nil
}

func (d *doNothing) ValidateUser(ctx context.Context, username string, opts pb.UserOptions) error {
	return nil
}

func (d *doNothing) UserInfo(ctx context.Context, username string, opts pb.UserOptions) (*ent.User, error) {
	return nil, nil
}

func (d *doNothing) GetCollections(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (d *doNothing) PutData(ctx context.Context, data *pb.Data, opts pb.PutDataOptions) error {
	return nil
}

func (d *doNothing) PatchData(ctx context.Context, collection string, id string, content io.Reader, size int64, opts pb.PatchOptions) error {
	return nil
}

func (d *doNothing) GetData(ctx context.Context, collection string, id string, opts pb.GetDataOptions) (*pb.Data, error) {
	return nil, nil
}

func (d *doNothing) Info(ctx context.Context, collection string, id string) (*pb.Info, error) {
	return nil, nil
}

func (d *doNothing) Delete(ctx context.Context, collection string, id string) error {
	return nil
}

func (d *doNothing) List(ctx context.Context, collection string, opts pb.ListOptions) (*pb.DataList, error) {
	return nil, nil
}

func (d *doNothing) SaveGraft(ctx context.Context, graft *pb.Graft) (string, error) {
	return "", nil
}

func (d *doNothing) GetGraft(ctx context.Context, collection string, dataID string, id string) (*pb.Graft, error) {
	return nil, nil
}

func (d *doNothing) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*pb.GraftInfo, error) {
	return nil, nil
}

func (d *doNothing) GraftBulk(ctx context.Context, collection string, dataID string, ids ...string) (*pb.GraftList, error) {
	return nil, nil
}

func (d *doNothing) ListGrafts(ctx context.Context, collection string, dataID string, opts pb.ListOptions) (*pb.GraftList, error) {
	return nil, nil
}

func (d *doNothing) DeleteGraft(ctx context.Context, collection string, dataID string, id string) error {
	return nil
}
