package oms

import (
	"bytes"
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/clients"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/meta"
	"github.com/omecodes/store/pb"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"strconv"
)

func NewStoreGrpcClient() Objects {
	return &dbClient{}
}

type dbClient struct {
	pb.UnimplementedHandlerUnitServer
}

func (d *dbClient) Save(ctx context.Context, object *Object, index ...*pb.Index) error {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(object.GetContent())
	if err != nil {
		return err
	}

	_, err = objects.PutObject(ctx, &pb.PutObjectRequest{
		Header:  object.Header(),
		Data:    string(data),
		Indexes: index,
	})
	return err
}

func (d *dbClient) Patch(ctx context.Context, patch *Patch) error {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(patch.GetContent())
	if err != nil {
		return err
	}

	_, err = objects.UpdateObject(ctx, &pb.UpdateObjectRequest{
		ObjectId: patch.GetObjectID(),
		Data:     data,
	})
	return err
}

func (d *dbClient) Delete(ctx context.Context, objectID string) error {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return err
	}

	_, err = objects.DeleteObject(ctx, &pb.DeleteObjectRequest{
		ObjectId: objectID,
	})
	return err
}

func (d *dbClient) List(ctx context.Context, before int64, count int, filter ObjectFilter) (*ObjectList, error) {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}

	stream, err := objects.ListObjects(ctx, &pb.ListObjectsRequest{
		Before: before,
		Count:  uint32(count),
	})

	defer func() {
		if err := stream.CloseSend(); err != nil {
			log.Error("Objects client • close gRPC stream with error", log.Err(err))
		}
	}()

	result := &ObjectList{}

	md, err := stream.Header()
	if err != nil {
		log.Error("Objects client • stream › could not get metadata", log.Err(err))
		return nil, errors.Internal
	}

	result.Count, err = strconv.Atoi(md.Get(meta.Count)[0])
	if err != nil {
		log.Error("Objects client • stream › unreadable metadata 'count'", log.Err(err))
		return nil, errors.Internal
	}

	result.Before, err = strconv.ParseInt(md.Get(meta.Before)[0], 10, 64)
	if err != nil {
		log.Error("Objects client • stream › unreadable metadata 'count'", log.Err(err))
		return nil, errors.Internal
	}

	for len(result.Objects) < result.Count {
		do, err := stream.Recv()
		if err != nil {
			log.Error("Objects client • stream › could not get remaining objects", log.Err(err))
			return nil, errors.Internal
		}

		object := NewObject()
		object.SetHeader(do.Header)
		object.SetContent(bytes.NewBuffer(do.Data))
		result.Objects = append(result.Objects, object)
	}

	return result, nil
}

func (d *dbClient) ListAt(ctx context.Context, path string, before int64, count int, filter ObjectFilter) (*ObjectList, error) {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}

	outMD := metadata.MD{}
	outMD.Set(meta.At, path)
	newCtx := metadata.NewOutgoingContext(ctx, outMD)

	stream, err := objects.ListObjects(newCtx, &pb.ListObjectsRequest{
		Before: before,
		Count:  uint32(count),
	})
	defer func() {
		if err := stream.CloseSend(); err != nil {
			log.Error("Objects client • close gRPC stream with error", log.Err(err))
		}
	}()

	result := &ObjectList{}

	md, err := stream.Header()
	if err != nil {
		log.Error("Objects client • stream › could not get metadata", log.Err(err))
		return nil, errors.Internal
	}

	result.Count, err = strconv.Atoi(md.Get(meta.Count)[0])
	if err != nil {
		log.Error("Objects client • stream › unreadable metadata 'count'", log.Err(err))
		return nil, errors.Internal
	}

	result.Before, err = strconv.ParseInt(md.Get(meta.Before)[0], 10, 64)
	if err != nil {
		log.Error("Objects client • stream › unreadable metadata 'count'", log.Err(err))
		return nil, errors.Internal
	}

	for len(result.Objects) < result.Count {
		do, err := stream.Recv()
		if err != nil {
			log.Error("Objects client • stream › could not get remaining objects", log.Err(err))
			return nil, errors.Internal
		}

		object := NewObject()
		object.SetHeader(do.Header)
		object.SetContent(bytes.NewBuffer(do.Data))
		result.Objects = append(result.Objects, object)
	}

	return result, nil
}

func (d *dbClient) Get(ctx context.Context, objectID string) (*Object, error) {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.GetObject(ctx, &pb.GetObjectRequest{ObjectId: objectID})
	if err != nil {
		return nil, err
	}

	object := &Object{}
	object.SetHeader(rsp.Data.Header)
	object.SetContent(bytes.NewBuffer(rsp.Data.Data))

	return object, nil
}

func (d *dbClient) GetAt(ctx context.Context, objectID string, path string) (*Object, error) {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}

	md := metadata.MD{}
	md.Set(meta.At, path)
	newCtx := metadata.NewOutgoingContext(ctx, md)
	rsp, err := objects.GetObject(newCtx, &pb.GetObjectRequest{ObjectId: objectID})
	if err != nil {
		return nil, err
	}

	object := &Object{}
	object.SetHeader(rsp.Data.Header)
	object.SetContent(bytes.NewBuffer(rsp.Data.Data))

	return object, nil
}

func (d *dbClient) Info(ctx context.Context, objectID string) (*pb.Header, error) {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.ObjectInfo(ctx, &pb.ObjectInfoRequest{ObjectId: objectID})
	if err != nil {
		return nil, err
	}
	return rsp.Header, nil
}

func (d *dbClient) Clear() error {
	return errors.Forbidden
}
