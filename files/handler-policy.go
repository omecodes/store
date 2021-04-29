package files

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/auth"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"
	"path"
)

type PolicyHandler struct {
	BaseHandler
}

func (h *PolicyHandler) isAdmin(ctx context.Context) bool {
	user := auth.Get(ctx)
	if user == nil {
		return false
	}
	return user.Name == "admin"
}

func (h *PolicyHandler) checkACL(ctx context.Context, action *pb.ActionAuthorization, fileID string) error {
	user := auth.Get(ctx)
	if user != nil && user.Name == "admin" {
		return nil
	}
	am := acl.GetManager(ctx)
	if am == nil {
		return errors.Internal("missing ACL manager in context")
	}

	if !action.Restricted {
		return nil
	}

	if user == nil {
		return errors.Forbidden("resource allowed only to authenticated users")
	}

	allowed, err := am.CheckACL(ctx, user.Name, &pb.SubjectSet{
		Object:   fileID,
		Relation: action.Relation,
	})
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Unauthorized("")
	}
	return nil
}

func (h *PolicyHandler) assertIsAllowedToRead(ctx context.Context, accessID string, fileID string) error {
	access, err := h.next.GetAccess(ctx, accessID)
	if access == nil {
		return err
	}

	if access.OperationRelationOverride != nil {
		return h.checkACL(ctx, access.OperationRelationOverride.Own, accessID)
	}

	attrs, err := h.next.GetFileAttributes(ctx, accessID, fileID, AttrPermissions)
	if err != nil {
		return err
	}

	attrsHolder := HoldAttributes(attrs)
	authorization, found, err := attrsHolder.GetPermissions()
	if err != nil {
		return err
	}

	if !found {
		return errors.Forbidden("access to this resources is forbidden")
	}

	return h.checkACL(ctx, authorization.Own, fileID)
}

func (h *PolicyHandler) assertIsAllowedToWrite(ctx context.Context, accessID string, fileID string) error {
	access, err := h.next.GetAccess(ctx, accessID)
	if access == nil {
		return err
	}

	if access.OperationRelationOverride != nil {
		return h.checkACL(ctx, access.OperationRelationOverride.Edit, accessID)
	}

	attrs, err := h.next.GetFileAttributes(ctx, accessID, fileID, AttrPermissions)
	if err != nil {
		return err
	}

	attrsHolder := HoldAttributes(attrs)
	authorization, found, err := attrsHolder.GetPermissions()
	if err != nil {
		return err
	}

	if !found {
		return errors.Forbidden("access to this resources is forbidden")
	}

	return h.checkACL(ctx, authorization.Edit, fileID)
}

func (h *PolicyHandler) assertIsAllowedToChmod(ctx context.Context, accessID string, fileID string) error {
	access, err := h.next.GetAccess(ctx, accessID)
	if access == nil {
		return err
	}

	if access.OperationRelationOverride != nil {
		return h.checkACL(ctx, access.OperationRelationOverride.Own, accessID)
	}

	attrs, err := h.next.GetFileAttributes(ctx, accessID, fileID, AttrPermissions)
	if err != nil {
		return err
	}

	attrsHolder := HoldAttributes(attrs)
	authorization, found, err := attrsHolder.GetPermissions()
	if err != nil {
		return err
	}

	if !found {
		return errors.Forbidden("access to this resources is forbidden")
	}

	return h.checkACL(ctx, authorization.Own, fileID)
}

func (h *PolicyHandler) assertAllowedToChmodSource(ctx context.Context, source *pb.Access) error {
	/*user := auth.Get(ctx)
	if user == nil {
		return errors.Forbidden("only authenticated users are allowed to perform this action")
	}

	if user.Name == "admin" {
		return nil
	}

	sourceChain := []string{source.Id}
	sourceType := source.Type
	var refSourceID string

	for sourceType == pb.AccessType_Default {
		u, err := url.Parse(source.Uri)
		if err != nil {
			return errors.Internal("could not resolve source uri", errors.Details{Key: "uri", Value: err})
		}

		if u.Scheme != "ref" {
			return errors.Internal("unexpected source scheme")
		}

		refSourceID = u.Host
		refSource, err := h.next.GetAccess(ctx, refSourceID)
		if err != nil {
			return err
		}

		if refSource.PermissionOverrides != nil && len(refSource.PermissionOverrides.Chmod) > 0 {
			var rules []string
			for _, wr := range refSource.PermissionOverrides.Chmod {
				rules = append(rules, wr.Rule)
			}
			return h.assertPermissionIsGranted(ctx, rules...)
		}

		sourceType = refSource.Type

		for _, src := range sourceChain {
			if src == refSourceID {
				return errors.Internal("source cycle references")
			}
		}
		sourceChain = append(sourceChain, refSourceID)
		sourceType = source.Type
	}

	attrs, err := h.next.GetFileAttributes(ctx, refSourceID, "/", AttrPermissions)
	if err != nil {
		return err
	}

	attrsHolder := HoldAttributes(attrs)
	perms, found, err := attrsHolder.GetPermissions()
	if err != nil {
		return err
	}

	if !found {
		return errors.Forbidden("access to this resources is forbidden")
	}

	var rules []string
	for _, perm := range perms.Chmod {
		rules = append(rules, perm.Rule)
	}

	return h.assertPermissionIsGranted(ctx, rules...) */
	return nil
}

