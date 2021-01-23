package files

import (
	"context"
	"io"
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

func (h *PolicyHandler) evaluateRules(ctx context.Context, rules ...string) (bool, error) {
	var formattedRules []string
	for _, exp := range rules {
		if exp == "true" {
			return true, nil
		}
		formattedRules = append(formattedRules, "("+exp+")")
	}
	fullExpression := strings.Join(formattedRules, " || ")

	prg, err := cenv.GetProgram(fullExpression)
	if err != nil {
		return false, errors.Create(errors.Internal, "context missing access rule evaluator")
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
		return false, err
	}

	return out.Value().(bool), nil
}

func (h *PolicyHandler) assertIsAllowedToRead(ctx context.Context, sourceID string, filename string) (bool, error) {
	user := auth.Get(ctx)
	if user != nil && user.Name == "admin" {
		return true, nil
	}

	source, err := h.next.GetSource(ctx, sourceID)
	if source == nil {
		return false, err
	}

	if source.PermissionOverrides != nil && len(source.PermissionOverrides.Read) > 0 {
		var rules []string
		for _, wr := range source.PermissionOverrides.Read {
			rules = append(rules, wr.Rule)
		}
		return h.evaluateRules(ctx, rules...)
	}

	attrs, err := h.next.GetFileAttributes(ctx, sourceID, filename, AttrPermissions)
	if err != nil {
		return false, err
	}

	attrsHolder := HoldAttributes(attrs)
	perms, err := attrsHolder.GetPermissions()
	if err != nil {
		return false, err
	}

	var rules []string
	for _, perm := range perms.Read {
		rules = append(rules, perm.Rule)
	}

	return h.evaluateRules(ctx, rules...)
}

func (h *PolicyHandler) assertIsAllowedToWrite(ctx context.Context, sourceID string, filename string) (bool, error) {
	user := auth.Get(ctx)
	if user != nil && user.Name == "admin" {
		return true, nil
	}

	source, err := h.next.GetSource(ctx, sourceID)
	if source == nil {
		return false, err
	}

	if source.PermissionOverrides != nil && len(source.PermissionOverrides.Write) > 0 {
		var rules []string
		for _, wr := range source.PermissionOverrides.Write {
			rules = append(rules, wr.Rule)
		}
		return h.evaluateRules(ctx, rules...)
	}

	attrs, err := h.next.GetFileAttributes(ctx, sourceID, filename, AttrPermissions)
	if err != nil {
		return false, err
	}

	attrsHolder := HoldAttributes(attrs)
	perms, err := attrsHolder.GetPermissions()
	if err != nil {
		return false, err
	}

	var rules []string
	for _, perm := range perms.Write {
		rules = append(rules, perm.Rule)
	}

	return h.evaluateRules(ctx, rules...)
}

func (h *PolicyHandler) assertIsAllowedToChmod(ctx context.Context, sourceID string, filename string) (bool, error) {
	user := auth.Get(ctx)
	if user != nil && user.Name == "admin" {
		return true, nil
	}

	source, err := h.next.GetSource(ctx, sourceID)
	if source == nil {
		return false, err
	}

	if source.PermissionOverrides != nil && len(source.PermissionOverrides.Chmod) > 0 {
		var rules []string
		for _, wr := range source.PermissionOverrides.Chmod {
			rules = append(rules, wr.Rule)
		}
		return h.evaluateRules(ctx, rules...)
	}

	attrs, err := h.next.GetFileAttributes(ctx, sourceID, filename, AttrPermissions)
	if err != nil {
		return false, err
	}

	attrsHolder := HoldAttributes(attrs)
	perms, err := attrsHolder.GetPermissions()
	if err != nil {
		return false, err
	}

	var rules []string
	for _, perm := range perms.Chmod {
		rules = append(rules, perm.Rule)
	}

	return h.evaluateRules(ctx, rules...)
}

func (h *PolicyHandler) CreateSource(ctx context.Context, source *Source) error {
	return h.next.CreateSource(ctx, source)
}

func (h *PolicyHandler) ListSources(ctx context.Context) ([]*Source, error) {
	return h.next.ListSources(ctx)
}

func (h *PolicyHandler) GetSource(ctx context.Context, sourceID string) (*Source, error) {
	return h.next.GetSource(ctx, sourceID)
}

func (h *PolicyHandler) DeleteSource(ctx context.Context, sourceID string) error {
	return h.next.DeleteSource(ctx, sourceID)
}

func (h *PolicyHandler) CreateDir(ctx context.Context, sourceID string, filename string) error {
	allowed, err := h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not allowed")
	}
	return h.next.CreateDir(ctx, sourceID, filename)
}

