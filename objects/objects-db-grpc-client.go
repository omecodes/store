package objects

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/clients"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/meta"
	"github.com/omecodes/store/pb"
	"io"
	"strconv"
)

func NewStoreGrpcClient() Objects {
	return &dbClient{}
}

type dbClient struct {
	pb.UnimplementedHandlerUnitServer
}

func (d *dbClient) Save(ctx context.Context, object *pb.Object, index ...*pb.Index) error {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return err
	}

	_, err = objects.PutObject(ctx, &pb.PutObjectRequest{
		Object:  object,
		Indexes: index,
	})
	return err
}

func (d *dbClient) Patch(ctx context.Context, patch *pb.Patch) error {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return err
	}

	_, err = objects.UpdateObject(ctx, &pb.UpdateObjectRequest{
		Patch: patch,
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

func (d *dbClient) List(ctx context.Context, filter ObjectFilter, opts ListOptions) (*pb.ObjectList, error) {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}

	stream, err := objects.ListObjects(ctx, &pb.ListObjectsRequest{
		Before: opts.Before,
		After:  opts.After,
		Count:  uint32(opts.Count),
		At:     opts.At,
	})

	defer func() {
		if err := stream.CloseSend(); err != nil {
			log.Error("Objects client • close gRPC stream with error", log.Err(err))
		}
	}()

	result := &pb.ObjectList{}

	md, err := stream.Header()
	if err != nil {
		log.Error("Objects client • stream › could not get metadata", log.Err(err))
		return nil, errors.Internal
	}

	count, err := strconv.Atoi(md.Get(meta.Count)[0])
	if err != nil {
		log.Error("Objects client • stream › unreadable metadata 'count'", log.Err(err))
		return nil, errors.Internal
	}

	result.Before, err = strconv.ParseInt(md.Get(meta.Before)[0], 10, 64)
	if err != nil {
		log.Error("Objects client • stream › unreadable metadata 'count'", log.Err(err))
		return nil, errors.Internal
	}

	for len(result.Objects) < count {
		object, err := stream.Recv()
		if err != nil {
			if io.EOF == err {
				break
			}
			log.Error("Objects client • stream › could not get remaining objects", log.Err(err))
			return nil, errors.Internal
		}

		if filter != nil {
			allowed, err := filter.Filter(object)
			if err != nil {
				if errors.IsForbidden(err) {
					continue
				}
				return nil, err
			}

			if !allowed {
				continue
			}
		}
		result.Objects = append(result.Objects, object)
	}

	return result, nil
}

func (d *dbClient) Get(ctx context.Context, objectID string, opts GetObjectOptions) (*pb.Object, error) {
	objects, err := clients.RouterGrpc(ctx, common.ServiceTypeObjects)
	if err != nil {
		return nil, err
	}

	rsp, err := objects.GetObject(ctx, &pb.GetObjectRequest{ObjectId: objectID})
	if err != nil {
		return nil, err
	}

	return rsp.Object, nil
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
