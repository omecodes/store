package accounts

import "database/sql"

func NewSQLManager(db *sql.DB, dialect string, tablePrefix string) (Manager, error) {
	return &sqlManager{}, nil
}

type sqlManager struct {
}

func (s *sqlManager) Create(account *Account) error {
	panic("implement me")
}

func (s *sqlManager) Get(username string) (*Account, error) {
	panic("implement me")
}

func (s *sqlManager) Find(provider string, originalName string) (*Account, error) {
	panic("implement me")
}

func (s *sqlManager) Search(pattern string) ([]string, error) {
	panic("implement me")
}