func (h *PolicyHandler) WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	allowed, err := h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not allowed")
	}
	return h.next.WriteFileContent(ctx, sourceID, filename, content, size, opts)
}

func (h *PolicyHandler) ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	allowed, err := h.assertIsAllowedToRead(ctx, sourceID, dirname)
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, errors.Create(errors.Unauthorized, "not allowed")
	}
	return h.next.ListDir(ctx, sourceID, dirname, opts)
}

func (h *PolicyHandler) ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	allowed, err := h.assertIsAllowedToRead(ctx, sourceID, filename)
	if err != nil {
		return nil, 0, err
	}

	if !allowed {
		return nil, 0, errors.Create(errors.Unauthorized, "not allowed")
	}
	return h.next.ReadFileContent(ctx, sourceID, filename, opts)
}

func (h *PolicyHandler) GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileInfoOptions) (*File, error) {
	allowed, err := h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, errors.Create(errors.Unauthorized, "not allowed")
	}
	return h.next.GetFileInfo(ctx, sourceID, filename, opts)
}

func (h *PolicyHandler) DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error {
	allowed, err := h.assertIsAllowedToWrite(ctx, sourceID, filename)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not allowed")
	}

	return h.next.DeleteFile(ctx, sourceID, filename, opts)
}

func (h *PolicyHandler) SetFileMetaData(ctx context.Context, sourceID string, filename string, attrs Attributes) error {
	allowed, err := h.assertIsAllowedToWrite(ctx, sourceID, filename)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not allowed")
	}

	return h.next.SetFileMetaData(ctx, sourceID, filename, attrs)
}

func (h *PolicyHandler) GetFileAttributes(ctx context.Context, sourceID string, filename string, name ...string) (Attributes, error) {
	allowed, err := h.assertIsAllowedToRead(ctx, sourceID, filename)
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, errors.Create(errors.Unauthorized, "not allowed")
	}
	return h.next.GetFileAttributes(ctx, sourceID, filename, name...)
}

func (h *PolicyHandler) RenameFile(ctx context.Context, sourceID string, filename string, newName string) error {
	allowed, err := h.assertIsAllowedToRead(ctx, sourceID, filename)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not allowed")
	}

	allowed, err = h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not allowed")
	}

	return h.next.RenameFile(ctx, sourceID, filename, newName)
}

func (h *PolicyHandler) MoveFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	allowed, err := h.assertIsAllowedToWrite(ctx, sourceID, filename)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not allowed")
	}

	allowed, err = h.assertIsAllowedToWrite(ctx, sourceID, dirname)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not allowed")
	}

	return h.next.MoveFile(ctx, filename, sourceID, dirname)
}

func (h *PolicyHandler) CopyFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	allowed, err := h.assertIsAllowedToRead(ctx, sourceID, filename)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not allowed")
	}

	allowed, err = h.assertIsAllowedToWrite(ctx, sourceID, dirname)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not allowed")
	}

	return h.next.CopyFile(ctx, sourceID, filename, dirname)
}

func (h *PolicyHandler) OpenMultipartSession(ctx context.Context, sourceID string, filename string, info *MultipartSessionInfo) (string, error) {
	allowed, err := h.assertIsAllowedToWrite(ctx, sourceID, path.Dir(filename))
	if err != nil {
		return "", err
	}

	if !allowed {
		return "", errors.Create(errors.Unauthorized, "not allowed")
	}
	return h.next.OpenMultipartSession(ctx, sourceID, filename, info)
}

func (h *PolicyHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *ContentPartInfo) error {
	panic("implement me")
}

func (h *PolicyHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	panic("implement me")
}
