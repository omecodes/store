package objects

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/omecodes/bome"
)

func NewSQLACLStore(db *sql.DB, dialect string, tableName string) (ACLManager, error) {
	store, err := bome.NewJSONDoubleMap(db, dialect, tableName)
	return &sqlPermStore{
		store: store,
	}, err
}

type sqlPermStore struct {
	store *bome.JSONDoubleMap
}

func (p *sqlPermStore) SaveRules(ctx context.Context, collection string, objectID string, rules *PathAccessRules) error {
	rulesBytes, _ := json.Marshal(rules.AccessRules)
	entry := &bome.DoubleMapEntry{
		FirstKey:  collection,
		SecondKey: objectID,
		Value:     string(rulesBytes),
	}
	return p.store.Upsert(entry)
}

func (p *sqlPermStore) GetRules(ctx context.Context, collection string, objectID string) (*PathAccessRules, error) {
	value, err := p.store.Get(collection, objectID)
	if err != nil {
		return nil, err
	}
	pr := &PathAccessRules{AccessRules: map[string]*AccessRules{}}
	err = json.Unmarshal([]byte(value), &pr.AccessRules)
	return pr, err
}

func (p *sqlPermStore) GetForPath(ctx context.Context, collection string, objectID string, path string) (*AccessRules, error) {
	readRules, err := p.store.ExtractAt(collection, objectID, path+".read")
	if err != nil {
		return nil, err
	}

	writeRules, err := p.store.ExtractAt(collection, objectID, path+".write")
	if err != nil {
		return nil, err
	}

	ar := new(AccessRules)
	err = json.Unmarshal([]byte(readRules), &ar.Read)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(writeRules), &ar.Write)
	if err != nil {
		return nil, err
	}

	return ar, nil
}

func (p *sqlPermStore) Delete(ctx context.Context, collection string, objectID string) error {
	return p.store.Delete(collection, objectID)
}

func (p *sqlPermStore) DeleteForPath(ctx context.Context, collection string, objectID string, path string) error {
	rules, err := p.GetRules(ctx, collection, objectID)
	if err != nil {
		return err
	}

	delete(rules.AccessRules, path)
	return p.SaveRules(ctx, collection, objectID, rules)
}
