package acl

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/omecodes/bome"
	pb "github.com/omecodes/store/gen/go/proto"
)

type NamespaceConfigStore interface {
	GetNamespace(namespaceId string) (*pb.NamespaceConfig, error)
	GetRelationDefinition(namespaceID string, relationName string) (*pb.RelationDefinition, error)
	SaveNamespace(config *pb.NamespaceConfig) error
	DeleteNamespace(namespaceId string) error
}

func NewNamespaceSQLStore(db *sql.DB, tablePrefix string) (NamespaceConfigStore, error) {
	builder := bome.Build()
	jm, err := builder.SetDialect(bome.MySQL).SetTableName(tablePrefix + "_namespace_configs").SetConn(db).JSONMap()
	if err != nil {
		return nil, err
	}
	return &namespaceSQLStore{db: jm}, nil
}

type namespaceSQLStore struct {
	db *bome.JSONMap
}

func (n *namespaceSQLStore) GetNamespace(namespaceId string) (*pb.NamespaceConfig, error) {
	encoded, err := n.db.Get(namespaceId)
	if err != nil {
		return nil, err
	}

	var config *pb.NamespaceConfig
	err = json.Unmarshal([]byte(encoded), config)
	return config, err
}

func (n *namespaceSQLStore) GetRelationDefinition(namespaceID string, relationName string) (*pb.RelationDefinition, error) {
	relationPath := fmt.Sprintf("$.relations.%s", relationName)
	encoded, err := n.db.ExtractAt(namespaceID, relationPath)
	if err != nil {
		return nil, err
	}

	var relationDefinition *pb.RelationDefinition
	err = json.Unmarshal([]byte(encoded), relationDefinition)
	return relationDefinition, err
}

func (n *namespaceSQLStore) SaveNamespace(config *pb.NamespaceConfig) error {
	encoded, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return n.db.Upsert(&bome.MapEntry{
		Key:   config.Namespace,
		Value: string(encoded),
	})
}

func (n *namespaceSQLStore) DeleteNamespace(namespaceId string) error {
	return n.db.Delete(namespaceId)
}
