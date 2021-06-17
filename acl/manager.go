package acl

import (
	"context"
	"encoding/json"
	"github.com/omecodes/errors"
	pb "github.com/omecodes/store/gen/go/proto"
	"strings"
)

type Manager interface {
	SaveACL(ctx context.Context, relation *pb.ACL) error
	DeleteACL(ctx context.Context, relation *pb.ACL) error
	CheckACL(ctx context.Context, username string, subjectSet *pb.SubjectSet) (bool, error)
	GetObjectACL(ctx context.Context, objectID string) ([]*pb.ACL, error)
	GetSubjectACL(ctx context.Context, subjectID string) ([]*pb.ACL, error)

	SaveNamespaceConfig(ctx context.Context, config *pb.NamespaceConfig) error
	GetNamespaceConfig(ctx context.Context, name string) (*pb.NamespaceConfig, error)
	DeleteNamespaceConfig(ctx context.Context, name string) error

	GetSubjectsNames(ctx context.Context, set *pb.SubjectSet) ([]string, error)
	GetObjectsNames(ctx context.Context, set *pb.ObjectSet) ([]string, error)
}

type DefaultManager struct{}

func (d *DefaultManager) SaveACL(ctx context.Context, a *pb.ACL) error {
	store := getTupleStore(ctx)
	if store == nil {
		return errors.Internal("could not find relation tuple store in context")
	}

	return store.Save(ctx, &pb.DBEntry{
		Object:      a.Object,
		Relation:    a.Relation,
		Subject:     a.Subject,
		StateMinAge: getStateMinAge(ctx),
	})
}

func (d *DefaultManager) DeleteACL(ctx context.Context, a *pb.ACL) error {
	store := getTupleStore(ctx)
	if store == nil {
		return errors.Internal("could not find relation tuple store in context")
	}
	return store.Delete(ctx, &pb.DBEntry{
		Sid:      0,
		Object:   a.Object,
		Relation: a.Relation,
		Subject:  a.Subject,
	})
}

func (d *DefaultManager) CheckACL(ctx context.Context, username string, subjectSet *pb.SubjectSet) (bool, error) {
	checker := &Checker{
		Subject:    username,
		SubjectSet: subjectSet,
	}
	return checker.Check(ctx)
}

func (d *DefaultManager) GetObjectACL(ctx context.Context, objectID string) ([]*pb.ACL, error) {
	store := getTupleStore(ctx)
	if store == nil {
		return nil, errors.Internal("could not find relation tuple store in context")
	}

	entries, err := store.GetForObject(ctx, objectID, 0)
	if err != nil {
		return nil, err
	}

	var list []*pb.ACL
	for _, entry := range entries {
		list = append(list, &pb.ACL{
			Object:   entry.Object,
			Relation: entry.Relation,
			Subject:  entry.Subject,
		})
	}
	return list, err
}

func (d *DefaultManager) GetSubjectACL(ctx context.Context, subjectID string) ([]*pb.ACL, error) {
	store := getTupleStore(ctx)
	if store == nil {
		return nil, errors.Internal("could not find relation tuple store in context")
	}

	entries, err := store.GetForSubject(ctx, subjectID, 0)
	if err != nil {
		return nil, err
	}

	var list []*pb.ACL
	for _, entry := range entries {
		list = append(list, &pb.ACL{
			Object:   entry.Object,
			Relation: entry.Relation,
			Subject:  entry.Subject,
		})
	}
	return list, err
}

func (d *DefaultManager) FindObjects(ctx context.Context, subject string, relation string, namespaceID string) ([]string, error) {
	return nil, nil
}

func (d *DefaultManager) FindSubjects(ctx context.Context, relation string, objectID string) ([]string, error) {
	return nil, nil
}

func (d *DefaultManager) SaveNamespaceConfig(ctx context.Context, config *pb.NamespaceConfig) error {
	store := getNamespaceConfigStore(ctx)
	if store == nil {
		return errors.Internal("could not find namespace configs tuples store in context")
	}
	return store.SaveNamespace(config)
}

