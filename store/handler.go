package store

import (
	"context"
	"github.com/omecodes/omestore/ent"
	"github.com/omecodes/omestore/pb"
	"io"
)

type Handler interface {
	SetSettings(ctx context.Context, value *JSON, opts pb.SettingsOptions) error
	GetSettings(ctx context.Context, opts pb.SettingsOptions) (*JSON, error)

	RegisterUser(ctx context.Context, user *ent.User, opts pb.UserOptions) error
	ListUsers(ctx context.Context, opts pb.UserOptions) ([]*ent.User, error)
	CreateUser(ctx context.Context, user *ent.User) error
	ValidateUser(ctx context.Context, username string, opts pb.UserOptions) error
	UserInfo(ctx context.Context, username string, opts pb.UserOptions) (*ent.User, error)

	GetCollections(ctx context.Context) ([]string, error)

	PutData(ctx context.Context, data *pb.Data, opts pb.PutDataOptions) error
	PatchData(ctx context.Context, collection string, id string, content io.Reader, size int64, opts pb.PatchOptions) error
	GetData(ctx context.Context, collection string, id string, opts pb.GetDataOptions) (*pb.Data, error)
	Info(ctx context.Context, collection string, id string) (*pb.Info, error)
	Delete(ctx context.Context, collection string, id string) error
	List(ctx context.Context, collection string, opts pb.ListOptions) (*pb.DataList, error)

	SaveGraft(ctx context.Context, graft *pb.Graft) (string, error)
	GetGraft(ctx context.Context, collection string, dataID string, id string) (*pb.Graft, error)
	GraftInfo(ctx context.Context, collection string, dataID string, id string) (*pb.GraftInfo, error)
	ListGrafts(ctx context.Context, collection string, dataID string, opts pb.ListOptions) (*pb.GraftList, error)
	DeleteGraft(ctx context.Context, collection string, dataID string, id string) error
}
