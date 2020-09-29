package store

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/omestore/ent"
	entUser "github.com/omecodes/omestore/ent/user"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/omestore/store/internals"
	"io"
	"time"
)

type execHandler struct {
	base
}

func (e *execHandler) RegisterUser(ctx context.Context, user *ent.User, opts pb.UserOptions) error {
	db := getDB(ctx)
	if db == nil {
		log.Info("missing users DB in context")
		return errors.New("wrong context")
	}

	_, err := db.User.Create().
		SetID(user.ID).
		SetCreatedAt(time.Now().Unix()).
		SetValidated(user.Validated).
		SetPassword(user.Password).Save(ctx)

	if err != nil {
		log.Error("failed to save user in db", log.Err(err))
	}
	return err
}

func (e *execHandler) ListUsers(ctx context.Context, opts pb.UserOptions) ([]*ent.User, error) {
	db := getDB(ctx)
	if db == nil {
		log.Info("missing users DB in context")
		return nil, errors.New("wrong context")
	}

	query := db.User.Query()
	if opts.WithAccessList {
		query = query.WithAccesses()
	}

	if opts.WithPermissions {
		query = query.WithPermissions()
	}

	if opts.WithGroups {
		query = query.WithGroup()
	}

	users, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	if !opts.WithPassword {
		for _, u := range users {
			u.Password = ""
		}
	}
	return users, nil
}

func (e *execHandler) CreateUser(ctx context.Context, user *ent.User) error {
	db := getDB(ctx)
	if db == nil {
		log.Info("missing storage in context")
		return errors.Internal
	}

	sh := sha512.New()
	_, err := sh.Write([]byte(user.Password))
	if err != nil {
		log.Error("could not create hashed password", log.Err(err))
		return errors.Internal
	}
	hashed := sh.Sum(nil)
	hexPass := hex.EncodeToString(hashed[:])

	_, err = db.User.Create().SetID(user.ID).
		SetValidated(true).
		SetPassword(hexPass).
		SetEmail(user.Email).
		SetCreatedAt(time.Now().Unix()).Save(ctx)

	if err != nil {
		log.Error("could not create user", log.Err(err))
		if ent.IsConstraintError(err) || ent.IsNotSingular(err) {
			return errors.BadInput
		}
		return err
	}

	return nil
}

func (e *execHandler) ValidateUser(ctx context.Context, username string, opts pb.UserOptions) error {
	db := getDB(ctx)
	if db == nil {
		log.Info("missing storage in context")
		return errors.Internal
	}

	_, err := db.User.UpdateOneID(username).SetValidated(true).Save(ctx)
	return err
}

func (e *execHandler) UserInfo(ctx context.Context, username string, opts pb.UserOptions) (*ent.User, error) {
	db := getDB(ctx)
	if db == nil {
		log.Info("missing users DB in context")
		return nil, errors.New("wrong context")
	}

	query := db.User.Query().Where(entUser.ID(username))

	if opts.WithAccessList {
		query = query.WithAccesses()
	}

	if opts.WithPermissions {
		query = query.WithPermissions()
	}

	if opts.WithGroups {
		query = query.WithGroup()
	}

	u, err := query.First(ctx)
	if err != nil {
		return nil, err
	}

	if !opts.WithPassword {
		u.Password = ""
	}
	return u, nil
}

func (e *execHandler) PatchData(ctx context.Context, collection string, id string, content io.Reader, size int64, opts pb.PatchOptions) error {
	panic("implement me")
}

func (e *execHandler) SetSettings(ctx context.Context, value *JSON, opts pb.SettingsOptions) error {
	appData := getAppdata(ctx)
	if appData == nil {
		log.Info("missing settings database in context")
		return errors.Internal
	}

	settings, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return appData.Set(internals.Settings, string(settings))
}

func (e *execHandler) GetSettings(ctx context.Context, opts pb.SettingsOptions) (*JSON, error) {
	appData := getAppdata(ctx)
	if appData == nil {
		log.Info("missing settings database in context")
		return nil, errors.Internal
	}

	value, err := appData.Get(internals.Settings, opts.Path)
	if err != nil {
		return nil, err
	}

	var o interface{}
	err = json.Unmarshal([]byte(value), &o)
	return &JSON{Object: o}, err
}

func (e *execHandler) GetCollections(ctx context.Context) ([]string, error) {
	storage := getStorage(ctx)
	if storage == nil {
		log.Info("missing storage in context")
		return nil, errors.Internal
	}
	return storage.Collections(ctx)
}

func (e *execHandler) PutData(ctx context.Context, data *pb.Data, opts pb.PutDataOptions) error {
	storage := getStorage(ctx)
	if storage == nil {
		log.Info("missing storage in context")
		return errors.Internal
	}

	db := getDB(ctx)
	if db == nil {
		log.Error("could not get db from context")
		return errors.Internal
	}

	cred := ome.CredentialsFromContext(ctx)
	if cred != nil {
		data.CreatedBy = cred.Username
	}
	data.CreatedAt = time.Now().Unix()

	return storage.Save(ctx, data)
}

func (e *execHandler) GetData(ctx context.Context, collection string, id string, opts pb.GetDataOptions) (*pb.Data, error) {
	storage := getStorage(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.New("wrong context")
	}
	return storage.Get(ctx, collection, id, pb.DataOptions{Path: opts.Path})
}

func (e *execHandler) Info(ctx context.Context, collection string, id string) (*pb.Info, error) {
	storage := getStorage(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.New("wrong context")
	}
	return storage.Info(ctx, collection, id)
}

func (e *execHandler) Delete(ctx context.Context, collection string, id string) error {
	db := getStorage(ctx)
	if db == nil {
		log.Info("missing DB in context")
		return errors.New("wrong context")
	}
	return db.Delete(ctx, &pb.Data{
		Collection: collection,
		ID:         id,
	})
}

func (e *execHandler) List(ctx context.Context, collection string, opts pb.ListOptions) (*pb.DataList, error) {
	db := getStorage(ctx)
	if db == nil {
		log.Info("missing DB in context")
		return nil, errors.New("wrong context")
	}
	return db.List(ctx, collection, opts)
}

func (e *execHandler) SaveGraft(ctx context.Context, graft *pb.Graft) (string, error) {
	storage := getStorage(ctx)
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

func (e *execHandler) GetGraft(ctx context.Context, collection string, dataID string, id string) (*pb.Graft, error) {
	storage := getStorage(ctx)
	if storage == nil {
		log.Error("could not get storage from context")
		return nil, errors.Internal
	}
	return storage.GetGraft(ctx, collection, dataID, id)
}

func (e *execHandler) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*pb.GraftInfo, error) {
	storage := getStorage(ctx)
	if storage == nil {
		log.Error("could not get storage from context")
		return nil, errors.Internal
	}
	return storage.GraftInfo(ctx, collection, dataID, id)
}

func (e *execHandler) ListGrafts(ctx context.Context, collection string, dataID string, opts pb.ListOptions) (*pb.GraftList, error) {
	storage := getStorage(ctx)
	if storage == nil {
		log.Error("could not get storage from context")
		return nil, errors.Internal
	}
	return storage.GetAllGraft(ctx, collection, dataID, opts)
}

func (e *execHandler) DeleteGraft(ctx context.Context, collection string, dataID string, id string) error {
	storage := getStorage(ctx)
	if storage == nil {
		log.Error("could not get storage from context")
		return errors.Internal
	}
	return storage.DeleteGraft(ctx, collection, dataID, id)
}
