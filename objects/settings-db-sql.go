package objects

import (
	"database/sql"
	"github.com/omecodes/bome"
)

func NewSQLSettings(db *sql.DB, dialect string, tableName string) (SettingsManager, error) {
	m, err := bome.Build().SetConn(db).SetDialect(dialect).SetTableName(tableName).Map()
	if err != nil {
		return nil, err
	}

	settings := &settingsSQL{bMap: m}
	_, err = settings.Get(SettingsDataMaxSizePath)
	if err != nil {
		if !bome.IsNotFound(err) {
			return nil, err
		}
		err = settings.Set(SettingsDataMaxSizePath, DefaultSettings[SettingsDataMaxSizePath])
		if err != nil && !bome.IsPrimaryKeyConstraintError(err) {
			return nil, err
		}
	}

	_, err = settings.Get(SettingsDataMaxSizePath)
	if err != nil {
		if !bome.IsNotFound(err) {
			return nil, err
		}
		err = settings.Set(SettingsCreateDataSecurityRule, DefaultSettings[SettingsCreateDataSecurityRule])
		if err != nil && !bome.IsPrimaryKeyConstraintError(err) {
			return nil, err
		}
	}

	_, err = settings.Get(SettingsObjectListMaxCount)
	if err != nil {
		if !bome.IsNotFound(err) {
			return nil, err
		}
		err = settings.Set(SettingsObjectListMaxCount, DefaultSettings[SettingsObjectListMaxCount])
		if err != nil && !bome.IsPrimaryKeyConstraintError(err) {
			return nil, err
		}
	}

	return settings, nil
}

type settingsSQL struct {
	bMap *bome.Map
}

func (s *settingsSQL) Set(name string, value string) error {
	return s.bMap.Upsert(&bome.MapEntry{
		Key:   name,
		Value: value,
	})
}

func (s *settingsSQL) Get(name string) (string, error) {
	return s.bMap.Get(name)
}

func (s *settingsSQL) Delete(name string) error {
	return s.bMap.Delete(name)
}

func (s *settingsSQL) Clear() error {
	return s.bMap.Clear()
}

func (s *settingsSQL) Close() error {
	return s.bMap.Close()
}
