package acl

import (
	"context"
	"github.com/omecodes/store/clients"
	"github.com/omecodes/store/pb"
)

func NewStoreClient() *gRPCClient {
	return &gRPCClient{}
}

type gRPCClient struct{}

func (g *gRPCClient) SaveRules(ctx context.Context, objectID string, rules *pb.PathAccessRules) error {
	client, err := clients.ACLGrpc(ctx)
	if err != nil {
		return err
	}
	_, err = client.PutRules(ctx, &pb.PutRulesRequest{ObjectId: objectID, Rules: rules})
	return err
}

func (g *gRPCClient) GetRules(ctx context.Context, objectID string) (*pb.PathAccessRules, error) {
	client, err := clients.ACLGrpc(ctx)
	if err != nil {
		return nil, err
	}
	rsp, err := client.GetRules(ctx, &pb.GetRulesRequest{ObjectId: objectID})
	if err != nil {
		return nil, err
	}
	return rsp.GetRules(), nil
}

func (g *gRPCClient) GetForPath(ctx context.Context, objectID string, path string) (*pb.AccessRules, error) {
	client, err := clients.ACLGrpc(ctx)
	if err != nil {
		return nil, err
	}
	rsp, err := client.GetRulesForPath(ctx, &pb.GetRulesForPathRequest{ObjectId: objectID, Paths: []string{path}})
	if err != nil {
		return nil, err
	}
	return rsp.Rules.AccessRules[path], nil
}

func (g *gRPCClient) Delete(ctx context.Context, objectID string) error {
	client, err := clients.ACLGrpc(ctx)
	if err != nil {
		return err
	}
	_, err = client.DeleteRules(ctx, &pb.DeleteRulesRequest{ObjectId: objectID})
	return err
}

func (g *gRPCClient) DeleteForPath(ctx context.Context, objectID string, path string) error {
	client, err := clients.ACLGrpc(ctx)
	if err != nil {
		return err
	}
	_, err = client.DeleteRulesForPath(ctx, &pb.DeleteRulesForPathRequest{ObjectId: objectID, Paths: []string{path}})
	return err
}
