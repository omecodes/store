package acl

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/omecodes/bome"
	"github.com/omecodes/store/pb"
)

func NewSQLStore(db *sql.DB, dialect string, tableName string) (Store, error) {
	store, err := bome.NewJSONMap(db, dialect, tableName)
	return &sqlPermStore{
		store: store,
	}, err
}

type sqlPermStore struct {
	store *bome.JSONMap
}

func (p *sqlPermStore) SaveRules(ctx context.Context, objectID string, rules *pb.PathAccessRules) error {
	rulesBytes, _ := json.Marshal(rules.AccessRules)
	entry := &bome.MapEntry{
		Key:   objectID,
		Value: string(rulesBytes),
	}
	return p.store.Upsert(entry)
}

func (p *sqlPermStore) GetRules(ctx context.Context, objectID string) (*pb.PathAccessRules, error) {
	value, err := p.store.Get(objectID)
	if err != nil {
		return nil, err
	}
	pr := &pb.PathAccessRules{AccessRules: map[string]*pb.AccessRules{}}
	err = json.Unmarshal([]byte(value), &pr.AccessRules)
	return pr, err
}

func (p *sqlPermStore) GetForPath(ctx context.Context, objectID string, path string) (*pb.AccessRules, error) {
	readRules, err := p.store.ExtractAt(objectID, path+".read")
	if err != nil {
		return nil, err
	}

	writeRules, err := p.store.ExtractAt(objectID, path+".write")
	if err != nil {
		return nil, err
	}

	ar := new(pb.AccessRules)
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

func (p *sqlPermStore) Delete(ctx context.Context, objectID string) error {
	return p.store.Delete(objectID)
}

func (p *sqlPermStore) DeleteForPath(ctx context.Context, objectID string, path string) error {
	rules, err := p.GetRules(ctx, objectID)
	if err != nil {
		return err
	}

	delete(rules.AccessRules, path)
	return p.SaveRules(ctx, objectID, rules)
}
