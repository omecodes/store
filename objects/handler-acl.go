package objects

import (
	"context"
	"fmt"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common"
	pb "github.com/omecodes/store/gen/go/proto"
)

type ACLHandler struct {
	BaseHandler
}

func (p *ACLHandler) assertUserIsAdmin(ctx context.Context) error {
	user := auth.Get(ctx)
	if user == nil {
		return errors.Forbidden("no authenticated user")
	}

	checked, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{
		Object:   common.GroupAdmins,
		Relation: common.RelationMember,
	}, acl.CheckACLOptions{})

	if err != nil {
		logs.Error("Check ACL", logs.Err(err))
		return errors.Internal("could not check ACL")
	}

	if !checked {
		return errors.Forbidden("user is not among admins")
	}

	return nil
}

func (p *ACLHandler) checkObjectReadable(ctx context.Context, collection string, objectID string, at string) error {
	header, err := p.next.GetObjectHeader(ctx, collection, objectID, GetHeaderOptions{})
	if err != nil {
		return err
	}

	collectionInfo, err := p.next.GetCollection(ctx, collection, GetCollectionOptions{})
	if err != nil {
		return err
	}

	var username string
	user := auth.Get(ctx)
	if user != nil {
		username = user.Name
	}

	if header.ActionAuthorizedUsersForPaths == nil {
		header.ActionAuthorizedUsersForPaths = collectionInfo.ActionAuthorizedUsers.AccessRules
	}

	var action *pb.ObjectActionsUsers
	if at != "" {
		action = header.ActionAuthorizedUsersForPaths[at]
	}

	if action == nil {
		logs.Info("acl info are not in object header")
		action = header.ActionAuthorizedUsersForPaths["$"]
	}

	if action.View.Object == "" {
		action.View.Object = fmt.Sprintf("%s:%s", collectionInfo.AclConfig.Namespace, objectID)
	}

	logs.Info("ACL check:", logs.Details("user", username), logs.Details("set", action.View))

	checked, err := acl.CheckACL(ctx, username, action.View, acl.CheckACLOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logs.Error("Check ACL", logs.Err(err))
		return err
	}

	if !checked {
		logs.Info("ACL check:", logs.Details("user", username), logs.Details("set", action.View), logs.Details("result", "not checked"))
		return errors.Unauthorized("permission denied")
	}
	return nil
}

func (p *ACLHandler) checkObjectEditable(ctx context.Context, collection string, objectID string, at string) error {
	header, err := p.next.GetObjectHeader(ctx, collection, objectID, GetHeaderOptions{})
	if err != nil {
		return err
	}

	collectionInfo, err := p.next.GetCollection(ctx, collection, GetCollectionOptions{})
	if err != nil {
		return err
	}

	var username string
	user := auth.Get(ctx)
	if user != nil {
		username = user.Name
	}

	if header.ActionAuthorizedUsersForPaths == nil {
		header.ActionAuthorizedUsersForPaths = collectionInfo.ActionAuthorizedUsers.AccessRules
	}

	var action *pb.ObjectActionsUsers
	if at != "" {
		action = header.ActionAuthorizedUsersForPaths[at]
	}

	if action == nil {
		action = header.ActionAuthorizedUsersForPaths["$"]
	}

	if action.Edit.Object == "" {
		action.Edit.Object = fmt.Sprintf("%s:%s", collectionInfo.AclConfig.Namespace, objectID)
	}

	checked, err := acl.CheckACL(ctx, username, action.Edit, acl.CheckACLOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logs.Error("Check ACL", logs.Err(err))
		return err
	}

	if !checked {
		return errors.Unauthorized("permission denied")
	}
	return nil
}

func (p *ACLHandler) checkObjectDeletable(ctx context.Context, collection string, objectID string, at string) error {
	header, err := p.next.GetObjectHeader(ctx, collection, objectID, GetHeaderOptions{})
	if err != nil {
		return err
	}

	collectionInfo, err := p.next.GetCollection(ctx, collection, GetCollectionOptions{})
	if err != nil {
		return err
	}

	var username string
	user := auth.Get(ctx)
	if user != nil {
		username = user.Name
	}

	if header.ActionAuthorizedUsersForPaths == nil {
		header.ActionAuthorizedUsersForPaths = collectionInfo.ActionAuthorizedUsers.AccessRules
	}

	var action *pb.ObjectActionsUsers
	if at != "" {
		action = header.ActionAuthorizedUsersForPaths[at]
	}

	if action == nil {
		action = header.ActionAuthorizedUsersForPaths["$"]
	}

	if action.Delete.Object == "" {
		action.Delete.Object = fmt.Sprintf("%s:%s", collectionInfo.AclConfig.Namespace, objectID)
	}

	checked, err := acl.CheckACL(ctx, username, action.Delete, acl.CheckACLOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logs.Error("Check ACL", logs.Err(err))
		return err
	}

	if !checked {
		return errors.Unauthorized("permission denied")
	}
	return nil
}

func (p *ACLHandler) CreateCollection(ctx context.Context, collection *pb.Collection, opts CreateCollectionOptions) error {
	if !auth.IsAdminAppFromContext(ctx) {
		return errors.Forbidden("only admin app are allowed to create collections")
	}

	err := p.assertUserIsAdmin(ctx)
	if err != nil {
		return err
	}

	return p.BaseHandler.CreateCollection(ctx, collection, opts)
}

