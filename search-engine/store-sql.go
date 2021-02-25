package se

import (
	"database/sql"
	"fmt"
	"github.com/omecodes/errors"
	"io"
	"strings"

	"github.com/omecodes/bome"
)

const wordsTableName = "$prefix$_words"

const numbersTableName = "$prefix$_numbers"

const propsTableName = "$prefix$_props"

const propsTablesDef = `
create table if not exists $prefix$_props (
  	object varchar(255) not null,
    value JSON not null,
    primary key(object)
);
`

const wordsTablesDef = `
create table if not exists $prefix$_words (
  	token varchar(255) not null,
    id varchar(255) not null,
    primary key(token, id)
);
`

const numbersTablesDef = `
create table if not exists $prefix$_numbers (
  	num bigInt not null,
    id varchar(255) not null,
    primary key(num, id)
);
`

const insertWord = `
insert into $prefix$_words values(?, ?);
`

const deleteObjectWordMappings = `
delete $prefix$_words where id=?;
`

const insertNumber = `
insert into $prefix$_numbers values(?, ?);
`

const deleteObjectNumberMapping = `
delete from $prefix$_numbers where id=?;
`

const insertProps = `
insert into $prefix$_props values(?, ?);
`

const deleteProps = `
delete from $prefix$_props where id=?;
`

func NewSQLIndexStore(db *sql.DB, dialect string, tablePrefix string) (Store, error) {
	s := new(sqlStore)
	var err error

	if dialect == bome.SQLite3 {
		s.db, err = bome.NewLite(db)

	} else if dialect == bome.MySQL {
		s.db, err = bome.New(db)

	} else {
		return nil, errors.Unsupported("sql dialect not supported", errors.Details{Key: "type", Value: "dialect"}, errors.Details{Key: "name", Value: dialect})
	}

	if err != nil {
		return nil, err
	}

	err = s.db.Init()
	if err != nil {
		return nil, err
	}

	s.db.SetTablePrefix(tablePrefix)

	err = s.db.Exec(wordsTablesDef).Error
	if err == nil {
		err = s.db.Exec(numbersTablesDef).Error
		if err == nil {
			err = s.db.Exec(propsTablesDef).Error
		}
	}
	return s, err
}

type sqlStore struct {
	db *bome.DB
}

func (s *sqlStore) SaveWordMapping(word string, id string) error {
	err := s.db.Exec(insertWord, word, id).Error
	if errors.IsConflict(err) {
		return nil
	}
	return err
}

func (s *sqlStore) SaveNumberMapping(num int64, id string) error {
	err := s.db.Exec(insertNumber, num, id).Error
	if errors.IsConflict(err) {
		return nil
	}
	return err
}

func (s *sqlStore) SavePropertiesMapping(id string, value string) error {
	err := s.db.Exec(insertProps, id, value).Error
	if errors.IsConflict(err) {
		return nil
	}
	return err
}

func (s *sqlStore) DeleteObjectMappings(id string) error {
	err := s.db.Exec(deleteObjectNumberMapping, id).Error
	if err == nil {
		err = s.db.Exec(deleteObjectNumberMapping, id).Error
		if err == nil {
			s.db.Exec(deleteProps, id)
		}
	}
	return err
}

func (s *sqlStore) Search(query *SearchQuery) (Cursor, error) {
	return s.performSearch(query)
}

func (s *sqlStore) performSearch(query *SearchQuery) (Cursor, error) {
	switch q := query.Query.(type) {

	case *SearchQuery_Text:
		expr, scorers := evaluateWordSearchingQuery(q.Text)
		sqlQuery := "select * from " + wordsTableName + " where " + expr
		c, err := s.db.Query(sqlQuery, bome.MapEntryScanner)
		if err != nil {
			return nil, err
		}

		if len(scorers) <= 1 {
			return &dbMapEntryCursorWrapper{cursor: c}, nil
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

	case *SearchQuery_Number:
		sqlQuery := "select id from " + numbersTableName + " where " + evaluateNumberSearchingQuery(q.Number)
		c, err := s.db.Query(sqlQuery, bome.StringScanner)
		return &aggregatedStrIdsCursor{cursor: c}, err

	case *SearchQuery_Fields:
		sqlQuery := "select id from " + propsTableName + " where " + evaluatePropertiesSearchingQuery(q.Fields)
		c, err := s.db.Query(sqlQuery, bome.StringScanner)
		return &dbStringCursorWrapper{cursor: c}, err
	}

	return nil, errors.Unsupported("sql dialect not supported", errors.Details{Key: "type", Value: "query"}, errors.Details{Key: "name", Value: query.Query})
}

func evaluateWordSearchingQuery(query *StrQuery) (string, []tokenMatchScorer) {
	var matchScorers []tokenMatchScorer
	textAnalyzer := getQueryTextAnalyzer()

	switch v := query.Bool.(type) {

	case *StrQuery_Or:
		var evaluatedExpression []string
		for _, ox := range v.Or.Queries {
			expr, scorers := evaluateWordSearchingQuery(ox)
			evaluatedExpression = append(evaluatedExpression, expr)
			matchScorers = append(matchScorers, scorers...)
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " OR ")), matchScorers

	case *StrQuery_Contains:
		value := textAnalyzer(v.Contains.Value)
		matchScorers = append(matchScorers, containsScorer(value))
		return "(token like '%" + value + "%')", matchScorers

	case *StrQuery_StartsWith:
		value := textAnalyzer(v.StartsWith.Value)
		matchScorers = append(matchScorers, startsWithScorer(value))
		return "(token like '" + value + "%')", matchScorers

	case *StrQuery_EndsWith:
		value := textAnalyzer(v.EndsWith.Value)
		matchScorers = append(matchScorers, endsWithScorer(value))
		return "(token='%" + textAnalyzer(v.EndsWith.Value) + "')", matchScorers

	case *StrQuery_Eq:
		value := textAnalyzer(v.Eq.Value)
		matchScorers = append(matchScorers, equalsScorer(value))
		return "(token='" + textAnalyzer(v.Eq.Value) + "')", matchScorers
	}
	return "", nil
}

