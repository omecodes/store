package router

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/oms"
	"strings"
	"time"
)

type celParams struct {
	auth *oms.Auth
	data *oms.Header
}

func evaluate(ctx *context.Context, state *celParams, rule string) (bool, error) {
	if rule == "" || rule == "false" {
		return false, nil
	}

	if rule == "true" {
		return true, nil
	}

	prg, err := loadProgramForAccessValidation(ctx, rule)
	if err != nil {
		return false, err
	}

	vars := map[string]interface{}{
		"auth": state.auth,
		"data": state.data,
		"at":   time.Now().Unix(),
	}
	out, details, err := prg.Eval(vars)

	if err != nil {
		log.Error("cel execution", log.Field("details", details))
		return false, err
	}

	return out.Value().(bool), nil
}

func assetActionAllowedOnObject(ctx *context.Context, action oms.AllowedTo, objectID string, path string) error {
	header, err := getObjectHeader(ctx, objectID)
	if err != nil {
		return err
	}

	authCEL := authInfo(*ctx)
	if authCEL == nil {
		authCEL = &oms.Auth{}
	}

	s := &celParams{
		data: header,
		auth: authCEL,
	}

	rule, err := getAccessRule(*ctx, action, objectID, path)
	if err != nil {
		return err
	}

	allowed, err := evaluate(ctx, s, rule)
	if err != nil {
		log.Error("failed to evaluate access rule", log.Err(err))
		return errors.Internal
	}

	if !allowed {
		return errors.Unauthorized
	}

	return nil
}

func getAccessRule(ctx context.Context, action oms.AllowedTo, objectID string, path string) (string, error) {
	accessStore := accessStore(ctx)
	if accessStore == nil {
		log.Error("ACL-Read-Check: missing access store in context")
		return "", errors.Internal
	}

	ruleCollection, err := accessStore.GetRules(objectID)
	if err != nil {
		return "", err
	}

	if path == "" {
		path = "$"
	}

	rules, found := ruleCollection.AccessRules[path]
	if !found {
		log.Error("ACL: could not find access security rule", log.Field("object", objectID), log.Field("path", path))
		return "", errors.Forbidden
	}

	var actionRules []string
	switch action {
	case oms.AllowedTo_read:
		actionRules = rules.Read
	case oms.AllowedTo_write:
		actionRules = rules.Write

	default:
		log.Error("ACL: no rule for this action", log.Field("action", action.String()))
		return "", errors.Internal
	}

	var formattedRules []string
	for _, exp := range actionRules {
		formattedRules = append(formattedRules, "("+exp+")")
	}
	rule := strings.Join(formattedRules, " || ")

	return rule, nil
}
