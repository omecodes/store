package auth

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/omecodes/bome"
	"github.com/omecodes/libome/logs"
	"golang.org/x/oauth2"
)

type Provider struct {
	Name   string         `json:"name,omitempty"`
	Label  string         `json:"label,omitempty"`
	Config *oauth2.Config `json:"config,omitempty"`
	Color  string         `json:"color,omitempty"`
	Active bool           `json:"active,omitempty"`
}

type ProviderManager interface {
	Save(provider *Provider) error
	Get(name string) (*Provider, error)
	GetAll(hideConfig bool) ([]*Provider, error)
	Delete(name string) error
}

func NewProviderSQLManager(db *sql.DB, dialect string, tableName string) (*sqlProviderManager, error) {
	store, err := bome.Build().
		SetConn(db).
		SetDialect(dialect).
		SetTableName(tableName).
		JSONMap()
	if err != nil {
		return nil, err
	}
	return &sqlProviderManager{store: store}, nil
}

type sqlProviderManager struct {
	store *bome.JSONMap
}

func (s *sqlProviderManager) Save(provider *Provider) error {
	data, err := json.Marshal(provider)
	if err != nil {
		return err
	}

	return s.store.Upsert(&bome.MapEntry{
		Key:   provider.Name,
		Value: string(data),
	})
}

func (s *sqlProviderManager) Get(name string) (*Provider, error) {
	strEncoded, err := s.store.Get(name)
	if err != nil {
		return nil, err
	}

	var provider *Provider
	err = json.NewDecoder(bytes.NewBufferString(strEncoded)).Decode(&provider)
	return provider, err
}

func (s *sqlProviderManager) GetAll(hideConfig bool) ([]*Provider, error) {
	c, err := s.store.List()
	if err != nil {
		return nil, err
	}

	defer func() {
		if cer := c.Close(); cer != nil {
			logs.Error("AUTH providers list cursor close", logs.Err(err))
		}
	}()

	var providers []*Provider

	for c.HasNext() {
		o, err := c.Next()
		if err != nil {
			return nil, err
		}

		entry := o.(*bome.MapEntry)
		var provider *Provider
		err = json.NewDecoder(bytes.NewBufferString(entry.Value)).Decode(&provider)
		if err != nil {
			return nil, err
		}

		providers = append(providers, provider)
	}

	return providers, nil
}

func (s *sqlProviderManager) Delete(name string) error {
	return s.store.Delete(name)
}
