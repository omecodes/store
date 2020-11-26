package router

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/oms"
	"time"
)

func (p *policyHandler) evaluate(ctx *context.Context, state *celParams, rule string) (bool, error) {
	if rule == "" || rule == "false" {
		return false, nil
	}

	if rule == "true" {
		return true, nil
	}

	prg, err := getACLProgram(ctx, rule)
	if err != nil {
		return false, err
	}

	vars := map[string]interface{}{
		"auth": state.auth,
		"data": state.data,
		"at":   state.at,
	}
	out, details, err := prg.Eval(vars)

	if err != nil {
		log.Error("cel execution", log.Field("details", details))
		return false, err
	}

	return out.Value().(bool), nil
}

func (p *policyHandler) actionRuleForData(ctx *context.Context, action oms.AllowedTo) (string, error) {
	route := Route(SkipPoliciesCheck(), SkipParamsCheck())
	settings, err := route.GetSettings(*ctx, oms.SettingsOptions{Path: oms.SettingsDataAccessSecurityRulesPath})
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

func (p *policyHandler) actionRuleForGraft(ctx *context.Context, action oms.AllowedTo) (string, error) {
	route := Route(SkipPoliciesCheck(), SkipParamsCheck())
	settings, err := route.GetSettings(*ctx, oms.SettingsOptions{Path: oms.SettingsGraftAccessSecurityRulesPath})
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

func (p *policyHandler) assertIsAllowedOnData(ctx *context.Context, action oms.AllowedTo, id string) error {
	authCEL := authInfo(*ctx)
	if authCEL == nil {
		authCEL = &oms.Auth{}
	}

	s := &celParams{
		at:   time.Now().Unix(),
		data: &oms.Info{},
		auth: authCEL,
	}

	// load default 'action' rule for data
	rule, err := p.actionRuleForData(ctx, action)
	if err != nil {
		return err
	}

	if action != oms.AllowedTo_create {
		route := Route(SkipPoliciesCheck(), SkipParamsCheck())
		info, err := route.Info(*ctx, id)
		if err != nil {
			return err
		}
		*ctx = contextWithDataInfo(*ctx, id, info)
		s.data.CreatedBy = info.CreatedBy
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
