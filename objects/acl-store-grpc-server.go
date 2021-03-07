package objects

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
)

func NewUnitServerHandler() ACLServer {
	return &aclHandler{}
}

type aclHandler struct {
	UnimplementedACLServer
}

func (h *aclHandler) PutRules(ctx context.Context, request *PutRulesRequest) (*PutRulesResponse, error) {
	store := GetACLStore(ctx)
	if store == nil {
		logs.Error("ACL Service • no store associated with context")
		return nil, errors.Internal("missing ACL store")
	}

	err := store.SaveRules(ctx, request.Collection, request.ObjectId, request.Rules)
	if err != nil {
		logs.Error("ACL Service • could not save rules", logs.Err(err))
		return nil, errors.Internal("could not save ACL rules")
	}

	return &PutRulesResponse{}, nil
}

func (h *aclHandler) GetRules(ctx context.Context, request *GetRulesRequest) (*GetRulesResponse, error) {
	store := GetACLStore(ctx)
	if store == nil {
		logs.Error("ACL Service • no store associated with context")
		return nil, errors.Internal("missing ACL store")
	}

	rules, err := store.GetRules(ctx, request.Collection, request.ObjectId)
	return &GetRulesResponse{Rules: rules}, err
}

func (h *aclHandler) GetRulesForPath(ctx context.Context, request *GetRulesForPathRequest) (*GetRulesForPathResponse, error) {
	store := GetACLStore(ctx)
	if store == nil {
		logs.Error("ACL Service • no store associated with context")
		return nil, errors.Internal("missing ACl store")
	}

	rsp := &GetRulesForPathResponse{
		Rules: &PathAccessRules{
			AccessRules: map[string]*AccessRules{},
		},
	}

	for _, path := range request.Paths {
		accessRules, err := store.GetForPath(ctx, request.Collection, request.ObjectId, path)
		if err != nil {
			return nil, err
		}
		rsp.Rules.AccessRules[path] = accessRules
	}

	return rsp, nil
}

func (h *aclHandler) DeleteRules(ctx context.Context, request *DeleteRulesRequest) (*DeleteRulesResponse, error) {
	store := GetACLStore(ctx)
	if store == nil {
		logs.Error("ACL Service • no store associated with context")
		return nil, errors.Internal("missing ACL store")
	}

	err := store.Delete(ctx, request.Collection, request.ObjectId)
	return &DeleteRulesResponse{}, err
}

func (h *aclHandler) DeleteRulesForPath(ctx context.Context, request *DeleteRulesForPathRequest) (*DeleteRulesForPathResponse, error) {
	store := GetACLStore(ctx)
	if store == nil {
		logs.Error("ACL Service • no store associated with context")
		return nil, errors.Internal("missing ACL store")
	}

	rules, err := store.GetRules(ctx, request.Collection, request.ObjectId)
	if err != nil {
		return nil, err
	}

	for _, requestPath := range request.Paths {
		delete(rules.AccessRules, requestPath)
	}

	err = store.SaveRules(ctx, request.Collection, request.ObjectId, rules)
	if err != nil {
		logs.Error("ACL Service • could not save update rules", logs.Err(err))
		return nil, err
	}

	return &DeleteRulesForPathResponse{}, nil
}
