package files

import (
	"context"
	"fmt"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/auth"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"
	"net/url"
	"strings"
)

type ACLHandler struct {
	BaseHandler
}

func (h *ACLHandler) isAdmin(ctx context.Context) bool {
	user := auth.Get(ctx)
	if user == nil {
		return false
	}
	return user.Name == "admin"
}

func (h *ACLHandler) checkACL(ctx context.Context, relation string, objectID string) error {
	user := auth.Get(ctx)
	if user != nil && user.Name == "admin" {
		return nil
	}

	if user == nil {
		return errors.Forbidden("resource allowed only to authenticated users")
	}

	allowed, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{
		Object:   objectID,
		Relation: relation,
	}, acl.CheckACLOptions{})
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Unauthorized("")
	}
	return nil
}

func (h *ACLHandler) CreateAccess(ctx context.Context, access *pb.FSAccess, opts CreateAccessOptions) error {
	clientApp := auth.App(ctx)
	if clientApp == nil {
		return errors.Forbidden("application is not allowed to create access")
	}

	user := auth.Get(ctx)
	if user == nil {
		return errors.Forbidden("only authenticated users are allowed to create access")
	}

	if access.Type == pb.AccessType_Reference {
		u, err := url.Parse(access.Uri)
		if err != nil {
			return errors.BadRequest("could not parse access URI", errors.Details{Key: "err", Value: access.Uri})
		}

		referencedAccess, err := h.BaseHandler.GetAccess(ctx, u.Host, GetAccessOptions{})
		if err != nil {
			return err
		}

		err = h.checkACL(ctx, relationSharer, fmt.Sprintf("%s:%s", fileNamespace, referencedAccess.Id))
		if err != nil {
			return err
		}

	} else {
		if !clientApp.AdminApp {
			return errors.Forbidden("creating this type of access requires client apps with admin flags")
		}

		success, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{Object: "group:admins", Relation: "member"}, acl.CheckACLOptions{})
		if err != nil {
			return err
		}

		if !success {
			return errors.Forbidden("only admins are allowed to create this type of access")
		}
	}

	err := h.next.CreateAccess(ctx, access, opts)
	if err != nil {
		return err
	}

	a := &pb.ACL{
		Object:   fmt.Sprintf("%s:%s", accessNamespace, access.Id),
		Relation: relationOwner,
		Subject:  adminsGroup + "#" + relationMember,
	}
	err = acl.SaveACL(ctx, a, acl.SaveACLOptions{})
	if err != nil {
		logs.Error("could not save ACL", logs.Details("ACL", a), logs.Err(err))
		err2 := h.next.DeleteAccess(ctx, access.Id, DeleteAccessOptions{})
		if err2 != nil {
			logs.Error("could not delete created access", logs.Err(err))
		}
	}
	return err
}

func (h *ACLHandler) GetAccessList(ctx context.Context, _ GetAccessListOptions) ([]*pb.FSAccess, error) {
	if !auth.IsContextFromAuthorizedApp(ctx) {
		return nil, errors.Forbidden("application is not allowed to list accessDB")
	}

	var (
		ids []string
		err error
	)

	user := auth.Get(ctx)
	if user == nil {
		ids, err = acl.GetObjectNames(ctx, &pb.ObjectSet{Relation: relationViewer, Subject: unauthenticatedUser}, acl.GetObjectsSetOptions{})
	} else {
		ids, err = acl.GetObjectNames(ctx, &pb.ObjectSet{Relation: relationViewer, Subject: user.Name}, acl.GetObjectsSetOptions{})
	}
	if err != nil {
		return nil, err
	}

	var accesses []*pb.FSAccess

	for _, id := range ids {
		access, err := h.BaseHandler.GetAccess(ctx, strings.TrimPrefix(id, accessNamespace+":"), GetAccessOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				return nil, err
			}
			continue
		}
		accesses = append(accesses, access)
	}
	return accesses, nil
}

func (h *ACLHandler) GetAccess(ctx context.Context, accessID string, opts GetAccessOptions) (*pb.FSAccess, error) {
	if !auth.IsContextFromAuthorizedApp(ctx) {
		return nil, errors.Forbidden("application is not allowed to list accessDB")
	}

	err := h.checkACL(ctx, relationViewer, fmt.Sprintf("%s:%s", accessNamespace, accessID))
	if err != nil {
		return nil, err
	}

	return h.next.GetAccess(ctx, accessID, opts)
}

