package acl

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/pb"
)

func NewUnitServerHandler() pb.ACLServer {
	return &handler{}
}

type handler struct {
	pb.UnimplementedACLServer
}

func (h *handler) PutRules(ctx context.Context, request *pb.PutRulesRequest) (*pb.PutRulesResponse, error) {
	store := GetStore(ctx)
	if store == nil {
		log.Error("ACL Service • no store associated with context")
		return nil, errors.Internal
	}

	err := store.SaveRules(ctx, request.Collection, request.ObjectId, request.Rules)
	if err != nil {
		log.Error("ACL Service • could not save rules", log.Err(err))
		return nil, errors.Internal
	}

	return &pb.PutRulesResponse{}, nil
}

func (h *handler) GetRules(ctx context.Context, request *pb.GetRulesRequest) (*pb.GetRulesResponse, error) {
	store := GetStore(ctx)
	if store == nil {
		log.Error("ACL Service • no store associated with context")
		return nil, errors.Internal
	}

	rules, err := store.GetRules(ctx, request.Collection, request.ObjectId)
	return &pb.GetRulesResponse{Rules: rules}, err
}

func (h *handler) GetRulesForPath(ctx context.Context, request *pb.GetRulesForPathRequest) (*pb.GetRulesForPathResponse, error) {
	store := GetStore(ctx)
	if store == nil {
		log.Error("ACL Service • no store associated with context")
		return nil, errors.Internal
	}

	rsp := &pb.GetRulesForPathResponse{
		Rules: &pb.PathAccessRules{
			AccessRules: map[string]*pb.AccessRules{},
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

func (h *handler) DeleteRules(ctx context.Context, request *pb.DeleteRulesRequest) (*pb.DeleteRulesResponse, error) {
	store := GetStore(ctx)
	if store == nil {
		log.Error("ACL Service • no store associated with context")
		return nil, errors.Internal
	}

	err := store.Delete(ctx, request.Collection, request.ObjectId)
	return &pb.DeleteRulesResponse{}, err
}

func (h *handler) DeleteRulesForPath(ctx context.Context, request *pb.DeleteRulesForPathRequest) (*pb.DeleteRulesForPathResponse, error) {
	store := GetStore(ctx)
	if store == nil {
		log.Error("ACL Service • no store associated with context")
		return nil, errors.Internal
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
		log.Error("ACL Service • could not save update rules", log.Err(err))
		return nil, err
	}

	return &pb.DeleteRulesForPathResponse{}, nil
}
