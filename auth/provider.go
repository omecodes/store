package auth

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/utils/log"
	"golang.org/x/oauth2"
)

type Provider struct {
	Name   string         `json:"name"`
	Label  string         `json:"label"`
	Config *oauth2.Config `json:"config"`
	Color  string         `json:"color"`
	Active bool           `json:"active"`
}

type ProviderStore interface {
	Save(provider *Provider) error
	Get(name string) (*Provider, error)
	GetAll(hideConfig bool) ([]*Provider, error)
	Delete(name string) error
}

func NewSQLProviderStore(db *sql.DB, dialect string, tableName string) (*sqlProviderStore, error) {
	store, err := bome.NewJSONMap(db, dialect, tableName)
	if err != nil {
		return nil, err
	}
	return &sqlProviderStore{store: store}, nil
}

type sqlProviderStore struct {
	store *bome.JSONMap
}

func (s *sqlProviderStore) Save(provider *Provider) error {
	data, err := json.Marshal(provider)
	if err != nil {
		return err
	}

	return s.store.Upsert(&bome.MapEntry{
		Key:   provider.Name,
		Value: string(data),
	})
}

func (s *sqlProviderStore) Get(name string) (*Provider, error) {
	strEncoded, err := s.store.Get(name)
	if err != nil {
		return nil, err
	}

	var provider *Provider
	err = json.NewDecoder(bytes.NewBufferString(strEncoded)).Decode(&provider)
	return provider, err
}

func (s *sqlProviderStore) GetAll(hideConfig bool) ([]*Provider, error) {
	c, err := s.store.List()
	if err != nil {
		return nil, err
	}

	defer func() {
		if cer := c.Close(); cer != nil {
			log.Error("AUTH providers list cursor close", log.Err(err))
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

func (s *sqlProviderStore) Delete(name string) error {
	return s.store.Delete(name)
}
