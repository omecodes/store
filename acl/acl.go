//Package acl provides API to manage ACL

package acl

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
)

// SaveNamespaceConfig saves the given namespace config to the store resolved from the given context

// GetNamespaceConfig gets namespace config from the store resolved from context

// DeleteNamespaceConfig deletes namespace config that match id from the store resolved from the given context

// SaveACL saves an ACL in the relations store resolved from the given context
func SaveACL(ctx context.Context, a *pb.ACL, opts SaveACLOptions) error {
	handler := GetHandler(ctx)
	return handler.SaveACL(ctx, a, opts)
}

// DeleteACL deletes the passed ACL from the relations store resolved from the given context
func DeleteACL(ctx context.Context, a *pb.ACL, opts DeleteACLOptions) error {
	handler := GetHandler(ctx)
	return handler.DeleteACL(ctx, a, opts)
}

// CheckACL checks if the given subject is an element of the given subject set
func CheckACL(ctx context.Context, subjectName string, set *pb.SubjectSet, opts CheckACLOptions) (bool, error) {
	handler := GetHandler(ctx)
	return handler.CheckACL(ctx, subjectName, set, opts)
}

// GetObjectACL retrieves all acl related to the passed objectID
func GetObjectACL(ctx context.Context, objectID string, opts GetObjectACLOptions) ([]*pb.ACL, error) {
	handler := GetHandler(ctx)
	return handler.GetObjectACL(ctx, objectID, opts)
}

// GetSubjectACL retrieves all acl related to the passed subjectID

// GetSubjectNames gets the names of the subjects from the store resolved from the given context, that are elements of the given subject set
// set is defined as the combination of an object and a relation

// GetObjectNames gets the names of the objects from the store resolved from the given context, that are elements of the given object set
// set is defined as the combination of a subject and a relation
func GetObjectNames(ctx context.Context, set *pb.ObjectSet, opts GetObjectsSetOptions) ([]string, error) {
	handler := GetHandler(ctx)
	return handler.GetObjectNames(ctx, set, opts)
}
