package auth

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/libome/crypt"
)

type APIAccess struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type CredentialsManager interface {
	VerifyAdminCredentials(passPhrase string) error

	SaveAccess(access *APIAccess) error
	GetAccess(key string) (*APIAccess, error)
	DeleteAccess(key string) error
}

func NewSQLCredentialsManager(db *sql.DB, dialect string, prefix string, adminInfo string) (*sqlCredentialsManager, error) {
	store, err := bome.NewJSONMap(db, dialect, prefix+"_api_accesses")
	if err != nil {
		return nil, err
	}

	data, err := base64.RawStdEncoding.DecodeString(adminInfo)
	if err != nil {
		log.Error("Unreadable admin info", log.Err(err))
		return nil, errors.BadInput
	}

	var info *crypt.Info

	err = json.Unmarshal(data, &info)
	if err != nil {
		log.Error("Unreadable admin info", log.Err(err))
		return nil, errors.BadInput
	}

	return &sqlCredentialsManager{store: store, adminInfo: info}, nil
}

type sqlCredentialsManager struct {
	store     *bome.JSONMap
	adminInfo *crypt.Info
}

func (s *sqlCredentialsManager) VerifyAdminCredentials(passPhrase string) error {
	_, err := crypt.Reveal(passPhrase, s.adminInfo)
	return err
}

func (s *sqlCredentialsManager) SaveAccess(access *APIAccess) error {
	data, err := json.Marshal(access)
	if err != nil {
		return err
	}

	return s.store.Upsert(&bome.MapEntry{
		Key:   access.Key,
		Value: string(data),
	})
}

func (s *sqlCredentialsManager) GetAccess(key string) (*APIAccess, error) {
	strEncoded, err := s.store.Get(key)
	if err != nil {
		return nil, err
	}

	var access *APIAccess
	err = json.NewDecoder(bytes.NewBufferString(strEncoded)).Decode(&access)
	return access, err
}

func (s *sqlCredentialsManager) DeleteAccess(key string) error {
	return s.store.Delete(key)
}
