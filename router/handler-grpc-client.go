package router

import (
	"bytes"
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/clients"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/oms"
	"github.com/omecodes/store/pb"
	"io"
	"io/ioutil"
)

// NewGRPCClientHandler creates a router Handler that embed that calls a gRPC service to perform final actions
func NewGRPCClientHandler(nodeType uint32) Handler {
	return &gRPCClientHandler{
		nodeType: nodeType,
	}
}

type gRPCClientHandler struct {
	nodeType uint32
	BaseHandler
}

func (g *gRPCClientHandler) PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {
	client, err := clients.RouterGrpc(ctx, g.nodeType)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(object.GetContent())
	if err != nil {
		log.Error("could not read object content", log.Err(err))
		return "", errors.BadInput
	}

	rsp, err := client.PutObject(auth.SetMetaWithExisting(ctx), &pb.PutObjectRequest{
		AccessSecurityRules: security,
		Header:              object.Header(),
		Data:                string(data),
		Indexes:             opts.Indexes,
	})
	if err != nil {
		return "", err
	}

	return rsp.ObjectId, nil
}

func (g *gRPCClientHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(patch.GetContent())
	if err != nil {
		log.Error("could not read object content", log.Err(err))
		return errors.BadInput
	}

	_, err = client.UpdateObject(auth.SetMetaWithExisting(ctx), &pb.UpdateObjectRequest{
		ObjectId: patch.GetObjectID(),
		Path:     patch.Path(),
		Data:     string(data),
	})
	return err
}

func (g *gRPCClientHandler) GetObject(ctx context.Context, id string, opts oms.GetObjectOptions) (*oms.Object, error) {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetObject(auth.SetMetaWithExisting(ctx), &pb.GetObjectRequest{
		ObjectId: id,
		Path:     opts.Path,
	})
	if err != nil {
		return nil, err
	}

	o := oms.NewObject()
	o.SetHeader(rsp.Data.Header)
	o.SetContent(bytes.NewBuffer(rsp.Data.Data))

	return o, nil
}

func (g *gRPCClientHandler) GetObjectHeader(ctx context.Context, id string) (*pb.Header, error) {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ObjectInfo(auth.SetMetaWithExisting(ctx), &pb.ObjectInfoRequest{
		ObjectId: id,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Header, nil
}

func (g *gRPCClientHandler) DeleteObject(ctx context.Context, id string) error {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(auth.SetMetaWithExisting(ctx), &pb.DeleteObjectRequest{
		ObjectId: id,
	})
	return err
}

func (g *gRPCClientHandler) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	stream, err := client.ListObjects(auth.SetMetaWithExisting(ctx), &pb.ListObjectsRequest{
		Before: opts.Before,
		Count:  uint32(opts.Count),
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stream.CloseSend(); err != nil {
			log.Error("gRPC Router Handler • error while closing stream", log.Err(err))
		}
	}()

	var objects []*oms.Object

	for {
		dataObject, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		o := oms.NewObject()
		o.SetHeader(dataObject.Header)
		o.SetContent(bytes.NewBuffer(dataObject.Data))

		if opts.Filter != nil {
			allowed, err := opts.Filter.Filter(o)
			if err != nil {
				if err == errors.Unauthorized || err == errors.Forbidden {
					continue
				}
				return nil, err
			}

			if !allowed {
				continue
			}
			// this is repeated in case the object content is consumed during filtering
			o.SetContent(bytes.NewBuffer(dataObject.Data))
		}
		objects = append(objects, o)
		if len(objects) == opts.Count {
			break
		}
	}

	return &oms.ObjectList{
		Before:  opts.Before,
		Count:   len(objects),
		Objects: objects,
	}, nil
}

func (g *gRPCClientHandler) SearchObjects(ctx context.Context, params oms.SearchParams, opts oms.SearchOptions) (*oms.ObjectList, error) {
	client, err := clients.RouterGrpc(ctx, common.ServiceTypeHandler)
	if err != nil {
		return nil, err
	}

	stream, err := client.SearchObjects(auth.SetMetaWithExisting(ctx), &pb.SearchObjectsRequest{
		Before:          opts.Before,
		Count:           uint32(opts.Count),
		MatchExpression: params.MatchedExpression,
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stream.CloseSend(); err != nil {
			log.Error("gRPC Router Handler • error while closing stream", log.Err(err))
		}
	}()

	var objects []*oms.Object

	for {
		dataObject, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		o := oms.NewObject()
		o.SetHeader(dataObject.Header)
		o.SetContent(bytes.NewBuffer(dataObject.Data))
		objects = append(objects, o)
		if len(objects) == opts.Count {
			break
		}
	}

	return &oms.ObjectList{
		Before:  opts.Before,
		Count:   len(objects),
		Objects: objects,
	}, nil
}
