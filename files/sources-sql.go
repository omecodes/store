package files

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/omecodes/store/auth"
	"time"

	"github.com/omecodes/bome"
	"github.com/omecodes/libome/logs"
)

func NewSourceSQLManager(db *sql.DB, dialect string, tablePrefix string) (*sourceSQLManager, error) {
	builder := bome.Build().
		SetConn(db).
		SetDialect(dialect)

	sources, err := builder.SetTableName(tablePrefix + "_sources").JSONMap()
	if err != nil {
		return nil, err
	}

	userRefs, err := builder.SetTableName(tablePrefix + "_sources_user_refs").DoubleMap()
	if err != nil {
		return nil, err
	}

	return &sourceSQLManager{
		sources:  sources,
		userRefs: userRefs,
	}, err
}

type sourceSQLManager struct {
	sources  *bome.JSONMap
	userRefs *bome.DoubleMap
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

func (s *sourceSQLManager) Save(ctx context.Context, source *Source) (string, error) {
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

	var sources *bome.JSONMap
	var userRefs *bome.DoubleMap

	ctx, sources, err = s.sources.Transaction(ctx)
	if err != nil {
		return "", err
	}

	err = sources.Upsert(&bome.MapEntry{
		Key:   source.ID,
		Value: string(encoded),
	})
	if err != nil {
		_ = sources.Rollback()
	}

	ctx, userRefs, err = s.userRefs.Transaction(ctx)
	if err != nil {
		if rbe := bome.Rollback(ctx); rbe != nil {
			logs.Error("could not rollback transaction", logs.Err(err))
		}
		return "", err
	}

	if source.PermissionOverrides != nil {
		creator := source.CreatedBy

		perms := append([]*auth.Permission{}, source.PermissionOverrides.Read...)
		perms = append([]*auth.Permission{}, source.PermissionOverrides.Write...)
		perms = append([]*auth.Permission{}, source.PermissionOverrides.Chmod...)

		for _, perm := range perms {
			for _, user := range perm.RelatedUsers {
				err = userRefs.Upsert(&bome.DoubleMapEntry{
					FirstKey:  user,
					SecondKey: source.ID,
					Value:     creator,
				})
				if err != nil {
					if rbe := bome.Rollback(ctx); rbe != nil {
						logs.Error("could not rollback transaction", logs.Err(err))
					}
					return "", err
				}
			}
		}
	}

	return source.ID, bome.Commit(ctx)
}

func (s *sourceSQLManager) Get(_ context.Context, id string) (*Source, error) {
	strEncoded, err := s.sources.Get(id)
	if err != nil {
		return nil, err
	}
	var source *Source

	err = json.NewDecoder(bytes.NewBufferString(strEncoded)).Decode(&source)
	return source, err
}

func (s *sourceSQLManager) Delete(ctx context.Context, id string) error {
	var (
		sources  *bome.JSONMap
		userRefs *bome.DoubleMap
		err      error
	)

	ctx, sources, err = s.sources.Transaction(ctx)
	if err != nil {
		return err
	}

	err = sources.Delete(id)
	if err != nil && !bome.IsNotFound(err) {
		return err
	}

	ctx, userRefs, err = s.userRefs.Transaction(ctx)
	if err != nil {
		if rbe := bome.Rollback(ctx); rbe != nil {
			logs.Error("could not rollback transaction", logs.Err(err))
		}
		return err
	}

	err = userRefs.DeleteAllMatchingSecondKey(id)
	if err != nil {
		if rbe := bome.Rollback(ctx); rbe != nil {
			logs.Error("could not rollback transaction", logs.Err(err))
		}
		return err
	}

	return bome.Commit(ctx)
}

func (s *sourceSQLManager) UserSources(ctx context.Context, username string) ([]*Source, error) {
	query := fmt.Sprintf("select sources.value from %s as sources, %s as refs where sources.name=refs.second_key and refs.first_key=?",
		s.sources.Table(),
		s.userRefs.Table(),
	)

	cursor, err := s.sources.Query(query, bome.StringScanner, username)
	if err != nil {
		return nil, err
	}

	defer func() {
		if clerr := cursor.Close(); clerr != nil {
			logs.Error("close cursor", logs.Err(clerr))
		}
	}()

	var sources []*Source
	for cursor.HasNext() {
		o, err := cursor.Next()
		if err != nil {
			return nil, err
		}

		var source *Source
		err = json.Unmarshal([]byte(o.(string)), &source)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}
