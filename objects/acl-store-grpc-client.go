package objects

import (
	"context"
)

func NewACLGrpcClient() *gRPCClient {
	return &gRPCClient{}
}

type gRPCClient struct{}

func (g *gRPCClient) SaveRules(ctx context.Context, collection string, objectID string, rules *PathAccessRules) error {
	client, err := AclGRPCClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.PutRules(ctx, &PutRulesRequest{Collection: collection, ObjectId: objectID, Rules: rules})
	return err
}

func (g *gRPCClient) GetRules(ctx context.Context, collection string, objectID string) (*PathAccessRules, error) {
	client, err := AclGRPCClient(ctx)
	if err != nil {
		return nil, err
	}
	rsp, err := client.GetRules(ctx, &GetRulesRequest{Collection: collection, ObjectId: objectID})
	if err != nil {
		return nil, err
	}
	return rsp.GetRules(), nil
}

func (g *gRPCClient) GetForPath(ctx context.Context, collection string, objectID string, path string) (*AccessRules, error) {
	client, err := AclGRPCClient(ctx)
	if err != nil {
		return nil, err
	}
	rsp, err := client.GetRulesForPath(ctx, &GetRulesForPathRequest{Collection: collection, ObjectId: objectID, Paths: []string{path}})
	if err != nil {
		return nil, err
	}
	return rsp.Rules.AccessRules[path], nil
}

func (g *gRPCClient) Delete(ctx context.Context, collection string, objectID string) error {
	client, err := AclGRPCClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.DeleteRules(ctx, &DeleteRulesRequest{Collection: collection, ObjectId: objectID})
	return err
}

func (g *gRPCClient) DeleteForPath(ctx context.Context, collection string, objectID string, path string) error {
	client, err := AclGRPCClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.DeleteRulesForPath(ctx, &DeleteRulesForPathRequest{Collection: collection, ObjectId: objectID, Paths: []string{path}})
	return err
}
