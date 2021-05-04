package acl

import (
	"context"
	"database/sql"
	"github.com/omecodes/bome"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	pb "github.com/omecodes/store/gen/go/proto"
)

const (
	relationTableSchema = `
create table if not exists $prefix$_tuples (
    sid varchar(255) not null,
	object varchar(255) not null,
	relation varchar(255) not null,
	subject varchar(255) not null,
	commit_time long not null,
	primary key (object, relation, subject)
)$engine$;
`
	queryInsertTuple   = `insert into $prefix$_tuples values (?, ?, ?, ?, ?);`
	queryTupleExists   = `select 1 from $prefix$_tuples where object=? and relation=? and subject=? and commit_time>=?;`
	queryTupleSubjects = `select subject from $prefix$_tuples where relation=? and object=? and commit_time>=?;`
	queryTupleObjects  = `select object from $prefix$_tuples where relation=? and subject=? and commit_time>=?;`
	queryDeleteTuples  = `delete from $prefix$_tuples where subject=? and relation=? and object=? and commit_time>=?;`
	tupleScanner       = "relation_scanner_key"
)

type TupleStore interface {
	Save(ctx context.Context, a *pb.DBEntry) error
	Check(ctx context.Context, entry *pb.DBEntry) (bool, error)
	GetSubjects(ctx context.Context, info *pb.DBSubjectSetInfo) ([]string, error)
	GetObjects(ctx context.Context, info *pb.DBObjectSetInfo) ([]string, error)
	Delete(ctx context.Context, entry *pb.DBEntry) error
}

func NewTupleSQLStore(db *sql.DB, dialect string, tablePrefix string) (TupleStore, error) {
	var (
		bm  *bome.DB
		err error
	)

	if dialect == bome.MySQL {
		bm, err = bome.New(db)
	} else {
		bm, err = bome.NewLite(db)
	}

	if err != nil {
		return nil, err
	}

	bm.SetTablePrefix(tablePrefix)
	bm.AddTableDefinition(relationTableSchema)
	bm.RegisterScanner(tupleScanner, bome.NewScannerFunc(func(row bome.Row) (interface{}, error) {
		entry := new(pb.DBEntry)
		err = row.Scan(&entry.Sid, &entry.Object, &entry.Relation, &entry.Subject, &entry.StateMinAge)
		return entry, err
	}))
	store := &relationSQLStore{
		db: bm,
	}
	err = bm.Init()
	return store, err
}

type relationSQLStore struct {
	db *bome.DB
}

func (r *relationSQLStore) Save(_ context.Context, entry *pb.DBEntry) error {
	return r.db.Exec(queryInsertTuple, entry.Sid, entry.Object, entry.Relation, entry.Subject, entry.StateMinAge).Error
}

func (r *relationSQLStore) Check(_ context.Context, entry *pb.DBEntry) (bool, error) {
	o, err := r.db.QueryFirst(queryTupleExists, bome.IntScanner, entry.Object, entry.Relation, entry.Subject, entry.StateMinAge)
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}
	return o != nil && o.(int64) == 1, nil
}

func (r *relationSQLStore) GetSubjects(_ context.Context, info *pb.DBSubjectSetInfo) ([]string, error) {
	c, err := r.db.Query(queryTupleSubjects, bome.StringScanner, info.Relation, info.Object, info.StateMinAge)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cErr := c.Close(); cErr != nil {
			logs.Error("get subjects cursor closing", logs.Details("info", info), logs.Err(err))
		}
	}()

	var (
		o        interface{}
		subjects []string
	)

	for c.HasNext() {
		o, err = c.Next()
		if err != nil {
			break
		}
		subjects = append(subjects, o.(string))
	}
	return subjects, err
}

func (r *relationSQLStore) GetObjects(_ context.Context, info *pb.DBObjectSetInfo) ([]string, error) {
	c, err := r.db.Query(queryTupleObjects, bome.StringScanner, info.Relation, info.Subject, info.StateMinAge)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cErr := c.Close(); cErr != nil {
			logs.Error("get subjects cursor closing", logs.Details("info", info), logs.Err(err))
		}
	}()

	var (
		o        interface{}
		subjects []string
	)

	for c.HasNext() {
		o, err = c.Next()
		if err != nil {
			break
		}
		subjects = append(subjects, o.(string))
	}
	return subjects, err
}

func (r *relationSQLStore) Delete(_ context.Context, entry *pb.DBEntry) error {
	return r.db.Exec(queryDeleteTuples, entry.Subject, entry.Relation, entry.Object, entry.StateMinAge).Error
}
