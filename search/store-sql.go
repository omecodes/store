package se

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/omecodes/bome"
	"github.com/omecodes/store/pb"
	"io"
	"strings"
)

const wordsTableName = "$prefix$_words_mapping"

const numbersTableName = "$prefix$_numbers_mapping"

const propsTableName = "$prefix$_props_mapping"

const propsTablesDef = `
create table if not exists $prefix$_props_mapping (
  	object varchar(255) not null,
    value JSON not null,
    primary key(object)
);
`

const wordsTablesDef = `
create table if not exists $prefix$_words_mapping (
  	token varchar(255) not null,
    id varchar(255) not null,
    primary key(token, id)
);
`

const numbersTablesDef = `
create table if not exists $prefix$_numbers_mapping (
  	num bigInt not null,
    id varchar(255) not null,
    primary key(num, id)
);
`

const insertWord = `
insert into $prefix$_words_mapping values(?, ?);
`

const deleteObjectWordMappings = `
delete $prefix$_words_mapping where id=?;
`

const insertNumber = `
insert into $prefix$_numbers_mapping values(?, ?);
`

const deleteObjectNumberMapping = `
delete from $prefix$_numbers_mapping where id=?;
`

const insertProps = `
insert into $prefix$_props_mapping values(?, ?);
`

const deleteProps = `
delete from $prefix$_props_mapping where id=?;
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
		if err == nil {
			err = s.db.RawExec(propsTablesDef).Error
		}
	}
	return s, err
}

type sqlStore struct {
	db *bome.Bome
}

func (s *sqlStore) SaveWordMapping(word string, id string) error {
	err := s.db.RawExec(insertWord, word, id).Error
	if bome.IsPrimaryKeyConstraintError(err) {
		return nil
	}
	return err
}

func (s *sqlStore) SaveNumberMapping(num int64, id string) error {
	err := s.db.RawExec(insertNumber, num, id).Error
	if bome.IsPrimaryKeyConstraintError(err) {
		return nil
	}
	return err
}

func (s *sqlStore) SavePropertiesMapping(id string, value string) error {
	err := s.db.RawExec(insertProps, id, value).Error
	if bome.IsPrimaryKeyConstraintError(err) {
		return nil
	}
	return err
}

func (s *sqlStore) DeleteObjectMappings(id string) error {
	err := s.db.RawExec(deleteObjectNumberMapping, id).Error
	if err == nil {
		err = s.db.RawExec(deleteObjectNumberMapping, id).Error
		if err == nil {
			s.db.RawExec(deleteProps, id)
		}
	}
	return err
}

func (s *sqlStore) Search(query *pb.SearchQuery) (Cursor, error) {
	return s.performSearch(query)
}

func (s *sqlStore) performSearch(query *pb.SearchQuery) (Cursor, error) {
	switch q := query.Query.(type) {

	case *pb.SearchQuery_Text:
		sqlQuery := "select value from " + wordsTableName + " where " + evaluateWordSearchingQuery(q.Text)
		c, err := s.db.RawQuery(sqlQuery, bome.StringScanner)
		return &aggregatedStrIdsCursor{cursor: c}, err

	case *pb.SearchQuery_Number:
		sqlQuery := "select value from " + numbersTableName + " where " + evaluateNumberSearchingQuery(q.Number)
		c, err := s.db.RawQuery(sqlQuery, bome.StringScanner)
		return &aggregatedStrIdsCursor{cursor: c}, err

	case *pb.SearchQuery_Fields:
		sqlQuery := "select id from " + propsTableName + " where " + evaluatePropertiesSearchingQuery(q.Fields)
		c, err := s.db.RawQuery(sqlQuery, bome.StringScanner)
		return &bomeCursorWrapper{cursor: c}, err
	}

	return nil, errors.New("unsupported query type")
}

func evaluateWordSearchingQuery(query *pb.StrQuery) string {
	textAnalyzer := getQueryTextAnalyzer()

	switch v := query.Bool.(type) {

	case *pb.StrQuery_And:
		var evaluatedExpression []string
		for _, ox := range v.And.Queries {
			evaluatedExpression = append(evaluatedExpression, evaluateWordSearchingQuery(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " AND "))

	case *pb.StrQuery_Or:
		var evaluatedExpression []string
		for _, ox := range v.Or.Queries {
			evaluatedExpression = append(evaluatedExpression, evaluateWordSearchingQuery(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " OR "))

	case *pb.StrQuery_Contains:
		return "(token like '%" + textAnalyzer(v.Contains.Value) + "%')"

	case *pb.StrQuery_StartsWith:
		return "(token like '" + textAnalyzer(v.StartsWith.Value) + "%')"

	case *pb.StrQuery_EndsWith:
		return "(token='%" + textAnalyzer(v.EndsWith.Value) + "')"

	case *pb.StrQuery_Eq:
		return "(token='" + textAnalyzer(v.Eq.Value) + "')"
	}
	return ""
}

func evaluateNumberSearchingQuery(query *pb.NumQuery) string {
	switch v := query.Bool.(type) {

	case *pb.NumQuery_And:
		var evaluatedExpression []string
		for _, ox := range v.And.Queries {
			evaluatedExpression = append(evaluatedExpression, evaluateNumberSearchingQuery(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " AND "))

	case *pb.NumQuery_Or:
		var evaluatedExpression []string
		for _, ox := range v.Or.Queries {
			evaluatedExpression = append(evaluatedExpression, evaluateNumberSearchingQuery(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " OR "))

	case *pb.NumQuery_Eq:
		return fmt.Sprintf("(num=%d)", v.Eq.Value)

	case *pb.NumQuery_Gt:
		return fmt.Sprintf("(num>%d)", v.Gt.Value)

	case *pb.NumQuery_Gte:
		return fmt.Sprintf("(num>=%d)", v.Gte.Value)

	case *pb.NumQuery_Lt:
		return fmt.Sprintf("(num<%d)", v.Lt.Value)

	case *pb.NumQuery_Lte:
		return fmt.Sprintf("(num<=%d)", v.Lte.Value)
	}

	return ""
}

func evaluatePropertiesSearchingQuery(query *pb.FieldQuery) string {
	textAnalyzer := propsMappingTextAnalyzer()

	switch v := query.Bool.(type) {

	case *pb.FieldQuery_And:
		var evaluatedExpression []string
		for _, ox := range v.And.Queries {
			evaluatedExpression = append(evaluatedExpression, evaluatePropertiesSearchingQuery(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " AND "))

	case *pb.FieldQuery_Or:
		var evaluatedExpression []string
		for _, ox := range v.Or.Queries {
			evaluatedExpression = append(evaluatedExpression, evaluatePropertiesSearchingQuery(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " OR "))

	case *pb.FieldQuery_Contains:
		return fmt.Sprintf("(value->>'$.%s' like '%%%s%%')", v.Contains.Field, escape(textAnalyzer(v.Contains.Value)))

	case *pb.FieldQuery_StartsWith:
		return fmt.Sprintf("(value->>'$.%s' like '%s%%')", v.StartsWith.Field, escape(textAnalyzer(v.StartsWith.Value)))

	case *pb.FieldQuery_EndsWith:
		return fmt.Sprintf("(value->>'$.%s' like '%s%%')", v.EndsWith.Field, escape(textAnalyzer(v.EndsWith.Value)))

	case *pb.FieldQuery_StrEqual:
		return fmt.Sprintf("(value->>'$.%s'='%s')", v.StrEqual.Field, escape(textAnalyzer(v.StrEqual.Value)))

	case *pb.FieldQuery_Lt:
		return fmt.Sprintf("(value->>'$.%s'<%d)", v.Lt.Field, v.Lt.Value)

	case *pb.FieldQuery_Lte:
		return fmt.Sprintf("(value->>'$.%s'<=%d)", v.Lte.Field, v.Lte.Value)

	case *pb.FieldQuery_Gt:
		return fmt.Sprintf("(value->>'$.%s'>%d)", v.Gt.Field, v.Gt.Value)

	case *pb.FieldQuery_Gte:
		return fmt.Sprintf("(value->>'$.%s'>=%d)", v.Gte.Field, v.Gte.Value)

	case *pb.FieldQuery_NumbEq:
		return fmt.Sprintf("(value->>'$.%s'=%d)", v.NumbEq.Field, v.NumbEq.Value)
	}

	return ""
}

func escape(value string) string {
	return strings.Replace(value, "'", `\'`, -1)
}

type aggregatedStrIdsCursor struct {
	cursor      bome.Cursor
	currentList []string
}

func (c *aggregatedStrIdsCursor) Next() (string, error) {
	for {
		if len(c.currentList) == 0 {
			if !c.cursor.HasNext() {
				return "", io.EOF
			}

			o, err := c.cursor.Next()
			if err != nil {
				return "", err
			}

			c.currentList = strings.Split(o.(string), "<>")
		}

		next := strings.Trim(c.currentList[0], " ")
		if next == "" {
			continue
		}

		c.currentList = c.currentList[1:]
		return next, nil
	}
}

func (c *aggregatedStrIdsCursor) Close() error {
	return nil
}

type bomeCursorWrapper struct {
	cursor bome.Cursor
}

func (c *bomeCursorWrapper) Next() (string, error) {
	if c.cursor.HasNext() {
		o, err := c.cursor.Next()
		if err == nil {
			return o.(string), nil
		}
		return "", err
	}
	return "", io.EOF
}

func (c *bomeCursorWrapper) Close() error {
	return c.cursor.Close()
}
