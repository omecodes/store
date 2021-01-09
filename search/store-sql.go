package se

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/pb"
	"strings"
)

const wordsTableName = "$prefix$_words_mapping"
const numbersTableName = "$prefix$_numbers_mapping"

const wordsTablesDef = `
create table if not exists $prefix$_words_mapping (
  	token varchar(255) not null,
    field varchar(255) not null,
    objects LONGTEXT not null,
    primary key(token, field)
);
`

const numbersTablesDef = `
create table if not exists $prefix$_numbers_mapping (
  	num bigInt not null,
    field varchar(255) not null,
    objects LONGTEXT not null,
    primary key(num, field)
);
`

const insertWord = `
insert into $prefix$_words_mapping values(?, ?, ?);
`

const appendToWord = `
update $prefix$_words_mapping set objects=concat(objects, ?) where token=? and field=?;
`

const insertNumber = `
insert into $prefix$_numbers_mapping values(?, ?, ?);
`

const appendToNumber = `
update $prefix$_numbers_mapping set objects=concat(objects, ?) where num=? and field=?;
`

func NewSQLIndexStore(db *sql.DB, dialect string, tablePrefix string) (Store, error) {
	s := new(sqlStore)
	var err error

	if dialect == bome.SQLite3 {
		s.db, err = bome.NewLite(db)

	} else if dialect == bome.MySQL {
		s.db, err = bome.New(db)

	} else {
		return nil, errors.New("not supported")
	}

	if err != nil {
		return nil, err
	}

	s.db.SetTablePrefix(tablePrefix)

	err = s.db.RawExec(wordsTablesDef).Error
	if err == nil {
		err = s.db.RawExec(numbersTablesDef).Error
	}
	return s, err
}

type sqlStore struct {
	db *bome.Bome
}

func (s *sqlStore) SaveWordMapping(word string, field string, id string) error {
	err := s.db.RawExec(insertWord, word, field, id).Error
	if err != nil {
		if bome.IsPrimaryKeyConstraintError(err) {
			err = s.db.RawExec(appendToWord, " "+id, word, field).Error
			if err != nil {
				log.Error("failed to create index mapping", log.Err(err))
			}
		}
	}
	return err
}

func (s *sqlStore) SaveNumberMapping(num int64, field string, id string) error {
	err := s.db.RawExec(insertWord, num, field, id).Error
	if err != nil && bome.IsPrimaryKeyConstraintError(err) {
		err = s.db.RawExec(appendToWord, " "+id, num, field).Error
	}
	return err
}

func (s *sqlStore) Search(expression *pb.BooleanExp) (Cursor, error) {
	query := evaluate(expression)
	_, err := s.db.RawQuery(query, bome.StringScanner)
	if err != nil {
		return nil, err
	}
	return nil, err
}

func evaluate(expr *pb.BooleanExp) string {

	switch v := expr.Expression.(type) {
	case *pb.BooleanExp_Or:
		var evaluatedExpression []string
		for _, ox := range v.Or.Expressions {
			evaluatedExpression = append(evaluatedExpression, evaluate(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " OR "))

	case *pb.BooleanExp_Contains:
		return "(word like '%" + v.Contains.Value + "%')"

	case *pb.BooleanExp_StartsWith:
		return "(word like '" + v.StartsWith.Value + "%')"

	case *pb.BooleanExp_EndsWith:
		return "(word like '%" + v.EndsWith.Value + "')"

	case *pb.BooleanExp_StrEqual:
		return "(word='" + v.StrEqual.Value + "')"

	case *pb.BooleanExp_Lt:
		return fmt.Sprintf("(number<%d)", v.Lt.Value)

	case *pb.BooleanExp_Lte:
		return fmt.Sprintf("(number<=%d)", v.Lte.Value)

	case *pb.BooleanExp_Gt:
		return fmt.Sprintf("(number>%d)", v.Gt.Value)

	case *pb.BooleanExp_Gte:
		return fmt.Sprintf("(number>=%d)", v.Gte.Value)

	case *pb.BooleanExp_NumbEq:
		return fmt.Sprintf("(number=%d)", v.NumbEq.Value)
	}

	return ""
}
