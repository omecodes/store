package router

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/pb"
	"strings"
	"time"
)

type celParams struct {
	auth *pb.Auth
	data *pb.Header
}

func evaluate(ctx *context.Context, state *celParams, rule string) (bool, error) {
	if rule == "" || rule == "false" {
		return false, nil
	}

	if rule == "true" {
		return true, nil
	}

	prg, err := LoadProgramForACLValidation(ctx, rule)
	if err != nil {
		return false, err
	}

	vars := map[string]interface{}{
		"auth": map[string]interface{}{
			"uid":    state.auth.Uid,
			"email":  state.auth.Email,
			"worker": state.auth.Worker,
			"scope":  state.auth.Scope,
			"group":  state.auth.Group,
		},
		"data": map[string]interface{}{
			"id":         state.data.Id,
			"created_by": state.data.CreatedBy,
			"created_at": state.data.CreatedAt,
			"size":       state.data.Size,
		},
		"at": time.Now().UnixNano() / 1e6,
	}

	out, details, err := prg.Eval(vars)
	if err != nil {
		log.Error("cel execution", log.Field("details", details))
		return false, err
	}

	return out.Value().(bool), nil
}

func assetActionAllowedOnObject(ctx *context.Context, collection string, objectID string, action pb.AllowedTo, path string) error {
	header, err := GetObjectHeader(ctx, collection, objectID)
	if err != nil {
		return err
	}

	authCEL := auth.Get(*ctx)
	if authCEL == nil {
		authCEL = &pb.Auth{}
	}

	s := &celParams{
		data: header,
		auth: authCEL,
	}

	rule, err := getAccessRule(*ctx, collection, objectID, action, path)
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

func getAccessRule(ctx context.Context, collection string, objectID string, action pb.AllowedTo, path string) (string, error) {
	accessStore := acl.GetStore(ctx)
	if accessStore == nil {
		log.Error("ACL-Read-Check: missing access store in context")
		return "", errors.Internal
	}

	ruleCollection, err := accessStore.GetRules(ctx, collection, objectID)
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
	case pb.AllowedTo_read:
		actionRules = rules.Read
	case pb.AllowedTo_write:
		actionRules = rules.Write
	case pb.AllowedTo_delete:
		actionRules = rules.Delete

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
