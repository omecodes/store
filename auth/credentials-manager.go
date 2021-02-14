package auth

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/omecodes/libome/logs"

	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/libome/crypt"
)

type CredentialsManager interface {
	ValidateAdminAccess(password string) error
	SaveAccess(access *APIAccess) error
	GetAccess(key string) (*APIAccess, error)
	GetAllAccesses() ([]*APIAccess, error)
	DeleteAccess(key string) error
}

func NewCredentialsSQLManager(db *sql.DB, dialect string, prefix string, adminInfo string) (*credentialsSQLManager, error) {
	store, err := bome.Build().
		SetConn(db).
		SetDialect(dialect).
		SetTableName(prefix + "_api_accesses").
		JSONMap()
	if err != nil {
		return nil, err
	}

	data, err := base64.RawStdEncoding.DecodeString(adminInfo)
	if err != nil {
		logs.Error("Unreadable admin info", logs.Err(err))
		return nil, errors.BadInput
	}

	var info *crypt.Info

	err = json.Unmarshal(data, &info)
	if err != nil {
		logs.Error("Unreadable admin info", logs.Err(err))
		return nil, errors.BadInput
	}

	return &credentialsSQLManager{store: store, adminInfo: info}, nil
}

type credentialsSQLManager struct {
	store     *bome.JSONMap
	adminInfo *crypt.Info
}

func (s *credentialsSQLManager) ValidateAdminAccess(passPhrase string) error {
	_, err := crypt.Reveal(passPhrase, s.adminInfo)
	return err
}

func (s *credentialsSQLManager) SaveAccess(access *APIAccess) error {
	data, err := json.Marshal(access)
	if err != nil {
		return err
	}

	return s.store.Upsert(&bome.MapEntry{
		Key:   access.Key,
		Value: string(data),
	})
}

func (s *credentialsSQLManager) GetAccess(key string) (*APIAccess, error) {
	strEncoded, err := s.store.Get(key)
	if err != nil {
		return nil, err
	}

	var access *APIAccess
	err = json.NewDecoder(bytes.NewBufferString(strEncoded)).Decode(&access)
	return access, err
}

func (s *credentialsSQLManager) GetAllAccesses() ([]*APIAccess, error) {
	cursor, err := s.store.List()
	if err != nil {
		return nil, err
	}

	defer func() {
		if cer := cursor.Close(); cer != nil {
			logs.Error("cursor close", logs.Err(cer))
		}
	}()

	var accesses []*APIAccess
	for cursor.HasNext() {
		o, err := cursor.Next()
		if err != nil {
			return nil, err
		}

		entry := o.(*bome.MapEntry)
		var access *APIAccess
		err = json.Unmarshal([]byte(entry.Value), &access)
		if err != nil {
			return nil, err
		}
		accesses = append(accesses, access)
	}
	return accesses, nil
}

func (s *credentialsSQLManager) DeleteAccess(key string) error {
	return s.store.Delete(key)
}
