package oms

import (
	"context"
	"time"

	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/libome/v2"
)

type celParams struct {
	auth  *Auth
	data  *Object
	graft *Graft
	at    int64
}

type policyHandler struct {
	base
}

func (p *policyHandler) evaluate(ctx *context.Context, state *celParams, rule string) (bool, error) {
	if rule == "" || rule == "false" {
		return false, nil
	}

	if rule == "true" {
		return true, nil
	}

	prg, err := getProgram(ctx, rule)
	if err != nil {
		return false, err
	}

	vars := map[string]interface{}{
		"auth":  state.auth,
		"data":  state.data,
		"graft": state.graft,
		"at":    state.at,
	}
	out, details, err := prg.Eval(vars)

	if err != nil {
		log.Error("cel execution", log.Field("details", details))
		return false, err
	}

	return out.Value().(bool), nil
}

func (p *policyHandler) actionRuleForData(ctx *context.Context, action AllowedTo) (string, error) {
	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
	settings, err := route.GetSettings(*ctx, SettingsOptions{Path: settingsDataAccessSecurityRulesPath})
	if err != nil {
		return "", errors.Internal
	}

	rule, err := settings.StringAt(action.String())
	if err != nil {
		log.Error("could not get 'read' rule for access security rules settings", log.Err(err))
		return "", errors.Internal
	}
	return rule, nil
}

func (p *policyHandler) actionRuleForGraft(ctx *context.Context, action AllowedTo) (string, error) {
	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
	settings, err := route.GetSettings(*ctx, SettingsOptions{Path: settingsGraftAccessSecurityRulesPath})
	if err != nil {
		return "", errors.Internal
	}

	rule, err := settings.StringAt(action.String())
	if err != nil {
		log.Error("could not get 'read' rule for access security rules settings", log.Err(err))
		return "", errors.Internal
	}
	return rule, nil
}

func (p *policyHandler) assertIsAllowedOnData(ctx *context.Context, action AllowedTo, collection string, id string) error {
	authCEL := authInfo(*ctx)
	if authCEL == nil {
		authCEL = &Auth{}
	}

	s := &celParams{
		at:    time.Now().Unix(),
		data:  &Object{},
		auth:  authCEL,
		graft: &Graft{},
	}

	// load default 'action' rule for data
	rule, err := p.actionRuleForData(ctx, action)
	if err != nil {
		return err
	}

	if action != AllowedTo_create {
		route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
		info, err := route.Info(*ctx, collection, id)
		if err != nil {
			return err
		}
		*ctx = contextWithDataInfo(*ctx, collection, id, info)
		s.data.CreatedBy = info.CreatedBy
		s.data.Collection = collection
		s.data.Id = id
	}

	allowed, err := p.evaluate(ctx, s, rule)
	if err != nil {
		log.Error("failed to evaluate access rule", log.Err(err))
		return errors.Internal
	}

	if !allowed {
		return errors.Unauthorized
	}

	return nil
}

func (p *policyHandler) assertIsAllowedOnGraft(ctx *context.Context, action AllowedTo, collection string, dataID string, id string) error {
	authCEL := authInfo(*ctx)
	if authCEL == nil {
		authCEL = &Auth{}
	}
	s := &celParams{
		at:    time.Now().Unix(),
		data:  &Object{},
		auth:  authCEL,
		graft: &Graft{},
	}

	// load default 'action' rule for data
	rule, err := p.actionRuleForGraft(ctx, action)
	if err != nil {
		return err
	}

	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())

	info, err := route.Info(*ctx, collection, dataID)
	if err != nil {
		return err
	}
	*ctx = contextWithDataInfo(*ctx, collection, id, info)
	s.data.CreatedBy = info.CreatedBy
	s.data.Collection = collection
	s.data.Id = id

	graftInfo, err := route.GraftInfo(*ctx, collection, dataID, id)
	if err != nil {
		return err
	}

	s.graft.Id = id
	s.graft.CreatedBy = graftInfo.CreatedBy
	s.graft.CreatedAt = graftInfo.CreatedAt

	allowed, err := p.evaluate(ctx, s, rule)
	if err != nil {
		log.Error("failed to evaluate access rule", log.Err(err))
		return errors.Internal
	}

	if !allowed {
		return errors.Unauthorized
	}

	return nil
}

func (p *policyHandler) isAdmin(ctx context.Context) bool {
	authCEL := authInfo(ctx)
	if authCEL == nil {
		return false
	}
	return authCEL.Uid == "admin"
}

// Route Handler methods
func (p *policyHandler) SetSettings(ctx context.Context, value *JSON, opts SettingsOptions) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}

	env := celEnv(ctx)
	if env == nil {
		return errors.Internal
	}
	return p.base.SetSettings(ctx, value, opts)
}

