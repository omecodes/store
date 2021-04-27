package acl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/omecodes/errors"
	pb "github.com/omecodes/store/gen/go/proto"
	"strings"
)

type Checker struct {
	Subject    string         `json:"subject"`
	SubjectSet *pb.SubjectSet `json:"subject_set"`
	Children   []*Checker     `json:"children"`
}

func (s *Checker) Check(ctx context.Context) (bool, error) {
	fmt.Println(s.String())

	store := getTupleStore(ctx)
	if store == nil {
		return false, errors.Internal("check acl: missing relation store in context")
	}

	nsConfigStore := getNamespaceConfigStore(ctx)
	if nsConfigStore == nil {
		return false, errors.Internal("check acl: missing namespace store in context")
	}

	subjects, err := store.GetSubjectSet(ctx, &pb.DBSubjectSetInfo{
		Object:   s.SubjectSet.Object,
		Relation: s.SubjectSet.Relation,
		MinAge:   getCommitTime(ctx),
	})
	if err != nil {
		return false, err
	}

	for _, subject := range subjects {
		if subject == s.Subject {
			return true, nil
		} else if strings.Contains(subject, "#") {
			parts := strings.Split(subject, "#")
			s.Children = append(s.Children, &Checker{
				Subject: s.Subject,
				SubjectSet: &pb.SubjectSet{
					Object:   parts[0],
					Relation: parts[1],
				},
			})
		}
	}

	objectParts := strings.Split(s.SubjectSet.Object, ":")
	namespaceId := strings.Trim(objectParts[0], " ")

	namespace, err := nsConfigStore.GetNamespace(namespaceId)
	if err != nil {
		return false, errors.Internal("resolve user-set: failed to load namespace config", errors.Details{
			Key:   "namespace-id",
			Value: namespaceId,
		}, errors.Details{
			Key:   "object",
			Value: s.SubjectSet.Object,
		})
	}

	rel, exists := namespace.Relations[s.SubjectSet.Relation]
	if !exists {
		return false, errors.NotFound("relation does not exists")
	}

	for _, rewrite := range rel.SubjectSetRewrite {
		switch rewrite.Type {

		case pb.SubjectSetType_Computed:
			checker := &Checker{
				Subject: s.Subject,
				SubjectSet: &pb.SubjectSet{
					Relation: rewrite.Value,
					Object:   s.SubjectSet.Object,
				},
			}
			s.Children = append(s.Children, checker)

		case pb.SubjectSetType_FromTuple:
			var definition *pb.SubjectsInRelationWithObjectRelatedObject
			err = json.Unmarshal([]byte(rewrite.Value), &definition)
			if err != nil {
				return false, err
			}

			var tupleSetSubjects []string
			tupleSetSubjects, err = store.GetSubjectSet(ctx, &pb.DBSubjectSetInfo{
				Relation: definition.ObjectRelation,
				Object:   s.SubjectSet.Object,
				MinAge:   getCommitTime(ctx),
			})
			if err != nil {
				return false, err
			}

			for _, subject := range tupleSetSubjects {
				s.Children = append(s.Children, &Checker{
					Subject: s.Subject,
					SubjectSet: &pb.SubjectSet{
						Object:   subject,
						Relation: definition.SubjectRelation,
					},
				})
			}
		}
	}

	var checked bool
	for _, child := range s.Children {
		checked, err = child.Check(ctx)
		if err != nil {
			return false, err
		}

		if checked {
			return true, nil
		}
	}
	return false, nil
}

func (s *Checker) String() string {
	return fmt.Sprintf("checking if %s has relationship %s with %s", s.Subject, s.SubjectSet.Relation, s.SubjectSet.Object)
}
