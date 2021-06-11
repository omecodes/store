package files

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/auth"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"
	"net/url"
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

func (h *ACLHandler) checkACL(ctx context.Context, authorizedUsers *pb.FileActionAuthorizedUsers, objectID string) error {
	user := auth.Get(ctx)
	if user != nil && user.Name == "admin" {
		return nil
	}

	if !authorizedUsers.Restricted {
		return nil
	}

	if authorizedUsers.Object != "" {
		objectID = authorizedUsers.Object
	}

	if user == nil {
		return errors.Forbidden("resource allowed only to authenticated users")
	}

	allowed, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{
		Object:   objectID,
		Relation: authorizedUsers.Relation,
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
		return errors.Forbidden("application is not allowed to create accessDB")
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

		err = h.checkACL(ctx, referencedAccess.ActionPermissions.Share, referencedAccess.Id)
		if err != nil {
			return err
		}

	} else {
		if !clientApp.AdminApp {
			return errors.Forbidden("creating this type of access requires client apps with admin flags")
		}

		user := auth.Get(ctx)
		if user == nil {
			return errors.Forbidden("Only authenticated users are allowed to create access")
		}

		success, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{Object: "group:admins", Relation: "member"}, acl.CheckACLOptions{})
		if err != nil {
			return err
		}

		if !success {
			return errors.Forbidden("only admins are allowed to create this type of access")
		}
	}

	return h.next.CreateAccess(ctx, access, opts)
}

func (h *ACLHandler) GetAccessList(ctx context.Context, opts GetAccessListOptions) ([]*pb.FSAccess, error) {
	if !auth.IsContextFromAuthorizedApp(ctx) {
		return nil, errors.Forbidden("application is not allowed to list accessDB")
	}

	var (
		ids []string
		err error
	)

	user := auth.Get(ctx)
	if user == nil {
		ids, err = acl.GetObjectNames(ctx, &pb.ObjectSet{Relation: "viewer", Subject: "public"}, acl.GetObjectsSetOptions{})
	} else {
		ids, err = acl.GetObjectNames(ctx, &pb.ObjectSet{Relation: "viewer", Subject: user.Name}, acl.GetObjectsSetOptions{})
	}
	if err != nil {
		return nil, err
	}

	var accesses []*pb.FSAccess

	for _, id := range ids {
		access, err := h.BaseHandler.GetAccess(ctx, id, GetAccessOptions{})
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

	var (
		checked bool
		err     error
	)

	user := auth.Get(ctx)
	if user == nil {
		checked, err = acl.CheckACL(ctx, "public", &pb.SubjectSet{Relation: "viewer", Object: "access:" + accessID}, acl.CheckACLOptions{})
	} else {
		checked, err = acl.CheckACL(ctx, user.Name, &pb.SubjectSet{Relation: "viewer", Object: "access:" + accessID}, acl.CheckACLOptions{})
	}
	if err != nil {
		return nil, err
	}

	if !checked {
		return nil, errors.Forbidden("this access is not allowed")
	}

	return h.next.GetAccess(ctx, accessID, opts)
}

func (h *ACLHandler) DeleteAccess(ctx context.Context, accessID string, opts DeleteAccessOptions) error {
	clientApp := auth.App(ctx)
	if clientApp == nil {
		return errors.Forbidden("application is not allowed to create accessDB")
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
			Object:   "access:" + accessID,
			Relation: "owner",
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

		checked, err := acl.CheckACL(ctx, user.Name, &pb.SubjectSet{Object: "group:admins", Relation: "member"}, acl.CheckACLOptions{})
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

	err = h.checkACL(ctx, access.ActionPermissions.Edit, accessID)
	if err != nil {
		return err
	}

	return h.next.CreateDir(ctx, accessID, dirname, opts)
}

func (h *ACLHandler) WriteFileContent(ctx context.Context, accessID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return err
	}

	err = h.checkACL(ctx, access.ActionPermissions.Edit, accessID)
	if err != nil {
		return err
	}

	err = h.next.WriteFileContent(ctx, accessID, filename, content, size, opts)
	return err
}

func (h *ACLHandler) ListDir(ctx context.Context, accessID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return nil, err
	}

	err = h.checkACL(ctx, access.ActionPermissions.View, accessID)
	if err != nil {
		return nil, err
	}
	return h.next.ListDir(ctx, accessID, dirname, opts)
}

