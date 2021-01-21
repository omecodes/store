package files

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"github.com/omecodes/bome"
	"github.com/omecodes/libome/logs"
	"time"
)

func NewSourceSQLManager(db *sql.DB, dialect string, tableName string) (*sourceSQLManager, error) {
	m, err := bome.NewMap(db, dialect, tableName)
	return &sourceSQLManager{bMap: m}, err
}

type sourceSQLManager struct {
	bMap *bome.Map
}

func (s *sourceSQLManager) generateID() (string, error) {
	idBytes := make([]byte, 6)
	_, err := rand.Read(idBytes[:2])
	if err != nil {
		return "", err
	}

	binary.BigEndian.PutUint64(idBytes[3:], uint64(time.Now().Unix()))
	return string(idBytes), nil
}

func (s *sourceSQLManager) Save(_ context.Context, source *Source) (string, error) {
	var err error
	if source.ID == "" {
		source.ID, err = s.generateID()
		if err != nil {
			return "", err
		}
	}

	encoded, err := json.Marshal(source)
	if err != nil {
		return "", err
	}

	return source.ID, s.bMap.Upsert(&bome.MapEntry{
		Key:   source.ID,
		Value: string(encoded),
	})
}

func (s *sourceSQLManager) Get(_ context.Context, id string) (*Source, error) {
	strEncoded, err := s.bMap.Get(id)
	if err != nil {
		return nil, err
	}
	var source *Source

	err = json.NewDecoder(bytes.NewBufferString(strEncoded)).Decode(&source)
	return source, err
}

func (s *sourceSQLManager) List(_ context.Context) ([]*Source, error) {
	cursor, err := s.bMap.List()
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := cursor.Close(); cerr != nil {
			logs.Error("cursor close", logs.Err(cerr))
		}
	}()

	var sources []*Source
	for cursor.HasNext() {
		o, err := cursor.Next()
		if err != nil {
			return nil, err
		}

		entry := o.(*bome.MapEntry)
		var source *Source
		err = json.Unmarshal([]byte(entry.Value), &source)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}

func (s *sourceSQLManager) Delete(_ context.Context, id string) error {
	return s.bMap.Delete(id)
}
