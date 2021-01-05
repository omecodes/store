package router

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/pb"
)

type PolicyHandler struct {
	BaseHandler
}

func (p *PolicyHandler) isAdmin(ctx context.Context) bool {
	authCEL := auth.Get(ctx)
	if authCEL == nil {
		return false
	}
	return authCEL.Uid == "admin"
}

func (p *PolicyHandler) PutObject(ctx context.Context, object *pb.Object, security *pb.PathAccessRules, indexes []*pb.Index, opts pb.PutOptions) (string, error) {
	ai := auth.Get(ctx)
	if ai == nil {
		return "", errors.Forbidden
	}

	docRules := security.AccessRules["$"]
	if docRules == nil {
		docRules = &pb.AccessRules{}
		security.AccessRules["$"] = docRules
	}

	userDefaultRule := fmt.Sprintf("auth.uid=='%s'", ai.Uid)
	if len(docRules.Read) == 0 {
		docRules.Read = append(docRules.Read, userDefaultRule, "auth.worker", "auth.uid=='admin'")
	} else {
		docRules.Read = append(docRules.Write, "auth.worker", "auth.uid=='admin'")
	}

	if len(docRules.Write) == 0 {
		docRules.Write = append(docRules.Write, userDefaultRule, "auth.worker", "auth.uid=='admin'")
	} else {
		docRules.Write = append(docRules.Write, "auth.worker", "auth.uid=='admin'")
	}

	if len(docRules.Delete) == 0 {
		docRules.Delete = append(docRules.Delete, userDefaultRule)
	}

	object.Header.CreatedBy = ai.Uid
	return p.BaseHandler.PutObject(ctx, object, security, indexes, opts)
}

func (p *PolicyHandler) GetObject(ctx context.Context, id string, opts pb.GetOptions) (*pb.Object, error) {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, id, opts.At)
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObject(ctx, id, opts)
}

func (p *PolicyHandler) PatchObject(ctx context.Context, patch *pb.Patch, opts pb.PatchOptions) error {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_delete, patch.ObjectId, "")
	if err != nil {
		return err
	}

	return p.BaseHandler.PatchObject(ctx, patch, opts)
}

func (p *PolicyHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, id, "")
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObjectHeader(ctx, id)
}

func (p *PolicyHandler) DeleteObject(ctx context.Context, id string) error {
	err := assetActionAllowedOnObject(&ctx, pb.AllowedTo_delete, id, "")
	if err != nil {
		return err
	}
	return p.BaseHandler.DeleteObject(ctx, id)
}

func (p *PolicyHandler) ListObjects(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error) {
	var (
		err           error
		searchProgram cel.Program
	)

	if opts.Condition != "" {
		if opts.Condition == "false" {
			return nil, errors.Forbidden
		}

		if opts.Condition != "true" {
			searchProgram, err = LoadProgramForSearch(&ctx, opts.Condition)
			if err != nil {
				log.Error("Expression compiling failure", log.Field("expression", opts.Condition), log.Err(err))
				return nil, errors.BadInput
			}
		}
	}

	cursor, err := p.BaseHandler.ListObjects(ctx, opts)
	if err != nil {
		return nil, err
	}

	cursorBrowser := cursor.GetBrowser()
	browser := pb.BrowseFunc(func() (*pb.Object, error) {
		for {
			o, err := cursorBrowser.Browse()
			if err != nil {
				return nil, err
			}

			err = assetActionAllowedOnObject(&ctx, pb.AllowedTo_read, o.Header.Id, opts.At)
			if err != nil {
				if errors.IsForbidden(err) {
					continue
				}
				return nil, err
			}

			if searchProgram != nil {
				object := map[string]interface{}{}
				err = json.Unmarshal([]byte(o.Data), &object)
				if err != nil {
					return nil, err
				}

				vars := map[string]interface{}{"o": object}
				out, details, err := searchProgram.Eval(vars)
				if err != nil {
					log.Error("cel execution", log.Field("details", details))
					return nil, err
				}

				if !out.Value().(bool) {
					continue
				}
			}
			return o, nil
		}
	})

	cursor.SetBrowser(browser)
	return cursor, nil
}