func (h *ACLHandler) ReadFileContent(ctx context.Context, accessID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return nil, 0, err
	}

	err = h.checkACL(ctx, access.ActionPermissions.View, accessID)
	if err != nil {
		return nil, 0, err
	}
	return h.next.ReadFileContent(ctx, accessID, filename, opts)
}

func (h *ACLHandler) GetFileInfo(ctx context.Context, accessID string, filename string, opts GetFileOptions) (*pb.File, error) {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return nil, err
	}

	err = h.checkACL(ctx, access.ActionPermissions.View, accessID)
	if err != nil {
		return nil, err
	}

	return h.next.GetFileInfo(ctx, accessID, filename, opts)
}

func (h *ACLHandler) DeleteFile(ctx context.Context, accessID string, filename string, opts DeleteFileOptions) error {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return err
	}

	err = h.checkACL(ctx, access.ActionPermissions.Delete, accessID)
	if err != nil {
		return err
	}

	return h.next.DeleteFile(ctx, accessID, filename, opts)
}

func (h *ACLHandler) SetFileAttributes(ctx context.Context, accessID string, filename string, attrs Attributes, opts SetFileAttributesOptions) error {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return err
	}

	err = h.checkACL(ctx, access.ActionPermissions.Edit, accessID)
	if err != nil {
		return err
	}
	return h.next.SetFileAttributes(ctx, accessID, filename, attrs, opts)
}

func (h *ACLHandler) GetFileAttributes(ctx context.Context, accessID string, filename string, names []string, opts GetFileAttributesOptions) (Attributes, error) {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return nil, err
	}

	err = h.checkACL(ctx, access.ActionPermissions.View, accessID)
	if err != nil {
		return nil, err
	}
	return h.next.GetFileAttributes(ctx, accessID, filename, names, opts)
}

func (h *ACLHandler) RenameFile(ctx context.Context, accessID string, filename string, newName string, opts RenameFileOptions) error {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return err
	}

	err = h.checkACL(ctx, access.ActionPermissions.View, accessID)
	if err != nil {
		return err
	}

	return h.next.RenameFile(ctx, accessID, filename, newName, opts)
}

func (h *ACLHandler) MoveFile(ctx context.Context, accessID string, filename string, dirname string, opts MoveFileOptions) error {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return err
	}

	err = h.checkACL(ctx, access.ActionPermissions.View, accessID)
	if err != nil {
		return err
	}

	return h.next.MoveFile(ctx, filename, accessID, dirname, opts)
}

func (h *ACLHandler) CopyFile(ctx context.Context, accessID string, filename string, dirname string, opts CopyFileOptions) error {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return err
	}

	err = h.checkACL(ctx, access.ActionPermissions.View, accessID)
	if err != nil {
		return err
	}

	return h.next.CopyFile(ctx, accessID, filename, dirname, opts)
}

func (h *ACLHandler) OpenMultipartSession(ctx context.Context, accessID string, filename string, info MultipartSessionInfo, opts OpenMultipartSessionOptions) (string, error) {
	access, err := h.next.GetAccess(ctx, accessID, GetAccessOptions{Resolved: true})
	if err != nil {
		logs.Error("could not get access details", logs.Err(err))
		return "", err
	}

	err = h.checkACL(ctx, access.ActionPermissions.View, accessID)
	if err != nil {
		return "", err
	}
	return h.next.OpenMultipartSession(ctx, accessID, filename, info, opts)
}

func (h *ACLHandler) WriteFilePart(ctx context.Context, accessID string, content io.Reader, size int64, info ContentPartInfo, opts WriteFilePartOptions) (int64, error) {
	panic("implement me")
}

func (h *ACLHandler) CloseMultipartSession(ctx context.Context, sessionId string, opts CloseMultipartSessionOptions) error {
	panic("implement me")
}