func (p *ACLHandler) GetCollection(ctx context.Context, id string, opts GetCollectionOptions) (*pb.Collection, error) {
	if !auth.IsContextFromAuthorizedApp(ctx) {
		return nil, errors.Forbidden("application is not allowed to read client app access database")
	}
	return p.BaseHandler.GetCollection(ctx, id, opts)
}

func (p *ACLHandler) ListCollections(ctx context.Context, opts ListCollectionOptions) ([]*pb.Collection, error) {
	if !auth.IsContextFromAuthorizedApp(ctx) {
		return nil, errors.Forbidden("application is not allowed to create accessDB")
	}
	return p.BaseHandler.ListCollections(ctx, opts)
}

func (p *ACLHandler) DeleteCollection(ctx context.Context, id string, opts DeleteCollectionOptions) error {
	if !auth.IsAdminAppFromContext(ctx) {
		return errors.Forbidden("only admin app are allowed to create collections")
	}

	err := p.assertUserIsAdmin(ctx)
	if err != nil {
		return err
	}
	return p.BaseHandler.DeleteCollection(ctx, id, opts)
}

func (p *ACLHandler) PutObject(ctx context.Context, collection string, object *pb.Object, authorizedUsers *pb.PathAccessRules, indexes []*pb.TextIndex, opts PutOptions) (string, error) {
	user := auth.Get(ctx)
	if user == nil {
		return "", errors.Forbidden("access forbidden")
	}

	collectionInfo, err := p.next.GetCollection(ctx, collection, GetCollectionOptions{})
	if err != nil {
		logs.Error("could not get collection", logs.Err(err))
		return "", err
	}

	if object.Header.ActionAuthorizedUsersForPaths == nil {
		authorizedUsers = collectionInfo.ActionAuthorizedUsers
	}

	object.Header.CreatedBy = user.Name

	id, err := p.BaseHandler.PutObject(ctx, collection, object, authorizedUsers, indexes, opts)
	if err != nil {
		return "", err
	}

	err = acl.SaveACL(ctx, &pb.ACL{
		Object:   fmt.Sprintf("%s:%s", collectionInfo.AclConfig.Namespace, id),
		Relation: collectionInfo.AclConfig.RelationWithCreated,
		Subject:  user.Name,
	}, acl.SaveACLOptions{})

	if err != nil {
		delErr := p.BaseHandler.DeleteObject(ctx, collection, id, DeleteObjectOptions{})
		if delErr != nil {
			logs.Error("could not delete new created object", logs.Details("id", id))
		}
	}
	return id, err
}

func (p *ACLHandler) GetObject(ctx context.Context, collection string, id string, opts GetObjectOptions) (*pb.Object, error) {
	err := p.checkObjectReadable(ctx, collection, id, opts.At)
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObject(ctx, collection, id, opts)
}

func (p *ACLHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts PatchOptions) error {
	err := p.checkObjectEditable(ctx, collection, patch.ObjectId, patch.At)
	if err != nil {
		return err
	}

	return p.BaseHandler.PatchObject(ctx, collection, patch, opts)
}

func (p *ACLHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, authorizedUsers *pb.PathAccessRules, opts MoveOptions) error {
	err := p.checkObjectDeletable(ctx, collection, objectID, "")
	if err != nil {
		return err
	}

	if authorizedUsers == nil {
		collectionInfo, err := p.next.GetCollection(ctx, targetCollection, GetCollectionOptions{})
		if err != nil {
			return err
		}
		authorizedUsers = collectionInfo.ActionAuthorizedUsers
	}

	return p.next.MoveObject(ctx, collection, objectID, targetCollection, authorizedUsers, opts)
}

func (p *ACLHandler) GetObjectHeader(ctx context.Context, collection string, id string, opts GetHeaderOptions) (*pb.Header, error) {
	err := p.checkObjectReadable(ctx, collection, id, "")
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObjectHeader(ctx, collection, id, opts)
}

func (p *ACLHandler) DeleteObject(ctx context.Context, collection string, id string, opts DeleteObjectOptions) error {
	err := p.checkObjectDeletable(ctx, collection, id, "")
	if err != nil {
		return err
	}
	return p.BaseHandler.DeleteObject(ctx, collection, id, opts)
}

func (p *ACLHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	var err error

	cursor, err := p.next.ListObjects(ctx, collection, opts)
	if err != nil {
		return nil, err
	}

	// todo: filter viewable objects
	cursorBrowser := cursor.GetBrowser()
	browser := BrowseFunc(func() (*pb.Object, error) {
		return cursorBrowser.Browse()
	})

	cursor.SetBrowser(browser)
	return cursor, nil
}

func (p *ACLHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery, opts SearchObjectsOptions) (*Cursor, error) {
	cursor, err := p.BaseHandler.SearchObjects(ctx, collection, query, opts)
	if err != nil {
		return nil, err
	}

	// todo: filter viewable objects
	cursorBrowser := cursor.GetBrowser()
	browser := BrowseFunc(func() (*pb.Object, error) {
		for {
			o, err := cursorBrowser.Browse()
			if err != nil {
				return nil, err
			}
			return o, nil
		}
	})

	cursor.SetBrowser(browser)
	return cursor, nil
}
