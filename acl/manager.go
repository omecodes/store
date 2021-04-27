package acl

import (
	"context"
	"encoding/json"
	"github.com/omecodes/errors"
	pb "github.com/omecodes/store/gen/go/proto"
	"strings"
)

/*
  ⟨tuple⟩ ::= ⟨object⟩‘#’⟨relation⟩‘@’⟨user⟩
 ⟨object⟩ ::= ⟨namespace⟩‘:’⟨object id⟩
   ⟨user⟩ ::= ⟨user id⟩ | ⟨userset⟩
⟨userset⟩ ::= ⟨object⟩‘#’⟨relation⟩
*/

type Manager interface {
	SaveACL(ctx context.Context, relation *pb.ACL) error
	DeleteACL(ctx context.Context, relation *pb.ACL) error
	CheckACL(ctx context.Context, username string, subjectSet *pb.SubjectSet) (bool, error)

	SaveNamespaceConfig(ctx context.Context, config *pb.NamespaceConfig) error
	GetNamespaceConfig(ctx context.Context, name string) (*pb.NamespaceConfig, error)
	DeleteNamespaceConfig(ctx context.Context, name string) error
}

type defaultManager struct{}

func (d *defaultManager) SaveACL(ctx context.Context, a *pb.ACL) error {
	store := getTupleStore(ctx)
	if store == nil {
		return errors.Internal("could not find relation tupleStore in context")
	}

	return store.Save(ctx, &pb.DBEntry{
		Object:     a.Object,
		Relation:   a.Relation,
		Subject:    a.Subject,
		CommitTime: getCommitTime(ctx),
	})
}

func (d *defaultManager) DeleteACL(ctx context.Context, a *pb.ACL) error {
	store := getTupleStore(ctx)
	if store == nil {
		return errors.Internal("could not find relation tupleStore in context")
	}
	return store.Delete(ctx, &pb.DBEntry{
		Sid:        0,
		Object:     a.Object,
		Relation:   a.Relation,
		Subject:    a.Subject,
		CommitTime: 0,
	})
}

func (d *defaultManager) CheckACL(ctx context.Context, username string, subjectSet *pb.SubjectSet) (bool, error) {
	checker := &Checker{
		Subject:    username,
		SubjectSet: subjectSet,
	}

	return checker.Check(ctx)
	/*store := getTupleStore(ctx)
	if store == nil {
		return false, errors.Internal("check acl: missing relation store in context")
	}

	nsConfigStore := getNamespaceConfigStore(ctx)
	if nsConfigStore == nil {
		return false, errors.Internal("check acl: missing namespace store in context")
	}

	objectParts := strings.Split(subjectSet.Object, ":")
	namespaceId := strings.Trim(objectParts[0], " ")

	namespace, err := nsConfigStore.GetNamespace(namespaceId)
	if err != nil {
		return false, errors.Internal("check acl: failed to load namespace config", errors.Details{
			Key:   "namespace-id",
			Value: namespaceId,
		})
	}

	rel, exists := namespace.Relations[subjectSet.Relation]
	if !exists {
		return false, errors.NotFound("check acl: relation does not exists", errors.Details{
			Key:   "relation",
			Value: subjectSet.Relation,
		}, errors.Details{
			Key:   "namespace",
			Value: namespace,
		})
	}

	for _, rewrite := range rel.SubjectSetRewrite {

		switch rewrite.Type {
		case pb.SubjectSetType_This:
			fmt.Printf("\ncheck if '%s' has direct '%s' relationship with '%s'\n", username, subjectSet.Relation, subjectSet.Object)
			// This relation is not referencing another relation, so we check for existing entry directly in database
			exists, err = store.Check(ctx, &pb.DBEntry{
				Object:     subjectSet.Object,
				Relation:   subjectSet.Relation,
				Subject:    username,
				CommitTime: minAge,
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

		case pb.SubjectSetType_Computed:
			fmt.Printf("check if '%s' has '%s' relationship with '%s'\n", username, rewrite.Value, subjectSet.Object)
			// This relation is referring to another relation, we then check if user has that relation with the input object
			exists, err = d.CheckACL(ctx, username, &pb.SubjectSet{
				Object:   subjectSet.Object,
				Relation: rewrite.Value,
			}, minAge)
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return false, err
			}
			if exists {
				return true, nil
			}

		case pb.SubjectSetType_FromTuple:
			// This relation is referring to a relation objects that are in relation with current object
			var subjects []string
			var definition *pb.SubjectsInRelationWithObjectRelatedObject
			err = json.Unmarshal([]byte(rewrite.Value), &definition)
			if err != nil {
				return false, err
			}

			fmt.Printf("check if '%s' has relation '%s' relationship with subject set in tuples '%s-%s'\n", username, definition.SubjectRelation, definition.ObjectRelation, subjectSet.Object)

			subjects, err = d.ResolveSubjectSet(ctx, &pb.DBSubjectSetInfo{
				Relation: definition.ObjectRelation,
				Object:   subjectSet.Object,
				MinAge:   minAge,
			})
			if err != nil {
				return false, err
			}

			fmt.Printf("subjects that are have '%s' relationship with %s are %v\n", definition.ObjectRelation, subjectSet.Object, subjects)

			for _, subject := range subjects {
				var allSubjects []string

				if strings.Contains(subject, "#") {
					parts := strings.Split(subject, "#")
					allSubjects, err = d.ResolveSubjectSet(ctx, &pb.DBSubjectSetInfo{
						Object:   parts[0],
						Relation: parts[1],
						MinAge:   minAge,
					})
					if err != nil {
						return false, err
					}
				} else {
					allSubjects = append(allSubjects, subject)
				}
				fmt.Printf("resolving subject set matching %s\n", subject)

				for _, s := range allSubjects {
					exists, err = d.CheckACL(ctx, username, &pb.SubjectSet{
						Object:   s,
						Relation: definition.SubjectRelation,
					}, minAge)
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
			}
		default:
			continue
		}
	}
	return false, nil */
}

