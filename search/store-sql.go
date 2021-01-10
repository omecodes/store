package se

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/pb"
	"io"
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
const deleteObjectWordMapping = `
update $prefix$_words_mapping set objects=replace(objects, ?, ' ') where objects like ?;
`

const insertNumber = `
insert into $prefix$_numbers_mapping values(?, ?, ?);
`

const appendToNumber = `
update $prefix$_numbers_mapping set objects=concat(objects, ?) where num=? and field=?;
`

const deleteObjectNumberMapping = `
update $prefix$_numbers_mapping set objects=replace(objects, ?, ' ') where objects like ?;
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

	err = s.db.Init()
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
	err := s.db.RawExec(insertWord, word, field, " "+id).Error
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
	err := s.db.RawExec(insertNumber, num, field, " "+id).Error
	if err != nil && bome.IsPrimaryKeyConstraintError(err) {
		err = s.db.RawExec(appendToNumber, " "+id, num, field).Error
	}
	return err
}

func (s *sqlStore) DeleteObjectMappings(id string) error {
	err := s.db.RawExec(deleteObjectNumberMapping, id, "' %"+id+"%'").Error
	if err == nil {
		err = s.db.RawExec(deleteObjectWordMapping, id, "' %"+id+"%'").Error
	}
	return err
}

func (s *sqlStore) Search(expression *pb.BooleanExp) (Cursor, error) {
	affectedTables := &tables{}
	var tableNames []string

	whereClause := evaluate(expression, affectedTables)

	if affectedTables.numberMapping {
		tableNames = append(tableNames, numbersTableName)
	}

	if affectedTables.textMapping {
		tableNames = append(tableNames, wordsTableName)
	}

	if len(tableNames) == 0 {
		return nil, errors.New("bad input")
	}

	var query string
	if len(tableNames) > 1 {
		var selection []string
		for _, tableName := range tableNames {
			selection = append(selection, tableName+".objects")
		}
		query += fmt.Sprintf("select concat(%s) from %s where %s)", strings.Join(selection, "'<>'"), strings.Join(tableNames, ","), whereClause)
	} else {
		query += fmt.Sprintf("select objects from %s where %s", tableNames[0], whereClause)
	}

	cursor, err := s.db.RawQuery(query, bome.StringScanner)
	if err != nil {
		return nil, err
	}

	return &sqlSearchCursor{
		cursor: cursor,
	}, err
}

type tables struct {
	numberMapping bool
	textMapping   bool
}

func evaluate(expr *pb.BooleanExp, tables *tables) string {
	textAnalyzer := getQueryTextAnalyzer()

	switch v := expr.Expression.(type) {
	case *pb.BooleanExp_Or:
		var evaluatedExpression []string
		for _, ox := range v.Or.Expressions {
			evaluatedExpression = append(evaluatedExpression, evaluate(ox, tables))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " OR "))

	case *pb.BooleanExp_Contains:
		tables.textMapping = true
		return "($prefix$_words_mapping.token like '%" + textAnalyzer(v.Contains.Value) + "%')"

	case *pb.BooleanExp_StartsWith:
		tables.textMapping = true
		return "($prefix$_words_mapping.token like '" + textAnalyzer(v.StartsWith.Value) + "%')"

	case *pb.BooleanExp_EndsWith:
		tables.textMapping = true
		return "($prefix$_words_mapping.token like '%" + textAnalyzer(v.EndsWith.Value) + "')"

	case *pb.BooleanExp_StrEqual:
		tables.textMapping = true
		return "($prefix$_words_mapping.token='" + textAnalyzer(v.StrEqual.Value) + "')"

	case *pb.BooleanExp_Lt:
		tables.numberMapping = true
		return fmt.Sprintf("($prefix$_numbers_mappingnumber<%d)", v.Lt.Value)

	case *pb.BooleanExp_Lte:
		tables.numberMapping = true
		return fmt.Sprintf("($prefix$_numbers_mappingnumber<=%d)", v.Lte.Value)

	case *pb.BooleanExp_Gt:
		tables.numberMapping = true
		return fmt.Sprintf("($prefix$_numbers_mappingnumber>%d)", v.Gt.Value)

	case *pb.BooleanExp_Gte:
		tables.numberMapping = true
		return fmt.Sprintf("($prefix$_numbers_mappingnumber>=%d)", v.Gte.Value)

	case *pb.BooleanExp_NumbEq:
		tables.numberMapping = true
		return fmt.Sprintf("($prefix$_numbers_mappingnumber=%d)", v.NumbEq.Value)
	}

	return ""
}

type sqlSearchCursor struct {
	cursor      bome.Cursor
	currentList []string
}

func (s *sqlSearchCursor) Next() (string, error) {
	for {
		if len(s.currentList) == 0 {
			if !s.cursor.HasNext() {
				return "", io.EOF
			}

			o, err := s.cursor.Next()
			if err != nil {
				return "", err
			}

			s.currentList = strings.Split(o.(string), "<>")
		}

		next := strings.Trim(s.currentList[0], " ")
		if next == "" {
			continue
		}

		s.currentList = s.currentList[1:]
		return next, nil
	}
}

func (s *sqlSearchCursor) Close() error {
	return nil
}
