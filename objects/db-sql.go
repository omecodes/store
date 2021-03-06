package objects

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/iancoleman/strcase"
	"github.com/omecodes/bome"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	pb "github.com/omecodes/store/gen/go/proto"
)

func NewSqlDB(db *sql.DB, dialect string, tablePrefix string) (DB, error) {
	col, err := bome.Build().
		SetDialect(dialect).
		SetConn(db).
		SetTableName(tablePrefix + "_collections").
		JSONMap()
	if err != nil {
		return nil, err
	}

	s := &sqlStore{
		db:                db,
		dialect:           dialect,
		collections:       col,
		tablePrefix:       tablePrefix,
		loadedCollections: &collectionContainer{container: make(map[string]CollectionDB)},
	}
	return s, nil
}

type sqlStore struct {
	loadedCollections *collectionContainer
	db                *sql.DB
	dialect           string
	tablePrefix       string
	collections       *bome.JSONMap
}

func (ms *sqlStore) ResolveCollection(_ context.Context, name string) (CollectionDB, error) {
	col, found := ms.loadedCollections.Get(name)
	if !found || col == nil {
		encoded, err := ms.collections.Get(name)
		if err != nil {
			return nil, err
		}

		var collection *pb.Collection
		err = json.Unmarshal([]byte(encoded), &collection)
		if err != nil {
			return nil, err
		}

		tableName := strcase.ToSnake(name)
		col, err = NewSQLCollection(collection, ms.db, ms.dialect, ms.tablePrefix+"_"+tableName)
		if err != nil {
			logs.Error("could not create collection", logs.Err(err))
			return nil, errors.Internal("failed to load collection manager")
		}
		ms.loadedCollections.Save(name, col)
	}
	return col, nil
}

func (ms *sqlStore) CreateCollection(_ context.Context, collection *pb.Collection) error {
	contains, err := ms.collections.Contains(collection.Id)
	if err != nil {
		return err
	}

	if contains {
		return errors.Conflict("duplicate collection")
	}

	tableName := strcase.ToSnake(collection.Id)
	col, err := NewSQLCollection(collection, ms.db, ms.dialect, ms.tablePrefix+"_"+tableName)
	if err != nil {
		return err
	}
	ms.loadedCollections.Save(collection.Id, col)

	encodedBytes, err := json.Marshal(collection)
	if err != nil {
		return err
	}

	return ms.collections.Save(&bome.MapEntry{
		Key:   collection.Id,
		Value: string(encodedBytes),
	})
}

func (ms *sqlStore) GetCollection(_ context.Context, id string) (*pb.Collection, error) {
	encoded, err := ms.collections.Get(id)
	if err != nil {
		return nil, err
	}

	var collection *pb.Collection
	err = json.Unmarshal([]byte(encoded), &collection)
	return collection, err
}

func (ms *sqlStore) ListCollections(_ context.Context) ([]*pb.Collection, error) {
	cursor, err := ms.collections.List()
	if err != nil {
		return nil, err
	}

	defer func() {
		if cer := cursor.Close(); cer != nil {
			logs.Error("DB cursor closing", logs.Err(err))
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

func (ms *sqlStore) DeleteCollection(_ context.Context, id string) error {
	return ms.collections.Delete(id)
}

func (ms *sqlStore) Save(ctx context.Context, collection string, object *pb.Object, indexes ...*pb.TextIndex) error {
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

func (ms *sqlStore) Get(ctx context.Context, collection string, objectID string, opts GetObjectOptions) (*pb.Object, error) {
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

func (ms *sqlStore) List(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return nil, err
	}
	return col.List(ctx, opts)
}

func (ms *sqlStore) Search(ctx context.Context, collection string, query *pb.SearchQuery) (*Cursor, error) {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return nil, err
	}
	return col.Search(ctx, query)
}
