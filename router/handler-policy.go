package router

import (
	"context"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/jsonpb"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/libome/v2"
	"github.com/omecodes/omestore/oms"
)

type celParams struct {
	auth *oms.Auth
	data *oms.Info
	at   int64
}

type policyHandler struct {
	base
}

func (p *policyHandler) isAdmin(ctx context.Context) bool {
	authCEL := authInfo(ctx)
	if authCEL == nil {
		return false
	}
	return authCEL.Uid == "admin"
}

// Route Handler methods
func (p *policyHandler) SetSettings(ctx context.Context, value *oms.JSON, opts oms.SettingsOptions) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}

	env := celPolicyEnv(ctx)
	if env == nil {
		return errors.Internal
	}
	return p.base.SetSettings(ctx, value, opts)
}

func (p *policyHandler) GetSettings(ctx context.Context, opts oms.SettingsOptions) (*oms.JSON, error) {
	if !p.isAdmin(ctx) {
		return nil, errors.Forbidden
	}
	return p.base.GetSettings(ctx, opts)
}

func (p *policyHandler) RegisterWorker(ctx context.Context, info *oms.JSON) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.base.RegisterWorker(ctx, info)
}

func (p *policyHandler) ListWorkers(ctx context.Context) ([]*oms.JSON, error) {
	if !p.isAdmin(ctx) {
		cred := ome.CredentialsFromContext(ctx)
		if cred == nil {
			return nil, errors.Forbidden
		}
	}
	return p.base.ListWorkers(ctx)
}

func (p *policyHandler) PutData(ctx context.Context, object *oms.Object, opts oms.PutDataOptions) (string, error) {
	err := p.assertIsAllowedOnData(&ctx, oms.AllowedTo_create, object.ID())
	if err != nil {
		return "", err
	}
	return p.base.PutData(ctx, object, opts)
}

func (p *policyHandler) GetData(ctx context.Context, id string, opts oms.GetDataOptions) (*oms.Object, error) {
	if !p.isAdmin(ctx) {
		err := p.assertIsAllowedOnData(&ctx, oms.AllowedTo_read, id)
		if err != nil {
			return nil, err
		}
	}
	return p.base.GetData(ctx, id, opts)
}

func (p *policyHandler) Info(ctx context.Context, id string) (*oms.Info, error) {
	if p.isAdmin(ctx) {
		return p.base.Info(ctx, id)
	}

	err := p.assertIsAllowedOnData(&ctx, oms.AllowedTo_read, id)
	if err != nil {
		return nil, err
	}
	return dataInfo(ctx, id), nil
}

func (p *policyHandler) Delete(ctx context.Context, id string) error {
	if !p.isAdmin(ctx) {
		err := p.assertIsAllowedOnData(&ctx, oms.AllowedTo_read, id)
		if err != nil {
			return err
		}
	}
	return p.base.Delete(ctx, id)
}

func (p *policyHandler) List(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	if !p.isAdmin(ctx) {
		opts.Filter = oms.FilterObjectFunc(func(o *oms.Object) (bool, error) {
			err := p.assertIsAllowedOnData(&ctx, oms.AllowedTo_read, o.ID())
			if err != nil {
				if err == errors.Unauthorized || err == errors.Forbidden {
					return false, nil
				}
				return false, err
			}
			return true, nil
		})
	}
	return p.base.List(ctx, opts)
}

func (p *policyHandler) Search(ctx context.Context, opts oms.SearchOptions) (*oms.ObjectList, error) {
	searchEnv := celSearchEnv(ctx)
	ast, issues := searchEnv.Compile(opts.MatchedExpression)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	matchProgram, err := searchEnv.Program(ast)
	if err != nil {
		return nil, err
	}

	if !p.isAdmin(ctx) {
		opts.Filter = oms.FilterObjectFunc(func(o *oms.Object) (bool, error) {
			err := p.assertIsAllowedOnData(&ctx, oms.AllowedTo_read, o.ID())
			if err != nil {
				if err == errors.Unauthorized || err == errors.Forbidden {
					return false, nil
				}
				return false, err
			}

			if opts.MatchedExpression == "" {
				return false, nil
			}
			if opts.MatchedExpression == "false" {
				return false, nil
			}
			if opts.MatchedExpression == "true" {
				return true, nil
			}
			if opts.MatchedExpression != "" {
				var object types.Struct
				err = jsonpb.Unmarshal(o.Content(), &object)
				if err != nil {
					return false, err
				}

				vars := map[string]interface{}{"o": object}
				out, details, err := matchProgram.Eval(vars)
				if err != nil {
					log.Error("cel execution", log.Field("details", details))
					return false, err
				}
				return out.Value().(bool), nil
			}
			return true, nil
		})
	}
	return p.base.Search(ctx, opts)
}