func (h *PolicyHandler) CreateSource(ctx context.Context, source *pb.Access) error {
	clientApp := auth.App(ctx)
	if clientApp == nil {
		return errors.Forbidden("application is not allowed to create accessDB")
	}

	err := h.assertAllowedToChmodSource(ctx, source)
	if err != nil {
		return err
	}
	return h.next.CreateSource(ctx, source)
}

func (h *PolicyHandler) GetAccessList(ctx context.Context) ([]*pb.Access, error) {
	clientApp := auth.App(ctx)
	if clientApp == nil {
		return nil, errors.Forbidden("application is not allowed to list accessDB")
	}

	sources, err := h.next.GetAccessList(ctx)
	if err != nil {
		return nil, err
	}

	var allowedSources []*pb.Access
	for _, source := range sources {
		err = h.assertIsAllowedToRead(ctx, source.Id, "/")
		if err != nil {
			continue
		}
		allowedSources = append(allowedSources, source)
	}
	return allowedSources, nil
}

func (h *PolicyHandler) GetAccess(ctx context.Context, sourceID string) (*pb.Access, error) {
	clientApp := auth.App(ctx)
	if clientApp == nil {
		return nil, errors.Forbidden("application is not allowed to list accessDB")
	}

	err := h.assertIsAllowedToRead(ctx, sourceID, "/")
	if err != nil {
		return nil, err
	}
	return h.next.GetAccess(ctx, sourceID)
}

func (h *PolicyHandler) DeleteAccess(ctx context.Context, sourceID string) error {
	user := auth.Get(ctx)
	if user == nil {
		return errors.Forbidden("context missing user")
	}

	clientApp := auth.App(ctx)
	if clientApp == nil {
		return errors.Forbidden("application is not allowed to delete accessDB")
	}

	source, err := h.next.GetAccess(ctx, sourceID)
	if err != nil {
		return err
	}

	if user.Name != "admin" {
		if source.CreatedBy != user.Name {
			return errors.Forbidden("context missing user")
		}
	}
	return h.next.DeleteAccess(ctx, sourceID)
}

func (h *PolicyHandler) CreateDir(ctx context.Context, sourceID string, filename string) error {
	err := h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return err
	}
	return h.next.CreateDir(ctx, sourceID, filename)
}

func (h *PolicyHandler) WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	err := h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return err
	}
	err = h.next.WriteFileContent(ctx, sourceID, filename, content, size, opts)
	return err
}

func (h *PolicyHandler) ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	err := h.assertIsAllowedToRead(ctx, sourceID, dirname)
	if err != nil {
		return nil, err
	}
	return h.next.ListDir(ctx, sourceID, dirname, opts)
}

func (h *PolicyHandler) ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	err := h.assertIsAllowedToRead(ctx, sourceID, filename)
	if err != nil {
		return nil, 0, err
	}
	return h.next.ReadFileContent(ctx, sourceID, filename, opts)
}

func (h *PolicyHandler) GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileOptions) (*pb.File, error) {
	err := h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return nil, err
	}
	return h.next.GetFileInfo(ctx, sourceID, filename, opts)
}

func (h *PolicyHandler) DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error {
	err := h.assertIsAllowedToWrite(ctx, sourceID, filename)
	if err != nil {
		return err
	}

	return h.next.DeleteFile(ctx, sourceID, filename, opts)
}

func (h *PolicyHandler) SetFileAttributes(ctx context.Context, sourceID string, filename string, attrs Attributes) error {
	err := h.assertIsAllowedToWrite(ctx, sourceID, filename)
	if err != nil {
		return err
	}
	return h.next.SetFileAttributes(ctx, sourceID, filename, attrs)
}

func (h *PolicyHandler) GetFileAttributes(ctx context.Context, sourceID string, filename string, name ...string) (Attributes, error) {
	err := h.assertIsAllowedToRead(ctx, sourceID, filename)
	if err != nil {
		return nil, err
	}
	return h.next.GetFileAttributes(ctx, sourceID, filename, name...)
}

func (h *PolicyHandler) RenameFile(ctx context.Context, sourceID string, filename string, newName string) error {
	err := h.assertIsAllowedToRead(ctx, sourceID, filename)
	if err != nil {
		return err
	}

	err = h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return err
	}

	return h.next.RenameFile(ctx, sourceID, filename, newName)
}

func (h *PolicyHandler) MoveFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	err := h.assertIsAllowedToWrite(ctx, sourceID, filename)
	if err != nil {
		return err
	}

	err = h.assertIsAllowedToWrite(ctx, sourceID, dirname)
	if err != nil {
		return err
	}

	return h.next.MoveFile(ctx, filename, sourceID, dirname)
}

func (h *PolicyHandler) CopyFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	err := h.assertIsAllowedToRead(ctx, sourceID, filename)
	if err != nil {
		return err
	}

	err = h.assertIsAllowedToWrite(ctx, sourceID, dirname)
	if err != nil {
		return err
	}

	return h.next.CopyFile(ctx, sourceID, filename, dirname)
}

func (h *PolicyHandler) OpenMultipartSession(ctx context.Context, sourceID string, filename string, info MultipartSessionInfo) (string, error) {
	err := h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return "", err
	}
	return h.next.OpenMultipartSession(ctx, sourceID, filename, info)
}

func (h *PolicyHandler) WriteFilePart(ctx context.Context, sessionID string, content io.Reader, size int64, info ContentPartInfo) (int64, error) {
	panic("implement me")
}

func (h *PolicyHandler) CloseMultipartSession(_ context.Context, _ string) error {
	panic("implement me")
}