func (h *ACLHandler) DeleteAccess(ctx context.Context, accessID string, _ DeleteAccessOptions) error {
	clientApp := auth.App(ctx)
	if clientApp == nil {
		return errors.Forbidden("application is not allowed to create access")
	}

	user := auth.Get(ctx)
	if user == nil {
		return errors.Forbidden("Only authenticated users are allowed to create access")
	}

	access, err := h.BaseHandler.GetAccess(ctx, accessID, GetAccessOptions{})
	if err != nil {
		return err
	}

	if access.Type == pb.AccessType_Reference {
		checked, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{
			Object:   fmt.Sprintf("%s:%s", accessNamespace, accessID),
			Relation: relationOwner,
		}, acl.CheckACLOptions{})
		if err != nil {
			return err
		}

		if !checked {
			return errors.Forbidden("not allowed to delete this file access")
		}

	} else {
		if !clientApp.AdminApp {
			return errors.Forbidden("creating this type of access requires client apps with admin flags")
		}

		checked, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{Object: adminsGroup, Relation: relationMember}, acl.CheckACLOptions{})
		if err != nil {
			return err
		}

		if !checked {
			return errors.Forbidden("only admins are allowed to create this type of access")
		}
	}

	return h.next.DeleteAccess(ctx, accessID, DeleteAccessOptions{})
}

func (h *ACLHandler) CreateDir(ctx context.Context, accessID string, dirname string, opts CreateDirOptions) error {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return err
	}

	logs.Debug("resolved access", logs.Details("value", access))

	err = h.checkACL(ctx, relationEditor, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return err
	}

	logs.Info("allowed to create directory in access", logs.Details("access", access))

	return h.next.CreateDir(ctx, accessID, dirname, opts)
}

func (h *ACLHandler) WriteFileContent(ctx context.Context, accessID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	err := h.checkACL(ctx, relationEditor, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return err
	}

	err = h.next.WriteFileContent(ctx, accessID, filename, content, size, opts)
	return err
}

// Share checks if the requester user is a sharer of the resource for each share info passed.
// Then we want to make sure the request user has the roles he wants to share to other users
func (h *ACLHandler) Share(ctx context.Context, shares []*pb.ShareInfo, _ ShareOptions) error {
	if !auth.IsContextFromAuthorizedApp(ctx) {
		return errors.New("this action must be requested from a registered application client")
	}

	user := auth.Get(ctx)
	if user == nil {
		return errors.New("authentication is required to perform this action")
	}

	for _, share := range shares {
		// todo: use cache in order not to request acl check for already checked ACL
		allowed, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{Object: share.AccessId, Relation: relationSharer}, acl.CheckACLOptions{})
		if err != nil {
			return err
		}

		if !allowed {
			return errors.Unauthorized("only allowed access can be shared")
		}

		allowed, err = acl.CheckACL(ctx, user.Name, &pb.SubjectSet{Object: share.AccessId, Relation: share.Role}, acl.CheckACLOptions{})
		if err != nil {
			return err
		}

		if !allowed {
			return errors.Unauthorized("only allowed access can be shared")
		}

		err = acl.SaveACL(ctx, &pb.ACL{
			Object:   share.AccessId,
			Relation: share.Role,
			Subject:  share.User,
		}, acl.SaveACLOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *ACLHandler) GetShares(ctx context.Context, accessID string, _ GetSharesOptions) ([]*pb.UserRole, error) {
	if !auth.IsContextFromAuthorizedApp(ctx) {
		return nil, errors.New("this action must be requested from a registered application client")
	}

	user := auth.Get(ctx)
	if user == nil {
		return nil, errors.New("authentication is required to perform this action")
	}

	allowed, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{Object: accessID, Relation: relationSharer}, acl.CheckACLOptions{})
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, errors.Unauthorized("only allowed access can be shared")
	}

	accessList, err := acl.GetObjectACL(ctx, fmt.Sprintf("%s:%s", fileNamespace, accessID), acl.GetObjectACLOptions{})
	if err != nil {
		return nil, err
	}

	var roles []*pb.UserRole

	for _, a := range accessList {
		roles = append(roles, &pb.UserRole{
			User: a.Subject,
			Role: a.Relation,
		})
	}

	return roles, nil
}

