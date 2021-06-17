package auth

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	pb "github.com/omecodes/store/gen/go/proto"

	"github.com/omecodes/bome"
	"github.com/omecodes/libome/crypt"
)

type CredentialsManager interface {
	ValidateAdminAccess(password string) error
	SaveClientApp(access *pb.ClientApp) error
	GetClientApp(key string) (*pb.ClientApp, error)
	GetAllClientApps() ([]*pb.ClientApp, error)
	DeleteClientApp(key string) error

	SaveUserCredentials(credentials *pb.UserCredentials) error
	GetUserPassword(username string) (string, error)
	GetMatchingUser(pattern string) ([]string, error)
	DeleteUserCredentials(username string) error
}

func NewCredentialsSQLManager(db *sql.DB, dialect string, tablePrefix string, adminInfo string) (*credentialsSQLManager, error) {
	clientsTableName := tablePrefix + "_client_apps"
	clients, err := bome.Build().
		SetConn(db).
		SetDialect(dialect).
		SetTableName(clientsTableName).
		JSONMap()
	if err != nil {
		return nil, err
	}

	usersTableName := tablePrefix + "_users"
	users, err := bome.Build().
		SetConn(db).
		SetDialect(dialect).
		SetTableName(usersTableName).
		Map()
	if err != nil {
		return nil, err
	}

	data, err := base64.RawStdEncoding.DecodeString(adminInfo)
	if err != nil {
		logs.Error("Unreadable admin info", logs.Err(err))
		return nil, errors.BadRequest("")
	}

	var info *crypt.Info

	err = json.Unmarshal(data, &info)
	if err != nil {
		logs.Error("Unreadable admin info", logs.Err(err))
		return nil, errors.BadRequest("")
	}

	return &credentialsSQLManager{
		clientsTableName: clientsTableName,
		usersTableName:   usersTableName,
		clients:          clients,
		users:            users,
		adminInfo:        info,
	}, nil
}

type credentialsSQLManager struct {
	clientsTableName string
	usersTableName   string
	clients          *bome.JSONMap
	users            *bome.Map
	adminInfo        *crypt.Info
}

func (s *credentialsSQLManager) ValidateAdminAccess(passPhrase string) error {
	_, err := crypt.Reveal(passPhrase, s.adminInfo)
	return err
}

func (s *credentialsSQLManager) SaveClientApp(access *pb.ClientApp) error {
	data, err := json.Marshal(access)
	if err != nil {
		return err
	}

	return s.clients.Upsert(&bome.MapEntry{
		Key:   access.Key,
		Value: string(data),
	})
}

func (s *credentialsSQLManager) GetClientApp(key string) (*pb.ClientApp, error) {
	strEncoded, err := s.clients.Get(key)
	if err != nil {
		return nil, err
	}

	var access *pb.ClientApp
	err = json.NewDecoder(bytes.NewBufferString(strEncoded)).Decode(&access)
	return access, err
}

func (s *credentialsSQLManager) GetAllClientApps() ([]*pb.ClientApp, error) {
	cursor, err := s.clients.List()
	if err != nil {
		return nil, err
	}

	defer func() {
		if cer := cursor.Close(); cer != nil {
			logs.Error("cursor close", logs.Err(cer))
		}
	}()

	var accesses []*pb.ClientApp
	for cursor.HasNext() {
		o, err := cursor.Next()
		if err != nil {
			return nil, err
		}

		entry := o.(*bome.MapEntry)
		var access *pb.ClientApp
		err = json.Unmarshal([]byte(entry.Value), &access)
		if err != nil {
			return nil, err
		}
		accesses = append(accesses, access)
	}
	return accesses, nil
}

func (s *credentialsSQLManager) DeleteClientApp(key string) error {
	return s.clients.Delete(key)
}

func (s *credentialsSQLManager) SaveUserCredentials(credentials *pb.UserCredentials) error {
	return s.users.Save(&bome.MapEntry{
		Key:   credentials.Username,
		Value: credentials.Password,
	})
}

func (s *credentialsSQLManager) GetMatchingUser(pattern string) ([]string, error) {
	sqlQuery := fmt.Sprintf("select name from %s where name like ? limit 10", s.usersTableName)
	c, err := s.users.Query(sqlQuery, bome.StringScanner, "%"+pattern+"%")
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := c.Close(); cerr != nil {
			logs.Error("user list cursor close", logs.Err(cerr))
		}
	}()

	var usernames []string
	for c.HasNext() {
		o, err := c.Next()
		if err != nil {
			logs.Error("could not get username from cursor", logs.Err(err))
			return nil, errors.Internal("internal db error")
		}
		usernames = append(usernames, o.(string))
	}
	return usernames, nil
}

func (s *credentialsSQLManager) GetUserPassword(username string) (string, error) {
	return s.users.Get(username)
}

func (s *credentialsSQLManager) DeleteUserCredentials(username string) error {
	return s.users.Delete(username)
}
