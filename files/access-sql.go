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
	"net/url"
	"strings"
	"time"

	"github.com/omecodes/bome"
	"github.com/omecodes/libome/logs"
)

func NewAccessSQLManager(db *sql.DB, dialect string, tablePrefix string) (*accessSQLManager, error) {
	builder := bome.Build().
		SetConn(db).
		SetDialect(dialect)

	accesses, err := builder.SetTableName(tablePrefix + "_fs_accesses").JSONMap()
	if err != nil {
		return nil, err
	}

	resolved, err := builder.SetTableName(tablePrefix + "_fs_accesses_resolved").JSONMap()
	if err != nil {
		return nil, err
	}

	return &accessSQLManager{
		accessesMap:         accesses,
		resolvedAccessesMap: resolved,
	}, err
}

type accessSQLManager struct {
	accessesMap         *bome.JSONMap
	resolvedAccessesMap *bome.JSONMap
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

func (s *accessSQLManager) Save(ctx context.Context, access *pb.FSAccess) (string, error) {
	var err error
	if access.Id == "" {
		access.Id, err = s.generateID()
		if err != nil {
			return "", err
		}
	}

	encoded, err := json.Marshal(access)
	if err != nil {
		return "", err
	}

	var (
		accesses        *bome.JSONMap
		resolved        *bome.JSONMap
		encodedResolved string
	)

	if access.Type == pb.AccessType_Default {
		resolvedAccess, err := s.resolveFSAccess(ctx, access)
		if err != nil {
			return "", err
		}

		encodedResolved, err = (&jsonpb.Marshaler{}).MarshalToString(resolvedAccess)
		if err != nil {
			return "", err
		}
	}

	ctx, accesses, err = s.accessesMap.Transaction(ctx)
	if err != nil {
		return "", err
	}

	err = accesses.Upsert(&bome.MapEntry{
		Key:   access.Id,
		Value: string(encoded),
	})
	if err != nil {
		_ = accesses.Rollback()
		return "", err
	}

	if encodedResolved != "" {
		ctx, resolved, err = s.resolvedAccessesMap.Transaction(ctx)
		if err != nil {
			if rbe := bome.Rollback(ctx); rbe != nil {
				logs.Error("could not rollback transaction", logs.Err(err))
			}
			return "", err
		}

		err = resolved.Upsert(&bome.MapEntry{
			Key:   access.Id,
			Value: encodedResolved,
		})
		if err != nil {
			_ = accesses.Rollback()
			return "", err
		}
	}

	return access.Id, bome.Commit(ctx)
}

func (s *accessSQLManager) Get(_ context.Context, id string) (*pb.FSAccess, error) {
	strEncoded, err := s.accessesMap.Get(id)
	if err != nil {
		return nil, err
	}

	access := &pb.FSAccess{}
	err = jsonpb.UnmarshalString(strEncoded, access)
	return access, err
}

func (s *accessSQLManager) GetResolved(_ context.Context, id string) (*pb.FSAccess, error) {

	strEncoded, err := s.resolvedAccessesMap.Get(id)
	if err != nil {
		return nil, err
	}

	access := &pb.FSAccess{}
	err = jsonpb.UnmarshalString(strEncoded, access)
	return access, err
}

func (s *accessSQLManager) Delete(ctx context.Context, id string) error {
	var (
		accesses *bome.JSONMap
		resolved *bome.JSONMap
		err      error
	)

	ctx, accesses, err = s.accessesMap.Transaction(ctx)
	if err != nil {
		return err
	}

	err = accesses.Delete(id)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	ctx, resolved, err = s.resolvedAccessesMap.Transaction(ctx)
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

func (s *accessSQLManager) resolveFSAccess(ctx context.Context, access *pb.FSAccess) (*pb.FSAccess, error) {
	resolvedAccess := access
	accessIDChain := []string{access.Id}

	actionAuthorizeUsers := access.ActionAclRelation

	for resolvedAccess.Type == pb.AccessType_Reference {
		u, err := url.Parse(access.Uri)
		if err != nil {
			return nil, errors.Internal("could not resolve access uri", errors.Details{Key: "access uri", Value: err})
		}

		refAccessID := u.Host
		resolvedAccess, err = s.Get(ctx, refAccessID)
		if err != nil {
			logs.Error("could not load access", logs.Details("access", refAccessID), logs.Err(err))
			return nil, err
		}

		if actionAuthorizeUsers.Edit == nil {
			actionAuthorizeUsers.Edit = resolvedAccess.ActionAclRelation.Edit
		}

		if actionAuthorizeUsers.Share == nil {
			actionAuthorizeUsers.Share = resolvedAccess.ActionAclRelation.Share
		}

		if actionAuthorizeUsers.View == nil {
			actionAuthorizeUsers.View = resolvedAccess.ActionAclRelation.View
		}

		if actionAuthorizeUsers.Delete == nil {
			actionAuthorizeUsers.Delete = resolvedAccess.ActionAclRelation.Delete
		}

		for _, src := range accessIDChain {
			if src == refAccessID {
				return nil, errors.Internal("access cycle referencing")
			}
		}
		accessIDChain = append(accessIDChain, refAccessID)
		resolvedAccess.Uri = strings.TrimSuffix(resolvedAccess.Uri, "/") + u.Path

		logs.Info("resolved access", logs.Details("uri", resolvedAccess.Uri))
	}

	resolvedAccess.ActionAclRelation = actionAuthorizeUsers
	return resolvedAccess, nil
}
