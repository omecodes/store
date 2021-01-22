package objects

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/iancoleman/strcase"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	se "github.com/omecodes/store/search-engine"
)

func NewSqlDB(db *sql.DB, dialect string, tablePrefix string) (DB, error) {
	col, err := bome.NewJSONMap(db, dialect, tablePrefix+"_collections")
	if err != nil {
		return nil, err
	}

	s := &sqlStore{
		db:                db,
		dialect:           dialect,
		collections:       col,
		loadedCollections: &collectionContainer{container: make(map[string]CollectionDB)},
	}
	return s, nil
}

type sqlStore struct {
	loadedCollections *collectionContainer
	db                *sql.DB
	dialect           string
	collections       *bome.JSONMap
}

func (ms *sqlStore) ResolveCollection(ctx context.Context, name string) (CollectionDB, error) {
	col, found := ms.loadedCollections.Get(name)
	if !found {
		encoded, err := ms.collections.Get(name)
		if err != nil {
			return nil, err
		}

		var collection *Collection
		err = json.Unmarshal([]byte(encoded), &collection)
		if err != nil {
			return nil, err
		}

		tableName := strcase.ToSnake(name)
		col, err = NewSQLCollection(collection, ms.db, ms.dialect, tableName)
		ms.loadedCollections.Save(name, col)
	}
	return col, nil
}

func (ms *sqlStore) CreateCollection(ctx context.Context, collection *Collection) error {
	contains, err := ms.collections.Contains(collection.Id)
	if err != nil {
		return err
	}

	if contains {
		return errors.Duplicate
	}

	tableName := strcase.ToSnake(collection.Id)
	col, err := NewSQLCollection(collection, ms.db, ms.dialect, tableName)
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

func (ms *sqlStore) GetCollection(ctx context.Context, id string) (*Collection, error) {
	encoded, err := ms.collections.Get(id)
	if err != nil {
		return nil, err
	}

	var collection *Collection
	err = json.Unmarshal([]byte(encoded), &collection)
	return collection, err
}

func (ms *sqlStore) ListCollections(ctx context.Context) ([]*Collection, error) {
	cursor, err := ms.collections.List()
	if err != nil {
		return nil, err
	}

	defer func() {
		if cer := cursor.Close(); cer != nil {
			log.Error("DB cursor closing", log.Err(err))
		}
	}()

	var collections []*Collection
	for cursor.HasNext() {
		o, err := cursor.Next()
		if err != nil {
			return nil, err
		}

		var collection *Collection
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

func (ms *sqlStore) Save(ctx context.Context, collection string, object *Object, indexes ...*se.TextIndex) error {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return err
	}
	return col.Save(ctx, object, indexes...)
}

func (ms *sqlStore) Patch(ctx context.Context, collection string, patch *Patch) error {
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

func (ms *sqlStore) Get(ctx context.Context, collection string, objectID string, opts GetOptions) (*Object, error) {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return nil, err
	}
	return col.Get(ctx, objectID, opts)
}

func (ms *sqlStore) Info(ctx context.Context, collection string, id string) (*Header, error) {
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

func (ms *sqlStore) Search(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error) {
	col, err := ms.ResolveCollection(ctx, collection)
	if err != nil {
		return nil, err
	}
	return col.Search(ctx, query)
}
