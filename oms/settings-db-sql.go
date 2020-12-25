package oms

import (
	"database/sql"
	"github.com/omecodes/bome"
)

func NewSQLSettings(db *sql.DB, dialect string, tableName string) (SettingsManager, error) {
	m, err := bome.NewMap(db, dialect, tableName)
	if err != nil {
		return nil, err
	}
	return &settingsSQL{bMap: m}, nil
}

type settingsSQL struct {
	bMap *bome.Map
}

func (s *settingsSQL) Set(name string, value string) error {
	return s.bMap.Save(&bome.MapEntry{
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
