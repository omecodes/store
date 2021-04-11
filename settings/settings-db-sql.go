package settings

import (
	"database/sql"
	"github.com/omecodes/bome"
	"github.com/omecodes/errors"
)

func NewSQLManager(db *sql.DB, dialect string, tableName string) (Manager, error) {
	m, err := bome.Build().SetConn(db).SetDialect(dialect).SetTableName(tableName).Map()
	if err != nil {
		return nil, err
	}

	settings := &sqlManager{bMap: m}
	_, err = settings.Get(DataMaxSizePath)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		err = settings.Set(DataMaxSizePath, Default[DataMaxSizePath])
		if err != nil && !errors.IsConflict(err) {
			return nil, err
		}
	}

	_, err = settings.Get(DataMaxSizePath)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		err = settings.Set(CreateDataSecurityRule, Default[CreateDataSecurityRule])
		if err != nil && !errors.IsConflict(err) {
			return nil, err
		}
	}

	_, err = settings.Get(ObjectListMaxCount)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		err = settings.Set(ObjectListMaxCount, Default[ObjectListMaxCount])
		if err != nil && !errors.IsConflict(err) {
			return nil, err
		}
	}

	return settings, nil
}

type sqlManager struct {
	bMap *bome.Map
}

func (s *sqlManager) Set(name string, value string) error {
	return s.bMap.Upsert(&bome.MapEntry{
		Key:   name,
		Value: value,
	})
}

func (s *sqlManager) Get(name string) (string, error) {
	return s.bMap.Get(name)
}

func (s *sqlManager) Delete(name string) error {
	return s.bMap.Delete(name)
}

func (s *sqlManager) Clear() error {
	return s.bMap.Clear()
}

func (s *sqlManager) Close() error {
	return s.bMap.Close()
}
