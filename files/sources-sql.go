package files

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/omecodes/errors"
	pb "github.com/omecodes/store/gen/go/proto"
	"time"

	"github.com/omecodes/bome"
	"github.com/omecodes/libome/logs"
)

func NewAccessSQLManager(db *sql.DB, dialect string, tablePrefix string) (*accessSQLManager, error) {
	builder := bome.Build().
		SetConn(db).
		SetDialect(dialect)

	sources, err := builder.SetTableName(tablePrefix + "_sources").JSONMap()
	if err != nil {
		return nil, err
	}

	resolved, err := builder.SetTableName(tablePrefix + "_sources_resolved").JSONMap()
	if err != nil {
		return nil, err
	}

	return &accessSQLManager{
		accessDB: sources,
		resolved: resolved,
	}, err
}

type accessSQLManager struct {
	accessDB *bome.JSONMap
	resolved *bome.JSONMap
}

func (s *accessSQLManager) generateID() (string, error) {
	idBytes := make([]byte, 6)
	_, err := rand.Read(idBytes[:2])
	if err != nil {
		return "", err
	}

	binary.BigEndian.PutUint64(idBytes[3:], uint64(time.Now().Unix()))
	return string(idBytes), nil
}

func (s *accessSQLManager) Save(ctx context.Context, source *pb.Access) (string, error) {
	var err error
	if source.Id == "" {
		source.Id, err = s.generateID()
		if err != nil {
			return "", err
		}
	}

	encoded, err := json.Marshal(source)
	if err != nil {
		return "", err
	}

	var (
		sources         *bome.JSONMap
		resolved        *bome.JSONMap
		encodedResolved string
	)

	if source.Type == pb.AccessType_Default {
		resolvedSource, err := s.resolveSource(ctx, source)
		if err != nil {
			return "", err
		}

		encodedResolved, err = (&jsonpb.Marshaler{}).MarshalToString(resolvedSource)
		if err != nil {
			return "", err
		}
	}

	ctx, sources, err = s.accessDB.Transaction(ctx)
	if err != nil {
		return "", err
	}

	err = sources.Upsert(&bome.MapEntry{
		Key:   source.Id,
		Value: string(encoded),
	})
	if err != nil {
		_ = sources.Rollback()
		return "", err
	}

	if encodedResolved != "" {
		ctx, resolved, err = s.resolved.Transaction(ctx)
		if err != nil {
			if rbe := bome.Rollback(ctx); rbe != nil {
				logs.Error("could not rollback transaction", logs.Err(err))
			}
			return "", err
		}

		err = resolved.Upsert(&bome.MapEntry{
			Key:   source.Id,
			Value: encodedResolved,
		})
		if err != nil {
			_ = sources.Rollback()
			return "", err
		}
	}

	return source.Id, bome.Commit(ctx)
}

func (s *accessSQLManager) Get(_ context.Context, id string) (*pb.Access, error) {
	hasResolvedVersion, err := s.resolved.Contains(id)
	if err != nil {
		return nil, err
	}

	var strEncoded string

	if hasResolvedVersion {
		strEncoded, err = s.resolved.Get(id)
	} else {
		strEncoded, err = s.accessDB.Get(id)
	}
	if err != nil {
		return nil, err
	}

	source := &pb.Access{}
	err = jsonpb.UnmarshalString(strEncoded, source)
	return source, err
}

func (s *accessSQLManager) Delete(ctx context.Context, id string) error {
	var (
		sources  *bome.JSONMap
		resolved *bome.JSONMap
		err      error
	)

	ctx, sources, err = s.accessDB.Transaction(ctx)
	if err != nil {
		return err
	}

	err = sources.Delete(id)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	ctx, resolved, err = s.resolved.Transaction(ctx)
	if err != nil {
		if rbe := bome.Rollback(ctx); rbe != nil {
			logs.Error("could not rollback transaction", logs.Err(err))
		}
		return err
	}

	err = resolved.Delete(id)
	if err != nil {
		if rbe := bome.Rollback(ctx); rbe != nil {
			logs.Error("could not rollback transaction", logs.Err(err))
		}
		return err
	}

	return bome.Commit(ctx)
}

func (s *accessSQLManager) resolveSource(ctx context.Context, access *pb.Access) (*pb.Access, error) {
	/*resolvedSource := source
	sourceChain := []string{source.Id}

	permissionOverrides := source.PermissionOverrides

	for resolvedSource.Type == pb.SourceType_Reference {
		u, err := url.Parse(source.Uri)
		if err != nil {
			return nil, errors.Internal("could not resolve source uri", errors.Details{Key: "source uri", Value: err})
		}

		refSourceID := u.Host
		resolvedSource, err = s.Get(ctx, refSourceID)
		if err != nil {
			logs.Error("could not load source", logs.Details("source", refSourceID), logs.Err(err))
			return nil, err
		}

		if permissionOverrides != nil {
			resolvedSource.PermissionOverrides = permissionOverrides
		} else {
			permissionOverrides = resolvedSource.PermissionOverrides
		}

		for _, src := range sourceChain {
			if src == refSourceID {
				return nil, errors.Internal("source cycle references")
			}
		}
		sourceChain = append(sourceChain, refSourceID)
		resolvedSource.Uri = strings.TrimSuffix(resolvedSource.Uri, "/") + u.Path

		logs.Info("resolved source", logs.Details("uri", resolvedSource.Uri))
	}

	return resolvedSource, nil */
	return access, nil
}
