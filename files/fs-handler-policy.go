package files

import (
	"context"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common/cenv"
)

type PolicyHandler struct {
	BaseHandler
}

func (h *PolicyHandler) assertPermissionIsGranted(ctx context.Context, rules ...string) error {
	var formattedRules []string
	for _, exp := range rules {
		if exp == "true" {
			return nil
		}
		formattedRules = append(formattedRules, "("+exp+")")
	}
	fullExpression := strings.Join(formattedRules, " || ")

	prg, err := cenv.GetProgram(fullExpression)
	if err != nil {
		return errors.Create(errors.Internal, "context missing access rule evaluator")
	}

	vars := map[string]interface{}{}
	user := auth.Get(ctx)
	if user != nil {
		vars["user"] = map[string]interface{}{
			"name":   user.Name,
			"access": user.Access,
			"group":  user.Group,
		}
	}

	out, details, err := prg.Eval(vars)
	if err != nil {
		logs.Error("file permission evaluation", logs.Details("details", details))
		return err
	}

	if out.Value().(bool) {
		return nil
	}

	return errors.Create(errors.Forbidden, "permission denied")
}

func (h *PolicyHandler) assertIsAllowedToRead(ctx context.Context, sourceID string, filename string) error {
	user := auth.Get(ctx)
	if user != nil && user.Name == "admin" {
		return nil
	}

	source, err := h.next.GetSource(ctx, sourceID)
	if source == nil {
		return err
	}

	if source.PermissionOverrides != nil && len(source.PermissionOverrides.Read) > 0 {
		var rules []string
		for _, wr := range source.PermissionOverrides.Read {
			rules = append(rules, wr.Rule)
		}
		return h.assertPermissionIsGranted(ctx, rules...)
	}

	attrs, err := h.next.GetFileAttributes(ctx, sourceID, filename, AttrPermissions)
	if err != nil {
		return err
	}

	attrsHolder := HoldAttributes(attrs)
	perms, found, err := attrsHolder.GetPermissions()
	if err != nil {
		return err
	}

	if !found {
		return errors.Create(errors.Forbidden, "access to this resources is forbidden")
	}

	var rules []string
	for _, perm := range perms.Read {
		rules = append(rules, perm.Rule)
	}

	return h.assertPermissionIsGranted(ctx, rules...)
}

func (h *PolicyHandler) assertIsAllowedToWrite(ctx context.Context, sourceID string, filename string) error {
	user := auth.Get(ctx)
	if user != nil && user.Name == "admin" {
		return nil
	}

	source, err := h.next.GetSource(ctx, sourceID)
	if source == nil {
		return err
	}

	if source.PermissionOverrides != nil && len(source.PermissionOverrides.Write) > 0 {
		var rules []string
		for _, wr := range source.PermissionOverrides.Write {
			rules = append(rules, wr.Rule)
		}
		return h.assertPermissionIsGranted(ctx, rules...)
	}

	attrs, err := h.next.GetFileAttributes(ctx, sourceID, filename, AttrPermissions)
	if err != nil {
		return err
	}

	attrsHolder := HoldAttributes(attrs)
	perms, found, err := attrsHolder.GetPermissions()
	if err != nil {
		return err
	}

	if !found {
		return errors.Create(errors.Forbidden, "access to this resources is forbidden")
	}

	var rules []string
	for _, perm := range perms.Write {
		rules = append(rules, perm.Rule)
	}

	return h.assertPermissionIsGranted(ctx, rules...)
}

func (h *PolicyHandler) assertIsAllowedToChmod(ctx context.Context, sourceID string, filename string) error {
	user := auth.Get(ctx)
	if user != nil && user.Name == "admin" {
		return nil
	}

	source, err := h.next.GetSource(ctx, sourceID)
	if source == nil {
		return err
	}

	if source.PermissionOverrides != nil && len(source.PermissionOverrides.Chmod) > 0 {
		var rules []string
		for _, wr := range source.PermissionOverrides.Chmod {
			rules = append(rules, wr.Rule)
		}
		return h.assertPermissionIsGranted(ctx, rules...)
	}

	attrs, err := h.next.GetFileAttributes(ctx, sourceID, filename, AttrPermissions)
	if err != nil {
		return err
	}

	attrsHolder := HoldAttributes(attrs)
	perms, found, err := attrsHolder.GetPermissions()
	if err != nil {
		return err
	}

	if !found {
		return errors.Create(errors.Forbidden, "access to this resources is forbidden")
	}

	var rules []string
	for _, perm := range perms.Chmod {
		rules = append(rules, perm.Rule)
	}

	return h.assertPermissionIsGranted(ctx, rules...)
}

