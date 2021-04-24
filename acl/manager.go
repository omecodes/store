package acl

import (
	"context"
	"encoding/json"
	"github.com/omecodes/errors"
	pb "github.com/omecodes/store/gen/go/proto"
	"strings"
)

type Manager interface {
	SaveRelation(ctx context.Context, relation *pb.Relation) error
	DeleteRelation(ctx context.Context, relation *pb.Relation) error
	IsValidRelation(ctx context.Context, relation *pb.Relation) (bool, error)

	SaveNamespaceConfig(ctx context.Context, config *pb.NamespaceConfig) error
	GetNamespaceConfig(ctx context.Context, name string) (*pb.NamespaceConfig, error)
	DeleteNamespaceConfig(ctx context.Context, name string) error
}

type defaultManager struct {}

func (d *defaultManager) SaveRelation(ctx context.Context, relation *pb.Relation) error {
	store := relationStore(ctx)
	if store == nil {
		return errors.Internal("could not find relation store in context")
	}
	return store.Save(relation)
}

func (d *defaultManager) DeleteRelation(ctx context.Context, relation *pb.Relation) error {
	store := relationStore(ctx)
	if store == nil {
		return errors.Internal("could not find relation store in context")
	}
	return store.Delete(relation)
}

func (d *defaultManager) IsValidRelation(ctx context.Context, relation *pb.Relation) (bool, error) {
	store := relationStore(ctx)
	if store == nil {
		return false, errors.Internal("missing relation store in context")
	}

	nsConfigStore := namespaceConfigStore(ctx)
	if nsConfigStore == nil {
		return false, errors.Internal("missing namespace config in context")
	}

	namespaceId := strings.Trim(strings.Split(relation.Target, ":")[0], " ")
	namespace, err := nsConfigStore.GetNamespace(namespaceId)
	if err != nil {
		return false, errors.Internal("missing relation store in context")
	}

	rel, exists := namespace.Relations[relation.Name]
	if !exists {
		return false, errors.NotFound("relation does not exists")
	}

	for _, rewrite := range rel.UserSetRewrite {

		switch rewrite.Type {

		case pb.UserSetType_This:

			exists, err = store.Exists(relation)
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return false, err
			}
			if exists {
				return true, nil
			}

		case pb.UserSetType_Computed:

			exists, err = d.IsValidRelation(ctx, &pb.Relation{
				Sid:        relation.Sid,
				Name:       rewrite.Value,
				Subject:    relation.Subject,
				Target:     relation.Name,
				CommitTime: relation.CommitTime,
			})
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return false, err
			}
			if exists {
				return true, nil
			}

		case pb.UserSetType_FromTuple:

			var subjects []string
			var definition *pb.UsersInRelationWithObjectRelatedObject
			err = json.Unmarshal([]byte(rewrite.Value), &definition)
			if err != nil {
				return false, err
			}

			subjects, err = store.GetSubjects(&pb.RelationSubjectInfo{
				Name:       definition.ObjectRelation,
				Target:     relation.Target,
				CommitTime: relation.CommitTime,
			})
			if err != nil {
				return false, err
			}

			for _, subject := range subjects {
				exists, err = d.IsValidRelation(ctx, &pb.Relation{
					Sid:        relation.Sid,
					Name:       definition.UserRelation,
					Subject:    relation.Subject,
					Target:     subject,
					CommitTime: relation.CommitTime,
				})

				if err != nil {
					if errors.IsNotFound(err) {
						continue
					}
					return false, err
				}

				if exists {
					return true, nil
				}
			}

		default:
			continue
		}
	}
	return false, nil
}

func (d *defaultManager) SaveNamespaceConfig(ctx context.Context, config *pb.NamespaceConfig) error {
	store := namespaceConfigStore(ctx)
	if store == nil {
		return errors.Internal("could not find namespace configs store in context")
	}
	return store.SaveNamespace(config)
}

func (d *defaultManager) GetNamespaceConfig(ctx context.Context, name string) (*pb.NamespaceConfig, error) {
	store := namespaceConfigStore(ctx)
	if store == nil {
		return nil, errors.Internal("could not find namespace configs store in context")
	}
	return store.GetNamespace(name)
}

func (d *defaultManager) DeleteNamespaceConfig(ctx context.Context, name string) error {
	store := namespaceConfigStore(ctx)
	if store == nil {
		return errors.Internal("could not find namespace configs store in context")
	}
	return store.DeleteNamespace(name)
}
