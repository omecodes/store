package router

import (
	"context"
	"github.com/google/uuid"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/oms"
	"github.com/omecodes/store/pb"
)

type ExecHandler struct {
	BaseHandler
}

func (e *ExecHandler) PutObject(ctx context.Context, object *oms.Object, security *pb.PathAccessRules, opts oms.PutDataOptions) (string, error) {

	id := object.ID()
	if id == "" {
		id = uuid.New().String()
		object.SetID(id)
	}

	accessStore := acl.GetStore(ctx)
	if accessStore == nil {
		log.Info("exec-handler.PutObject: missing access store in context")
		return "", errors.Internal
	}

	err := accessStore.SaveRules(ctx, object.ID(), security)
	if err != nil {
		log.Error("exec-handler.PutObject: failed to save object access security rules", log.Err(err))
		return "", errors.Internal
	}

	storage := oms.Get(ctx)
	if storage == nil {
		log.Error("exec-handler.PutObject: missing storage in context")
		if err2 := accessStore.Delete(ctx, object.ID()); err2 != nil {
			log.Error("exec-handler.PutObject: failed to clear access rules", log.Err(err2))
		}
		return "", errors.Internal
	}

	err = storage.Save(ctx, object)
	if err != nil {
		if err2 := accessStore.Delete(ctx, object.ID()); err2 != nil {
			log.Error("exec-handler.PutObject: failed to clear access rules", log.Err(err2))
		}
		return "", err
	}

	return id, nil
}

func (e *ExecHandler) PatchObject(ctx context.Context, patch *oms.Patch, opts oms.PatchOptions) error {
	storage := oms.Get(ctx)
	if storage == nil {
		log.Info("missing storage in context")
		return errors.Internal
	}
	return storage.Patch(ctx, patch)
}

func (e *ExecHandler) GetObject(ctx context.Context, objectID string, opts oms.GetObjectOptions) (*oms.Object, error) {
	storage := oms.Get(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.Internal
	}

	if opts.Path == "" {
		return storage.Get(ctx, objectID)
	} else {
		return storage.GetAt(ctx, objectID, opts.Path)
	}
}

func (e *ExecHandler) GetObjectHeader(ctx context.Context, objectID string) (*pb.Header, error) {
	storage := oms.Get(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.Internal
	}
	return storage.Info(ctx, objectID)
}

func (e *ExecHandler) DeleteObject(ctx context.Context, objectID string) error {
	storage := oms.Get(ctx)
	if storage == nil {
		log.Info("exec-handler.DeleteObjet: missing DB in context")
		return errors.Internal
	}

	err := storage.Delete(ctx, objectID)
	if err != nil {
		log.Error("exec-handler.DeleteObjet: failed to delete object from storage", log.Err(err))
		return err
	}

	accessStore := acl.GetStore(ctx)
	if accessStore == nil {
		log.Info("exec-handler.DeleteObjet: missing access store in context")
		return errors.Internal
	}

	return accessStore.Delete(ctx, objectID)
}

func (e *ExecHandler) ListObjects(ctx context.Context, opts oms.ListOptions) (*oms.ObjectList, error) {
	storage := oms.Get(ctx)
	if storage == nil {
		log.Info("missing DB in context")
		return nil, errors.Internal
	}
	return storage.List(ctx, opts.Before, opts.Count, opts.Filter)
}
