package objects

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/iancoleman/strcase"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/pb"
)

func NewSQLStore(db *sql.DB, dialect string, tableName string) (Objects, error) {
	col, err := bome.NewJSONMap(db, dialect, tableName+"_collections")
	if err != nil {
		return nil, err
	}

	s := &sqlStore{
		db:          db,
		dialect:     dialect,
		collections: col,
	}
	return s, nil
}

type sqlStore struct {
	db          *sql.DB
	dialect     string
	collections *bome.JSONMap
}

func (ms *sqlStore) ResolveCollection(ctx context.Context, name string) (Collection, error) {
	contains, err := ms.collections.Contains(name)
	if err != nil {
		return nil, err
	}

	if !contains {
		return nil, errors.NotFound
	}

	tableName := strcase.ToSnake(name)
	return NewSQLCollection(ms.db, ms.dialect, tableName)
}

func (ms *sqlStore) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	contains, err := ms.collections.Contains(collection.Id)
	if err != nil {
		return err
	}

	if contains {
		return errors.Duplicate
	}

	tableName := strcase.ToSnake(collection.Id)
	_, err = NewSQLCollection(ms.db, ms.dialect, tableName)
	if err != nil {
		return err
	}

	encodedBytes, err := json.Marshal(collection)
	if err != nil {
		return err
	}

	return ms.collections.Save(&bome.MapEntry{
		Key:   collection.Id,
		Value: string(encodedBytes),
	})
}

func (ms *sqlStore) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	encoded, err := ms.collections.Get(id)
	if err != nil {
		return nil, err
	}

	var collection *pb.Collection
	err = json.Unmarshal([]byte(encoded), &collection)
	return collection, err
}

func (ms *sqlStore) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	cursor, err := ms.collections.List()
	if err != nil {
		return nil, err
	}

	defer func() {
		if cer := cursor.Close(); cer != nil {
			log.Error("DB cursor closing", log.Err(err))
		}
	}()

	var collections []*pb.Collection
	for cursor.HasNext() {
		o, err := cursor.Next()
		if err != nil {
			return nil, err
		}

		var collection *pb.Collection
		err = json.Unmarshal([]byte(o.(*bome.MapEntry).Value), &collection)
		if err != nil {
			return nil, err
		}
		collections = append(collections, collection)
	}

	return collections, nil
}

func (ms *sqlStore) DeleteCollection(ctx context.Context, id string) error {
	return ms.collections.Delete(id)
}

func (ms *sqlStore) Save(ctx context.Context, collection string, object *pb.Object, indexes ...*pb.Index) error {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return err
	}
	return col.Save(ctx, object, indexes...)
}

func (ms *sqlStore) Patch(ctx context.Context, collection string, patch *pb.Patch) error {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return err
	}
	return col.Patch(ctx, patch)
}

func (ms *sqlStore) Delete(ctx context.Context, collection string, objectID string) error {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return err
	}
	return col.Delete(ctx, objectID)
}

func (ms *sqlStore) List(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error) {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return nil, err
	}
	return col.List(ctx, opts)
}

func (ms *sqlStore) Get(ctx context.Context, collection string, objectID string, opts pb.GetOptions) (*pb.Object, error) {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return nil, err
	}
	return col.Get(ctx, objectID, opts)
}

func (ms *sqlStore) Info(ctx context.Context, collection string, id string) (*pb.Header, error) {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return nil, err
	}

	return col.Info(ctx, id)
}
