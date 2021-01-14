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

const wordsTableName = "$prefix$_words"

const numbersTableName = "$prefix$_numbers"

const propsTableName = "$prefix$_props"

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
		expr, scorers := evaluateWordSearchingQuery(q.Text)
		sqlQuery := "select * from " + wordsTableName + " where " + expr
		c, err := s.db.RawQuery(sqlQuery, bome.MapEntryScanner)
		if err != nil {
			return nil, err
		}

		if len(scorers) <= 1 {
			return &dbCursorWrapper{cursor: c}, nil
		}

		defer func() {
			_ = c.Close()
		}()

		scorers = append(scorers, presenceScorer)
		records := &scoreRecords{}

		for c.HasNext() {
			o, err := c.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}

			entry := o.(*bome.MapEntry)
			for _, scorer := range scorers {
				scorer(entry.Key, entry.Value, records)
			}
		}

		return &idListCursor{ids: records.sorted(), pos: 0}, err

	case *pb.SearchQuery_Number:
		sqlQuery := "select id from " + numbersTableName + " where " + evaluateNumberSearchingQuery(q.Number)
		c, err := s.db.RawQuery(sqlQuery, bome.StringScanner)
		return &aggregatedStrIdsCursor{cursor: c}, err

	case *pb.SearchQuery_Fields:
		sqlQuery := "select id from " + propsTableName + " where " + evaluatePropertiesSearchingQuery(q.Fields)
		c, err := s.db.RawQuery(sqlQuery, bome.StringScanner)
		return &dbCursorWrapper{cursor: c}, err
	}

	return nil, errors.New("unsupported query type")
}

func evaluateWordSearchingQuery(query *pb.StrQuery) (string, []tokenMatchScorer) {
	var matchScorers []tokenMatchScorer
	textAnalyzer := getQueryTextAnalyzer()

	switch v := query.Bool.(type) {

	case *pb.StrQuery_Or:
		var evaluatedExpression []string
		for _, ox := range v.Or.Queries {
			expr, scorers := evaluateWordSearchingQuery(ox)
			evaluatedExpression = append(evaluatedExpression, expr)
			matchScorers = append(matchScorers, scorers...)
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " OR ")), matchScorers

	case *pb.StrQuery_Contains:
		value := textAnalyzer(v.Contains.Value)
		matchScorers = append(matchScorers, containsScorer(value))
		return "(token like '%" + value + "%')", matchScorers

	case *pb.StrQuery_StartsWith:
		value := textAnalyzer(v.StartsWith.Value)
		matchScorers = append(matchScorers, startsWithScorer(value))
		return "(token like '" + value + "%')", matchScorers

	case *pb.StrQuery_EndsWith:
		value := textAnalyzer(v.EndsWith.Value)
		matchScorers = append(matchScorers, endsWithScorer(value))
		return "(token='%" + textAnalyzer(v.EndsWith.Value) + "')", matchScorers

	case *pb.StrQuery_Eq:
		value := textAnalyzer(v.Eq.Value)
		matchScorers = append(matchScorers, equalsScorer(value))
		return "(token='" + textAnalyzer(v.Eq.Value) + "')", matchScorers
	}
	return "", nil
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
