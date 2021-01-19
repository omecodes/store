package objects

import (
	"context"
	"github.com/omecodes/store/auth"
	se "github.com/omecodes/store/search-engine"
)

// NewGRPCObjectsClientHandler creates a router ObjectsHandler that embed that calls a gRPC service to perform final actions
func NewGRPCObjectsClientHandler(nodeType uint32) ObjectsHandler {
	return &gRPCClientHandler{
		nodeType: nodeType,
	}
}

type gRPCClientHandler struct {
	nodeType uint32
	BaseHandler
}

func (g *gRPCClientHandler) CreateCollection(ctx context.Context, collection *Collection) error {
	client, err := RouterGrpc(ctx, g.nodeType)
	if err != nil {
		return err
	}

	_, err = client.CreateCollection(ctx, &CreateCollectionRequest{Collection: collection})
	return err
}

func (g *gRPCClientHandler) GetCollection(ctx context.Context, id string) (*Collection, error) {
	client, err := RouterGrpc(ctx, g.nodeType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetCollection(ctx, &GetCollectionRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return rsp.Collection, err
}

func (g *gRPCClientHandler) ListCollections(ctx context.Context) ([]*Collection, error) {
	client, err := RouterGrpc(ctx, g.nodeType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ListCollections(ctx, &ListCollectionsRequest{})
	if err != nil {
		return nil, err
	}
	return rsp.Collections, err
}

func (g *gRPCClientHandler) DeleteCollection(ctx context.Context, id string) error {
	client, err := RouterGrpc(ctx, g.nodeType)
	if err != nil {
		return err
	}

	_, err = client.DeleteCollection(ctx, &DeleteCollectionRequest{Id: id})
	return err
}

func (g *gRPCClientHandler) PutObject(ctx context.Context, collection string, object *Object, accessSecurityRules *PathAccessRules, indexes []*se.TextIndex, opts PutOptions) (string, error) {
	client, err := RouterGrpc(ctx, g.nodeType)
	if err != nil {
		return "", err
	}

	rsp, err := client.PutObject(auth.SetMetaWithExisting(ctx), &PutObjectRequest{
		Collection:          collection,
		Object:              object,
		Indexes:             indexes,
		AccessSecurityRules: accessSecurityRules,
	})
	if err != nil {
		return "", err
	}

	return rsp.ObjectId, nil
}

func (g *gRPCClientHandler) PatchObject(ctx context.Context, collection string, patch *Patch, opts PatchOptions) error {
	client, err := RouterGrpc(ctx, ServiceTypeHandler)
	if err != nil {
		return err
	}

	_, err = client.PatchObject(auth.SetMetaWithExisting(ctx), &PatchObjectRequest{
		Collection: collection,
		Patch:      patch,
	})
	return err
}

func (g *gRPCClientHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *PathAccessRules, opts MoveOptions) error {
	client, err := RouterGrpc(ctx, ServiceTypeHandler)
	if err != nil {
		return err
	}

	_, err = client.MoveObject(auth.SetMetaWithExisting(ctx), &MoveObjectRequest{
		SourceCollection:    collection,
		ObjectId:            objectID,
		TargetCollection:    targetCollection,
		AccessSecurityRules: accessSecurityRules,
	})
	return err
}

func (g *gRPCClientHandler) GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*Object, error) {
	client, err := RouterGrpc(ctx, ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetObject(auth.SetMetaWithExisting(ctx), &GetObjectRequest{
		Collection: collection,
		ObjectId:   id,
		At:         opts.At,
		InfoOnly:   opts.Info,
	})
	if err != nil {
		return nil, err
	}

	return rsp.Object, nil
}

func (g *gRPCClientHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*Header, error) {
	client, err := RouterGrpc(ctx, ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ObjectInfo(auth.SetMetaWithExisting(ctx), &ObjectInfoRequest{
		Collection: collection,
		ObjectId:   id,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Header, nil
}

func (g *gRPCClientHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	client, err := RouterGrpc(ctx, ServiceTypeHandler)
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(auth.SetMetaWithExisting(ctx), &DeleteObjectRequest{
		Collection: collection,
		ObjectId:   id,
	})
	return err
}

func (g *gRPCClientHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	client, err := RouterGrpc(ctx, ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	stream, err := client.ListObjects(auth.SetMetaWithExisting(ctx), &ListObjectsRequest{
		Offset:     opts.Offset,
		At:         opts.At,
		Collection: collection,
	})
	if err != nil {
		return nil, err
	}

	closer := CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := BrowseFunc(func() (*Object, error) {
		return stream.Recv()
	})

	return NewCursor(browser, closer), nil
}

func (g *gRPCClientHandler) SearchObjects(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error) {
	client, err := RouterGrpc(ctx, ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	stream, err := client.SearchObjects(auth.SetMetaWithExisting(ctx), &SearchObjectsRequest{
		Collection: collection,
		Query:      query,
	})
	if err != nil {
		return nil, err
	}

	closer := CloseFunc(func() error {
		return stream.CloseSend()
	})
	browser := BrowseFunc(func() (*Object, error) {
		return stream.Recv()
	})

	return NewCursor(browser, closer), nil
}