func (d *defaultManager) ResolveSubjectSet(ctx context.Context, info *pb.DBSubjectSetInfo) ([]string, error) {
	nsConfigStore := getNamespaceConfigStore(ctx)
	if nsConfigStore == nil {
		return nil, errors.Internal("missing namespace config in context")
	}

	store := getTupleStore(ctx)
	if store == nil {
		return nil, errors.Internal("resolve user-set: missing relation tupleStore in context")
	}

	objectParts := strings.Split(info.Object, ":")
	namespaceId := strings.Trim(objectParts[0], " ")

	namespace, err := nsConfigStore.GetNamespace(namespaceId)
	if err != nil {
		return nil, errors.Internal("resolve user-set: failed to load namespace config", errors.Details{
			Key:   "namespace-id",
			Value: namespaceId,
		})
	}

	rel, exists := namespace.Relations[info.Relation]
	if !exists {
		return nil, errors.NotFound("relation does not exists")
	}

	var subjectsSet []string
	for _, rewrite := range rel.SubjectSetRewrite {
		var subjects []string

		switch rewrite.Type {
		case pb.SubjectSetType_This:
			subjects, err = store.GetSubjectSet(ctx, &pb.DBSubjectSetInfo{
				Relation: info.Relation,
				Object:   info.Object,
				MinAge:   info.MinAge,
			})
			if err != nil {
				return nil, err
			}

		case pb.SubjectSetType_Computed:
			subjects, err = d.ResolveSubjectSet(ctx, &pb.DBSubjectSetInfo{
				Relation: rewrite.Value,
				Object:   info.Object,
				MinAge:   info.MinAge,
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
			tupleSetSubjects, err = d.ResolveSubjectSet(ctx, &pb.DBSubjectSetInfo{
				Relation: definition.ObjectRelation,
				Object:   info.Object,
				MinAge:   info.MinAge,
			})
			if err != nil {
				return nil, err
			}

			for _, subject := range tupleSetSubjects {
				var allSubjects []string
				if strings.Contains(subject, "#") {
					parts := strings.Split(subject, "#")
					allSubjects, err = d.ResolveSubjectSet(ctx, &pb.DBSubjectSetInfo{
						Object:   parts[0],
						Relation: parts[1],
						MinAge:   info.MinAge,
					})
					if err != nil {
						return nil, err
					}
				} else {
					allSubjects = append(allSubjects, subject)
				}

				var resolvedUsers []string
				for _, s := range allSubjects {
					resolvedUsers, err = d.ResolveSubjectSet(ctx, &pb.DBSubjectSetInfo{
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

func (d *defaultManager) SaveNamespaceConfig(ctx context.Context, config *pb.NamespaceConfig) error {
	store := getNamespaceConfigStore(ctx)
	if store == nil {
		return errors.Internal("could not find namespace configs tuples store in context")
	}
	return store.SaveNamespace(config)
}

func (d *defaultManager) GetNamespaceConfig(ctx context.Context, name string) (*pb.NamespaceConfig, error) {
	store := getNamespaceConfigStore(ctx)
	if store == nil {
		return nil, errors.Internal("could not find namespace configs tupleStore in context")
	}
	return store.GetNamespace(name)
}

func (d *defaultManager) DeleteNamespaceConfig(ctx context.Context, name string) error {
	store := getNamespaceConfigStore(ctx)
	if store == nil {
		return errors.Internal("could not find namespace configs tupleStore in context")
	}
	return store.DeleteNamespace(name)
}