func evaluateNumberSearchingQuery(query *NumQuery) string {
	switch v := query.Bool.(type) {

	case *NumQuery_And:
		var evaluatedExpression []string
		for _, ox := range v.And.Queries {
			evaluatedExpression = append(evaluatedExpression, evaluateNumberSearchingQuery(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " AND "))

	case *NumQuery_Or:
		var evaluatedExpression []string
		for _, ox := range v.Or.Queries {
			evaluatedExpression = append(evaluatedExpression, evaluateNumberSearchingQuery(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " OR "))

	case *NumQuery_Eq:
		return fmt.Sprintf("(num=%d)", v.Eq.Value)

	case *NumQuery_Gt:
		return fmt.Sprintf("(num>%d)", v.Gt.Value)

	case *NumQuery_Gte:
		return fmt.Sprintf("(num>=%d)", v.Gte.Value)

	case *NumQuery_Lt:
		return fmt.Sprintf("(num<%d)", v.Lt.Value)

	case *NumQuery_Lte:
		return fmt.Sprintf("(num<=%d)", v.Lte.Value)
	}

	return ""
}

func evaluatePropertiesSearchingQuery(query *FieldQuery) string {
	textAnalyzer := propsMappingTextAnalyzer()

	switch v := query.Bool.(type) {

	case *FieldQuery_And:
		var evaluatedExpression []string
		for _, ox := range v.And.Queries {
			evaluatedExpression = append(evaluatedExpression, evaluatePropertiesSearchingQuery(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " AND "))

	case *FieldQuery_Or:
		var evaluatedExpression []string
		for _, ox := range v.Or.Queries {
			evaluatedExpression = append(evaluatedExpression, evaluatePropertiesSearchingQuery(ox))
		}
		return fmt.Sprintf("(%s)", strings.Join(evaluatedExpression, " OR "))

	case *FieldQuery_Contains:
		return fmt.Sprintf("(value->>'$.%s' like '%%%s%%')", v.Contains.Field, escape(textAnalyzer(v.Contains.Value)))

	case *FieldQuery_StartsWith:
		return fmt.Sprintf("(value->>'$.%s' like '%s%%')", v.StartsWith.Field, escape(textAnalyzer(v.StartsWith.Value)))

	case *FieldQuery_EndsWith:
		return fmt.Sprintf("(value->>'$.%s' like '%s%%')", v.EndsWith.Field, escape(textAnalyzer(v.EndsWith.Value)))

	case *FieldQuery_StrEqual:
		return fmt.Sprintf("(value->>'$.%s'='%s')", v.StrEqual.Field, escape(textAnalyzer(v.StrEqual.Value)))

	case *FieldQuery_Lt:
		return fmt.Sprintf("(value->>'$.%s'<%d)", v.Lt.Field, v.Lt.Value)

	case *FieldQuery_Lte:
		return fmt.Sprintf("(value->>'$.%s'<=%d)", v.Lte.Field, v.Lte.Value)

	case *FieldQuery_Gt:
		return fmt.Sprintf("(value->>'$.%s'>%d)", v.Gt.Field, v.Gt.Value)

	case *FieldQuery_Gte:
		return fmt.Sprintf("(value->>'$.%s'>=%d)", v.Gte.Field, v.Gte.Value)

	case *FieldQuery_NumbEq:
		return fmt.Sprintf("(value->>'$.%s'=%d)", v.NumbEq.Field, v.NumbEq.Value)
	}

	return ""
}

func escape(value string) string {
	return strings.Replace(value, "'", `\'`, -1)
}