// DeleteShares checks if the requester user is a sharer of the resource for each share info passed.
// Then we want to make sure the request user has the roles he wants to delete from other users
func (h *ACLHandler) DeleteShares(ctx context.Context, shares []*pb.ShareInfo, _ DeleteSharesOptions) error {
	if !auth.IsContextFromAuthorizedApp(ctx) {
		return errors.New("this action must be requested from a registered application client")
	}

	user := auth.Get(ctx)
	if user == nil {
		return errors.New("authentication is required to perform this action")
	}

	for _, share := range shares {
		// todo: use cache in order not to request acl check for already checked ACL
		allowed, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{Object: share.AccessId, Relation: relationSharer}, acl.CheckACLOptions{})
		if err != nil {
			return err
		}

		if !allowed {
			return errors.Unauthorized("only allowed access can be shared")
		}

		allowed, err = acl.CheckACL(ctx, user.Name, &pb.SubjectSet{Object: share.AccessId, Relation: share.Role}, acl.CheckACLOptions{})
		if err != nil {
			return err
		}

		if !allowed {
			return errors.Unauthorized("only allowed access can be shared")
		}

		err = acl.DeleteACL(ctx, &pb.ACL{
			Object:   share.AccessId,
			Relation: share.Role,
			Subject:  share.User,
		}, acl.DeleteACLOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *ACLHandler) ListDir(ctx context.Context, accessID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	err := h.checkACL(ctx, relationViewer, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return nil, err
	}
	return h.next.ListDir(ctx, accessID, dirname, opts)
}

func (h *ACLHandler) ReadFileContent(ctx context.Context, accessID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	err := h.checkACL(ctx, relationViewer, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return nil, 0, err
	}
	return h.next.ReadFileContent(ctx, accessID, filename, opts)
}

func (h *ACLHandler) GetFileInfo(ctx context.Context, accessID string, filename string, opts GetFileOptions) (*pb.File, error) {
	err := h.checkACL(ctx, relationViewer, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return nil, err
	}

	return h.next.GetFileInfo(ctx, accessID, filename, opts)
}

func (h *ACLHandler) DeleteFile(ctx context.Context, accessID string, filename string, opts DeleteFileOptions) error {
	err := h.checkACL(ctx, relationOwner, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return err
	}

	return h.next.DeleteFile(ctx, accessID, filename, opts)
}

func (h *ACLHandler) SetFileAttributes(ctx context.Context, accessID string, filename string, attrs Attributes, opts SetFileAttributesOptions) error {
	err := h.checkACL(ctx, relationEditor, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return err
	}
	return h.next.SetFileAttributes(ctx, accessID, filename, attrs, opts)
}

func (h *ACLHandler) GetFileAttributes(ctx context.Context, accessID string, filename string, names []string, opts GetFileAttributesOptions) (Attributes, error) {
	err := h.checkACL(ctx, relationViewer, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return nil, err
	}
	return h.next.GetFileAttributes(ctx, accessID, filename, names, opts)
}

func (h *ACLHandler) RenameFile(ctx context.Context, accessID string, filename string, newName string, opts RenameFileOptions) error {
	err := h.checkACL(ctx, relationViewer, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return err
	}
	return h.next.RenameFile(ctx, accessID, filename, newName, opts)
}

func (h *ACLHandler) MoveFile(ctx context.Context, accessID string, filename string, dirname string, opts MoveFileOptions) error {
	err := h.checkACL(ctx, relationEditor, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return err
	}
	return h.next.MoveFile(ctx, filename, accessID, dirname, opts)
}

func (h *ACLHandler) CopyFile(ctx context.Context, accessID string, filename string, dirname string, opts CopyFileOptions) error {
	err := h.checkACL(ctx, relationEditor, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return err
	}

	return h.next.CopyFile(ctx, accessID, filename, dirname, opts)
}

func (h *ACLHandler) OpenMultipartSession(ctx context.Context, accessID string, filename string, info MultipartSessionInfo, opts OpenMultipartSessionOptions) (string, error) {
	err := h.checkACL(ctx, relationEditor, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return "", err
	}
	return h.next.OpenMultipartSession(ctx, accessID, filename, info, opts)
}

func (h *ACLHandler) WriteFilePart(ctx context.Context, accessID string, content io.Reader, size int64, info ContentPartInfo, opts WriteFilePartOptions) (int64, error) {
	err := h.checkACL(ctx, relationEditor, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return 0, err
	}
	return h.next.WriteFilePart(ctx, accessID, content, size, info, opts)
}

func (h *ACLHandler) CloseMultipartSession(ctx context.Context, accessID string, opts CloseMultipartSessionOptions) error {
	err := h.checkACL(ctx, relationEditor, fmt.Sprintf("%s:%s", fileNamespace, accessID))
	if err != nil {
		return err
	}
	return h.next.CloseMultipartSession(ctx, accessID, opts)
}