func (h *PolicyHandler) assertAllowedToChmodSource(ctx context.Context, source *Source) error {
	user := auth.Get(ctx)
	if user == nil {
		return errors.Create(errors.Forbidden, "")
	}

	if user.Name == "admin" {
		return nil
	}

	sourceChain := []string{source.ID}
	sourceType := source.Type
	var refSourceID string

	for sourceType == TypeReference {
		u, err := url.Parse(source.URI)
		if err != nil {
			return errors.Create(errors.Internal, "could not resolve source uri", errors.Info{Name: "uri", Details: err.Error()})
		}

		if u.Scheme != "ref" {
			return errors.Create(errors.Internal, "unexpected source scheme")
		}

		refSourceID = u.Host
		refSource, err := h.next.GetSource(ctx, refSourceID)
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
				return errors.Create(errors.Internal, "source cycle references")
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
		return errors.Create(errors.Forbidden, "access to this resources is forbidden")
	}

	var rules []string
	for _, perm := range perms.Chmod {
		rules = append(rules, perm.Rule)
	}

	return h.assertPermissionIsGranted(ctx, rules...)
}

func (h *PolicyHandler) CreateSource(ctx context.Context, source *Source) error {
	err := h.assertAllowedToChmodSource(ctx, source)
	if err != nil {
		return err
	}
	return h.next.CreateSource(ctx, source)
}

func (h *PolicyHandler) ListSources(ctx context.Context) ([]*Source, error) {
	sources, err := h.next.ListSources(ctx)
	if err != nil {
		return nil, err
	}

	var allowedSources []*Source
	for _, source := range sources {
		err = h.assertIsAllowedToRead(ctx, source.ID, "/")
		if err != nil {
			continue
		}
		allowedSources = append(allowedSources, source)
	}
	return allowedSources, nil
}

func (h *PolicyHandler) GetSource(ctx context.Context, sourceID string) (*Source, error) {
	err := h.assertIsAllowedToRead(ctx, sourceID, "/")
	if err != nil {
		return nil, err
	}
	return h.next.GetSource(ctx, sourceID)
}

func (h *PolicyHandler) DeleteSource(ctx context.Context, sourceID string) error {
	user := auth.Get(ctx)
	if user == nil {
		return errors.Create(errors.Forbidden, "context missing user")
	}

	if user.Name != "admin" {
		source, err := h.next.GetSource(ctx, sourceID)
		if err != nil {
			return err
		}
		if source.CreatedBy != user.Name {
			return errors.Create(errors.Forbidden, "context missing user")
		}
	}
	return h.next.DeleteSource(ctx, sourceID)
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
	return h.next.WriteFileContent(ctx, sourceID, filename, content, size, opts)
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

func (h *PolicyHandler) GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileInfoOptions) (*File, error) {
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

func (h *PolicyHandler) SetFileMetaData(ctx context.Context, sourceID string, filename string, attrs Attributes) error {
	err := h.assertIsAllowedToWrite(ctx, sourceID, filename)
	if err != nil {
		return err
	}
	return h.next.SetFileMetaData(ctx, sourceID, filename, attrs)
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

func (h *PolicyHandler) OpenMultipartSession(ctx context.Context, sourceID string, filename string, info *MultipartSessionInfo) (string, error) {
	err := h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return "", err
	}
	return h.next.OpenMultipartSession(ctx, sourceID, filename, info)
}

func (h *PolicyHandler) AddContentPart(_ context.Context, _ string, _ io.Reader, _ int64, _ *ContentPartInfo) error {
	panic("implement me")
}

func (h *PolicyHandler) CloseMultipartSession(_ context.Context, _ string) error {
	panic("implement me")
}