func (d *DefaultManager) GetNamespaceConfig(ctx context.Context, name string) (*pb.NamespaceConfig, error) {
	store := getNamespaceConfigStore(ctx)
	if store == nil {
		return nil, errors.Internal("could not find namespace configs tuple store in context")
	}
	return store.GetNamespace(name)
}

func (d *DefaultManager) DeleteNamespaceConfig(ctx context.Context, name string) error {
	store := getNamespaceConfigStore(ctx)
	if store == nil {
		return errors.Internal("could not find namespace configs tuple store in context")
	}
	return store.DeleteNamespace(name)
}

func (d *DefaultManager) GetSubjectsNames(ctx context.Context, set *pb.SubjectSet) ([]string, error) {
	nsConfigStore := getNamespaceConfigStore(ctx)
	if nsConfigStore == nil {
		return nil, errors.Internal("acl.DefaultManager: missing namespace config in context")
	}

	store := getTupleStore(ctx)
	if store == nil {
		return nil, errors.Internal("resolve user-set: missing relation tuple store in context")
	}

	objectParts := strings.Split(set.Object, ":")
	namespaceId := strings.Trim(objectParts[0], " ")

	namespace, err := nsConfigStore.GetNamespace(namespaceId)
	if err != nil {
		return nil, errors.Internal("resolve user-set: failed to load namespace config", errors.Details{
			Key:   "namespace-id",
			Value: namespaceId,
		})
	}

	rel, exists := namespace.Relations[set.Relation]
	if !exists {
		return nil, errors.NotFound("relation does not exists")
	}

	var subjectsSet []string
	for _, rewrite := range rel.SubjectSetRewrite {
		var subjects []string

		switch rewrite.Type {
		case pb.SubjectSetType_This:
			subjects, err = store.GetSubjects(ctx, &pb.DBSubjectSetInfo{
				Relation:    set.Relation,
				Object:      set.Object,
				StateMinAge: getStateMinAge(ctx),
			})
			if err != nil {
				return nil, err
			}

		case pb.SubjectSetType_Computed:
			subjects, err = d.GetSubjectsNames(ctx, &pb.SubjectSet{
				Relation: rewrite.Value,
				Object:   set.Object,
			})
			if err != nil {
				return nil, err
			}

		case pb.SubjectSetType_FromTuple:
			var definition *pb.SubjectsInRelationWithObjectRelatedObject
			err = json.Unmarshal([]byte(rewrite.Value), &definition)
			if err != nil {
				return nil, err
			}

			var tupleSetSubjects []string
			tupleSetSubjects, err = d.GetSubjectsNames(ctx, &pb.SubjectSet{
				Relation: definition.ObjectRelation,
				Object:   set.Object,
			})
			if err != nil {
				return nil, err
			}

			for _, subject := range tupleSetSubjects {
				var allSubjects []string
				if strings.Contains(subject, "#") {
					parts := strings.Split(subject, "#")
					allSubjects, err = d.GetSubjectsNames(ctx, &pb.SubjectSet{
						Object:   parts[0],
						Relation: parts[1],
					})
					if err != nil {
						return nil, err
					}
				} else {
					allSubjects = append(allSubjects, subject)
				}

				var resolvedUsers []string
				for _, s := range allSubjects {
					resolvedUsers, err = d.GetSubjectsNames(ctx, &pb.SubjectSet{
						Object:   s,
						Relation: definition.SubjectRelation,
					})
					if err != nil {
						return nil, err
					}
					subjects = append(subjects, resolvedUsers...)
				}
			}

		default:
			continue
		}
		subjectsSet = append(subjectsSet, subjects...)
	}
	return subjectsSet, nil
}

func (d *DefaultManager) GetObjectsNames(ctx context.Context, set *pb.ObjectSet) ([]string, error) {
	store := getTupleStore(ctx)
	if store == nil {
		return nil, errors.Internal("resolve user-set: missing relation tuple store in context")
	}
	return store.GetObjects(ctx, &pb.DBObjectSetInfo{
		Subject:     set.Subject,
		Relation:    set.Relation,
		StateMinAge: getStateMinAge(ctx),
	})
}
