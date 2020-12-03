package oms

import (
	"database/sql"
	"encoding/json"
	"github.com/omecodes/bome"
	"github.com/omecodes/omestore/pb"
)

type AccessStore interface {
	SaveRules(objectID string, rules *pb.PathAccessRules) error
	GetRules(objectID string) (*pb.PathAccessRules, error)
	GetForPath(objectID string, path string) (*pb.AccessRules, error)
	Delete(objectID string) error
}

func NewSQLAccessStore(db *sql.DB, dialect string, tableName string) (AccessStore, error) {
	store, err := bome.NewJSONMap(db, dialect, tableName)
	return &sqlPermStore{
		store: store,
	}, err
}

type sqlPermStore struct {
	store *bome.JSONMap
}

func (p *sqlPermStore) SaveRules(objectID string, rules *pb.PathAccessRules) error {
	rulesBytes, _ := json.Marshal(rules.AccessRules)
	entry := &bome.MapEntry{
		Key:   objectID,
		Value: string(rulesBytes),
	}
	return p.store.Save(entry)
}

func (p *sqlPermStore) GetRules(objectID string) (*pb.PathAccessRules, error) {
	value, err := p.store.Get(objectID)
	if err != nil {
		return nil, err
	}
	pr := &pb.PathAccessRules{AccessRules: map[string]*pb.AccessRules{}}
	err = json.Unmarshal([]byte(value), &pr.AccessRules)
	return pr, err
}

func (p *sqlPermStore) GetForPath(objectID string, path string) (*pb.AccessRules, error) {
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

func (p *sqlPermStore) Delete(objectID string) error {
	return p.store.Delete(objectID)
}
