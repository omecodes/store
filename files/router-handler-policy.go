package files

import (
	"context"
	"io"
	"path"
	"strings"

	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/cenv"
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
	authInfo := auth.Get(ctx)
	if authInfo != nil {
		vars["auth"] = map[string]interface{}{
			"uid":    authInfo.Uid,
			"email":  authInfo.Email,
			"worker": authInfo.Worker,
			"scope":  authInfo.Scope,
			"group":  authInfo.Group,
		}
	}

	out, details, err := prg.Eval(vars)
	if err != nil {
		logs.Error("file permission evaluation", logs.Details("details", details))
		return false, err
	}

	return out.Value().(bool), nil
}

func (h *PolicyHandler) isAllowedToRead(ctx context.Context, filename string) (bool, error) {
	attrs, err := h.GetFileAttributes(ctx, filename, AttrPermissions)
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

func (h *PolicyHandler) isAllowedToWrite(ctx context.Context, filename string) (bool, error) {
	attrs, err := h.GetFileAttributes(ctx, filename, AttrPermissions)
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

func (h *PolicyHandler) isAllowedToChmod(ctx context.Context, filename string) (bool, error) {
	attrs, err := h.GetFileAttributes(ctx, filename, AttrPermissions)
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

func (h *PolicyHandler) CreateDir(ctx context.Context, filename string) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToWrite(ctx, filename)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.CreateDir(ctx, filename)
}

func (h *PolicyHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts WriteOptions) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToWrite(ctx, filename)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.WriteFileContent(ctx, filename, content, size, opts)
}

func (h *PolicyHandler) ListDir(ctx context.Context, dirname string, opts ListDirOptions) (*DirContent, error) {
	source := GetSource(ctx)
	if source == nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToRead(ctx, dirname)
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, errors.Create(errors.Unauthorized, "not_allowed")
	}
	return h.next.ListDir(ctx, dirname, opts)
}

func (h *PolicyHandler) ReadFileContent(ctx context.Context, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	source := GetSource(ctx)
	if source == nil {
		return nil, 0, errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToRead(ctx, filename)
	if err != nil {
		return nil, 0, err
	}

	if !allowed {
		return nil, 0, errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.ReadFileContent(ctx, filename, opts)
}

func (h *PolicyHandler) GetFileInfo(ctx context.Context, filename string, opts GetFileInfoOptions) (*File, error) {
	source := GetSource(ctx)
	if source == nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToRead(ctx, filename)
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.GetFileInfo(ctx, filename, opts)
}

func (h *PolicyHandler) DeleteFile(ctx context.Context, filename string, opts DeleteFileOptions) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToWrite(ctx, filename)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.DeleteFile(ctx, filename, opts)
}

func (h *PolicyHandler) SetFileMetaData(ctx context.Context, filename string, attrs Attributes) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToWrite(ctx, filename)
	if err != nil {
		return err
	}

	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.SetFileMetaData(ctx, filename, attrs)
}

func (h *PolicyHandler) GetFileAttributes(ctx context.Context, filename string, name ...string) (Attributes, error) {
	source := GetSource(ctx)
	if source == nil {
		return nil, errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToRead(ctx, filename)
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.GetFileAttributes(ctx, filename, name...)
}

func (h *PolicyHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToRead(ctx, filename)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	allowed, err = h.isAllowedToWrite(ctx, path.Dir(filename))
	if err != nil {
		return err
	}
	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.RenameFile(ctx, filename, newName)
}

func (h *PolicyHandler) MoveFile(ctx context.Context, filename string, dirname string) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToRead(ctx, filename)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	allowed, err = h.isAllowedToWrite(ctx, path.Dir(dirname))
	if err != nil {
		return err
	}
	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.MoveFile(ctx, filename, dirname)
}

func (h *PolicyHandler) CopyFile(ctx context.Context, filename string, dirname string) error {
	source := GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToRead(ctx, filename)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	allowed, err = h.isAllowedToWrite(ctx, path.Dir(dirname))
	if err != nil {
		return err
	}
	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.CopyFile(ctx, filename, dirname)
}

func (h *PolicyHandler) OpenMultipartSession(ctx context.Context, filename string, info *MultipartSessionInfo) (string, error) {
	allowed, err := h.isAllowedToWrite(ctx, path.Dir(filename))
	if err != nil {
		return "", err
	}

	if !allowed {
		return "", errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.OpenMultipartSession(ctx, filename, info)
}

func (h *PolicyHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *ContentPartInfo) error {
	panic("implement me")
}

func (h *PolicyHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	panic("implement me")
}
