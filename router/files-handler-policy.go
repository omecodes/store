package router

import (
	"context"
	"io"
	"path"
	"strings"

	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/cenv"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/pb"
)

type FilesPolicyHandler struct {
	FilesBaseObjectsHandler
}

func (h *FilesPolicyHandler) evaluateRules(ctx context.Context, rules ...string) (bool, error) {
	var formattedRules []string
	for _, exp := range rules {
		if exp == "true" {
			return true, nil
		}
		formattedRules = append(formattedRules, "("+exp+")")
	}
	fullExpression := strings.Join(formattedRules, " || ")

	env := CELPolicyEnv(ctx)
	if env == nil {
		return false, errors.Create(errors.Internal, "context missing access rule evaluator")
	}

	prg, err := cenv.GetProgram(env, fullExpression)
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

func (h *FilesPolicyHandler) isAllowedToRead(ctx context.Context, filename string) (bool, error) {
	attrValue, err := h.GetFileMetaData(ctx, filename, files.AttrReadPermissions)
	if err != nil {
		return false, err
	}

	rules, err := files.DecodePermissions(attrValue)
	if err != nil {
		return false, err
	}
	return h.evaluateRules(ctx, rules...)
}

func (h *FilesPolicyHandler) isAllowedToWrite(ctx context.Context, filename string) (bool, error) {
	attrValue, err := h.GetFileMetaData(ctx, filename, files.AttrWritePermissions)
	if err != nil {
		return false, err
	}

	rules, err := files.DecodePermissions(attrValue)
	if err != nil {
		return false, err
	}
	return h.evaluateRules(ctx, rules...)
}

func (h *FilesPolicyHandler) isAllowedToChmod(ctx context.Context, filename string) (bool, error) {
	attrValue, err := h.GetFileMetaData(ctx, filename, files.AttrChmodPermissions)
	if err != nil {
		return false, err
	}

	rules, err := files.DecodePermissions(attrValue)
	if err != nil {
		return false, err
	}
	return h.evaluateRules(ctx, rules...)
}

func (h *FilesPolicyHandler) CreateDir(ctx context.Context, filename string) error {
	source := files.GetSource(ctx)
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

func (h *FilesPolicyHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts pb.PutFileOptions) error {
	source := files.GetSource(ctx)
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

func (h *FilesPolicyHandler) ListDir(ctx context.Context, dirname string, opts pb.GetFileInfoOptions) ([]*pb.File, error) {
	source := files.GetSource(ctx)
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

func (h *FilesPolicyHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	source := files.GetSource(ctx)
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

func (h *FilesPolicyHandler) GetFileInfo(ctx context.Context, filename string, opts pb.GetFileInfoOptions) (*pb.File, error) {
	source := files.GetSource(ctx)
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

func (h *FilesPolicyHandler) DeleteFile(ctx context.Context, filename string) error {
	source := files.GetSource(ctx)
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

	return h.next.DeleteFile(ctx, filename)
}

func (h *FilesPolicyHandler) SetFileMetaData(ctx context.Context, filename string, name files.AttrName, value string) error {
	source := files.GetSource(ctx)
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

	return h.next.SetFileMetaData(ctx, filename, name, value)
}

func (h *FilesPolicyHandler) GetFileMetaData(ctx context.Context, filename string, name files.AttrName) (string, error) {
	source := files.GetSource(ctx)
	if source == nil {
		return "", errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToRead(ctx, filename)
	if err != nil {
		return "", err
	}

	if !allowed {
		return "", errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.GetFileMetaData(ctx, filename, name)
}

func (h *FilesPolicyHandler) RenameFile(ctx context.Context, filename string, newName string) error {
	source := files.GetSource(ctx)
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

func (h *FilesPolicyHandler) MoveFile(ctx context.Context, srcFilename string, dstFilename string) error {
	source := files.GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToRead(ctx, srcFilename)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	allowed, err = h.isAllowedToWrite(ctx, path.Dir(dstFilename))
	if err != nil {
		return err
	}
	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.MoveFile(ctx, srcFilename, dstFilename)
}

func (h *FilesPolicyHandler) CopyFile(ctx context.Context, srcFilename string, dstFilename string) error {
	source := files.GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	allowed, err := h.isAllowedToRead(ctx, srcFilename)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	allowed, err = h.isAllowedToWrite(ctx, path.Dir(dstFilename))
	if err != nil {
		return err
	}
	if !allowed {
		return errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.CopyFile(ctx, srcFilename, dstFilename)
}

func (h *FilesPolicyHandler) OpenMultipartSession(ctx context.Context, filename string, info *pb.MultipartSessionInfo) (string, error) {
	allowed, err := h.isAllowedToWrite(ctx, path.Dir(filename))
	if err != nil {
		return "", err
	}

	if !allowed {
		return "", errors.Create(errors.Unauthorized, "not_allowed")
	}

	return h.next.OpenMultipartSession(ctx, filename, info)
}

func (h *FilesPolicyHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error {
	panic("implement me")
}

func (h *FilesPolicyHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	panic("implement me")
}
