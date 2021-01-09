package search

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/omecodes/bome"
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
insert into words_mapping values(?, ?, ?);
`

const appendToWord = `
update words_mapping set field=?, objects=? where token=?;
`

const insertNumber = `
insert into numbers_mapping values(?, ?, ?);
`

func NewSQLStore(db *sql.DB, dialect string, tablePrefix string) (Store, error) {
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

	_, err = db.Exec(wordsTablesDef)
	if err == nil {
		_, err = db.Exec(numbersTablesDef)
	}
	return s, err
}

type sqlStore struct {
	db *bome.Bome
}

func (s *sqlStore) SaveWordMapping(word string, field string, ids ...string) error {
	value := strings.Join(ids, " ")
	err := s.db.RawExec(insertWord, word, field, value).Error
	if err != nil && bome.IsPrimaryKeyConstraintError(err) {
		err = s.db.RawExec(appendToWord, field, value, word).Error
	}
	return err
}

func (s *sqlStore) SaveNumberMapping(num int64, field string, ids ...string) error {
	value := strings.Join(ids, " ")
	err := s.db.RawExec(insertWord, num, field, value).Error
	if err != nil && bome.IsPrimaryKeyConstraintError(err) {
		err = s.db.RawExec(appendToWord, field, value, num).Error
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