func (p *policyHandler) GetSettings(ctx context.Context, opts SettingsOptions) (*JSON, error) {
	if !p.isAdmin(ctx) {
		return nil, errors.Forbidden
	}
	return p.base.GetSettings(ctx, opts)
}

func (p *policyHandler) RegisterWorker(ctx context.Context, info *JSON) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.base.RegisterWorker(ctx, info)
}

func (p *policyHandler) ListWorkers(ctx context.Context) ([]*JSON, error) {
	if !p.isAdmin(ctx) {
		cred := ome.CredentialsFromContext(ctx)
		if cred == nil {
			return nil, errors.Forbidden
		}
	}
	return p.base.ListWorkers(ctx)
}

func (p *policyHandler) PutData(ctx context.Context, data *Object, opts PutDataOptions) error {
	err := p.assertIsAllowedOnData(&ctx, AllowedTo_create, data.Collection, data.Id)
	if err != nil {
		return err
	}
	return p.base.PutData(ctx, data, opts)
}

func (p *policyHandler) GetData(ctx context.Context, collection string, id string, opts GetDataOptions) (*Object, error) {
	if !p.isAdmin(ctx) {
		err := p.assertIsAllowedOnData(&ctx, AllowedTo_read, collection, id)
		if err != nil {
			return nil, err
		}
	}
	return p.base.GetData(ctx, collection, id, opts)
}

func (p *policyHandler) Info(ctx context.Context, collection string, id string) (*Info, error) {
	if p.isAdmin(ctx) {
		return p.base.Info(ctx, collection, id)
	}

	err := p.assertIsAllowedOnData(&ctx, AllowedTo_read, collection, id)
	if err != nil {
		return nil, err
	}
	return dataInfo(ctx, collection, id), nil
}

func (p *policyHandler) Delete(ctx context.Context, collection string, id string) error {
	if !p.isAdmin(ctx) {
		err := p.assertIsAllowedOnData(&ctx, AllowedTo_read, collection, id)
		if err != nil {
			return err
		}
	}
	return p.base.Delete(ctx, collection, id)
}

func (p *policyHandler) List(ctx context.Context, collection string, opts ListOptions) (*DataList, error) {
	if !p.isAdmin(ctx) {
		opts.IDFilter = IDFilterFunc(func(id string) (bool, error) {
			err := p.assertIsAllowedOnData(&ctx, AllowedTo_read, collection, id)
			if err != nil {
				if err == errors.Unauthorized || err == errors.Forbidden {
					return false, nil
				}
				return false, err
			}
			return true, nil
		})
	}
	return p.base.List(ctx, collection, opts)
}

func (p *policyHandler) SaveGraft(ctx context.Context, graft *Graft) (string, error) {
	if !p.isAdmin(ctx) {
		err := p.assertIsAllowedOnData(&ctx, AllowedTo_graft, graft.Collection, graft.DataId)
		if err != nil {
			return "", err
		}
	}
	return p.next.SaveGraft(ctx, graft)
}

func (p *policyHandler) GetGraft(ctx context.Context, collection string, dataID string, id string) (*Graft, error) {
	if !p.isAdmin(ctx) {
		err := p.assertIsAllowedOnGraft(&ctx, AllowedTo_read, collection, dataID, id)
		if err != nil {
			return nil, err
		}
	}
	return p.base.GetGraft(ctx, collection, dataID, id)
}

func (p *policyHandler) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*GraftInfo, error) {
	if p.isAdmin(ctx) {
		return p.base.GraftInfo(ctx, collection, dataID, id)
	}

	err := p.assertIsAllowedOnGraft(&ctx, AllowedTo_read, collection, dataID, id)
	if err != nil {
		return nil, err
	}
	return graftInfo(ctx, collection, dataID, id), nil
}

func (p *policyHandler) ListGrafts(ctx context.Context, collection string, dataID string, opts ListOptions) (*GraftList, error) {
	if !p.isAdmin(ctx) {
		opts.IDFilter = IDFilterFunc(func(id string) (bool, error) {
			err := p.assertIsAllowedOnGraft(&ctx, AllowedTo_read, collection, dataID, id)
			if err != nil {
				if err == errors.Unauthorized || err == errors.Forbidden {
					return false, nil
				}
				return false, err
			}
			return true, nil
		})
	}
	return p.base.ListGrafts(ctx, collection, dataID, opts)
}

func (p *policyHandler) DeleteGraft(ctx context.Context, collection string, dataID string, id string) error {
	if !p.isAdmin(ctx) {
		err := p.assertIsAllowedOnGraft(&ctx, AllowedTo_delete, collection, dataID, id)
		if err != nil {
			return err
		}
	}
	return p.base.DeleteGraft(ctx, collection, dataID, id)
}
