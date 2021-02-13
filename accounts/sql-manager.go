package accounts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/errors"
)

func NewSQLManager(db *sql.DB, dialect string, tablePrefix string) (Manager, error) {
	accounts, err := bome.Build().
		SetConn(db).
		SetDialect(dialect).
		SetTableName(tablePrefix + "_accounts").
		JSONMap()
	if err != nil {
		return nil, err
	}

	sources, err := bome.Build().
		SetConn(db).
		SetDialect(dialect).
		SetTableName(tablePrefix + "_account_sources").
		DoubleMap()
	if err != nil {
		return nil, err
	}

	return &sqlManager{
		tablePrefix: tablePrefix,
		accounts:    accounts,
		sources:     sources,
	}, nil
}

type sqlManager struct {
	tablePrefix string
	accounts    *bome.JSONMap
	sources     *bome.DoubleMap
}

func (s *sqlManager) Create(ctx context.Context, account *Account) error {
	encoded, err := json.Marshal(account)
	if err != nil {
		return err
	}

	txCtx, accounts, err := s.accounts.Transaction(ctx)
	if err != nil {
		return err
	}

	err = accounts.Save(&bome.MapEntry{
		Key:   account.Login,
		Value: string(encoded),
	})
	if err != nil {
		if rer := bome.Rollback(txCtx); rer != nil {
			log.Error("Transaction rollback failed", log.Err(err))
		}
		return errors.AppendDetails(err, errors.Info{
			Name:    "am",
			Details: "could not create entry",
		})
	}

	_, sources, _ := s.sources.Transaction(txCtx)
	err = sources.Save(&bome.DoubleMapEntry{
		FirstKey:  account.Source.Provider,
		SecondKey: account.Source.Name,
		Value:     account.Login,
	})
	if err != nil {
		if rer := bome.Rollback(txCtx); rer != nil {
			log.Error("Transaction rollback failed", log.Err(err))
		}
		return errors.AppendDetails(err, errors.Info{
			Name:    "am",
			Details: "could not create source reference",
		})
	}

	return nil
}

func (s *sqlManager) Get(ctx context.Context, username string) (*Account, error) {
	encoded, err := s.accounts.Get(username)
	if err != nil {
		return nil, errors.AppendDetails(err, errors.Info{
			Name:    "am",
			Details: "failed to get account by name",
		})
	}

	var account *Account
	err = json.Unmarshal([]byte(encoded), &account)
	if err != nil {
		return nil, errors.AppendDetails(err, errors.Info{
			Name:    "am",
			Details: "failed to decode account from db result",
		})
	}
	return account, nil
}

func (s *sqlManager) Find(ctx context.Context, provider string, originalName string) (*Account, error) {
	accountName, err := s.sources.Get(provider, originalName)
	if err != nil {
		return nil, errors.AppendDetails(err, errors.Info{
			Name:    "am",
			Details: "failed to get account by source",
		})
	}
	return s.Get(ctx, accountName)
}

func (s *sqlManager) Search(ctx context.Context, pattern string) ([]string, error) {
	query := fmt.Sprintf("select value from %s_accounts where name like ?", s.tablePrefix)
	cursor, err := s.accounts.Query(query, bome.StringScanner, fmt.Sprintf("%%%s%%", pattern))
	if err != nil {
		return nil, errors.AppendDetails(err, errors.Info{
			Name:    "am",
			Details: "failed to finds accounts matching pattern",
		})
	}

	defer func() {
		if cer := cursor.Close(); cer != nil {
			log.Error("cursor close", log.Err(cer))
		}
	}()

	var names []string

	for cursor.HasNext() {
		o, err := cursor.Next()
		if err != nil {
			return nil, errors.AppendDetails(err, errors.Info{
				Name:    "am",
				Details: "failed to decode account from db result",
			})
		}
		names = append(names, o.(string))
	}
	return names, nil
}
