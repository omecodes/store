package accounts

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/errors"
)

func NewSQLManager(db *sql.DB, dialect string, tablePrefix string) (Manager, error) {
	return &sqlManager{}, nil
}

type sqlManager struct {
	accounts bome.Map
	sources  bome.DoubleMap
}

func (s *sqlManager) Create(ctx context.Context, account *Account) error {
	encoded, err := json.Marshal(account)
	if err != nil {
		return err
	}

	tx, err := s.accounts.BeginTransaction()
	if err != nil {
		return err
	}

	err = tx.Save(&bome.MapEntry{
		Key:   account.Login,
		Value: string(encoded),
	})
	if err != nil {
		if rer := tx.Rollback(); rer != nil {
			log.Error("Transaction rollback failed", log.Err(err))
		}

		if bome.IsPrimaryKeyConstraintError(err) {
			return errors.Create(errors.DuplicateResource, "account_exists", errors.Info{
				Name:    "db",
				Details: err.Error(),
			})
		}
		return errors.Create(errors.Internal, "could_not_create_account", errors.Info{
			Name:    "bome.SQL",
			Details: err.Error(),
		})
	}

	stx := s.sources.ContinueTransaction(tx.TX())
	err = stx.Save(&bome.DoubleMapEntry{
		FirstKey:  account.Source.Provider,
		SecondKey: account.Source.Name,
		Value:     account.Login,
	})
	if err != nil {
		if rer := tx.Rollback(); rer != nil {
			log.Error("Transaction rollback failed", log.Err(err))
		}

		if bome.IsPrimaryKeyConstraintError(err) {
			return errors.Create(errors.DuplicateResource, "account_exists", errors.Info{
				Name:    "db",
				Details: err.Error(),
			})
		}
		return err
	}

	return nil
}

func (s *sqlManager) Get(ctx context.Context, username string) (*Account, error) {
	panic("implement me")
}

func (s *sqlManager) Find(ctx context.Context, provider string, originalName string) (*Account, error) {
	panic("implement me")
}

func (s *sqlManager) Search(ctx context.Context, pattern string) ([]string, error) {
	panic("implement me")
}
