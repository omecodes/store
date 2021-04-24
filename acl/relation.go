package acl

import (
	"database/sql"
	"github.com/omecodes/bome"
	pb "github.com/omecodes/store/gen/go/proto"
)

const (

	relationTableSchema = `
create table if not exists $prefix$_relations (
    sid int not null,
	subject varchar(255) not null,
	relation varchar(255) not null,
	object varchar(255) not null,
	commit_time long not null,
	primary key (subject, relation, object)
)$engine$;
`

	relationScanner = "relation_scanner_key"
)

type RelationStore interface {
	Save(relation *pb.Relation) error
	Exists(relation *pb.Relation) (bool, error)
	GetSubjects(info *pb.RelationSubjectInfo) ([]string, error)
	Delete(relation *pb.Relation) error
	DeleteAllSubjectRelation(subject, name string) error
	DeleteAllForSubject(subject string) error
}

func NewRelationSQLStore(db *sql.DB, tablePrefix string) (RelationStore, error) {
	bm, err := bome.New(db)
	if err != nil {
		return nil, err
	}
	bm.SetTablePrefix(tablePrefix)
	bm.AddTableDefinition(relationTableSchema)
	bm.RegisterScanner(relationScanner, nil)

	return nil, nil
}

type relationSQLStore struct {
	db *bome.DB
}

func (r *relationSQLStore) Save(relation *pb.Relation) error {
	panic("implement me")
}

func (r *relationSQLStore) Exists(relation *pb.Relation) (bool, error) {
	panic("implement me")
}

func (r *relationSQLStore) GetSubjects(info *pb.RelationSubjectInfo) ([]string, error) {
	panic("implement me")
}

func (r *relationSQLStore) Delete(relation *pb.Relation) error {
	panic("implement me")
}

func (r *relationSQLStore) DeleteAllSubjectRelation(subject, name string) error {
	panic("implement me")
}

func (r *relationSQLStore) DeleteAllForSubject(subject string) error {
	panic("implement me")
}
