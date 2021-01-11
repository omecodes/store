package router

import (
	"context"
	"fmt"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/pb"
)

type PolicyHandler struct {
	BaseHandler
}

func (p *PolicyHandler) isAdmin(ctx context.Context) bool {
	authCEL := auth.Get(ctx)
	if authCEL == nil {
		return false
	}
	return authCEL.Uid == "admin"
}

func (p *PolicyHandler) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.BaseHandler.CreateCollection(ctx, collection)
}

func (p *PolicyHandler) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	if !p.isAdmin(ctx) {
		return nil, errors.Forbidden
	}
	return p.BaseHandler.GetCollection(ctx, id)
}

func (p *PolicyHandler) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	if !p.isAdmin(ctx) {
		return nil, errors.Forbidden
	}
	return p.BaseHandler.ListCollections(ctx)
}

func (p *PolicyHandler) DeleteCollection(ctx context.Context, id string) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.BaseHandler.DeleteCollection(ctx, id)
}

func (p *PolicyHandler) PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.Index, opts pb.PutOptions) (string, error) {
	ai := auth.Get(ctx)
	if ai == nil {
		return "", errors.Forbidden
	}

	if accessSecurityRules == nil {
		collectionInfo, err := p.next.GetCollection(ctx, collection)
		if err != nil {
			return "", err
		}
		accessSecurityRules = collectionInfo.DefaultAccessSecurityRules
	}

	docRules := accessSecurityRules.AccessRules["$"]
	if docRules == nil {
		docRules = &pb.AccessRules{}
		accessSecurityRules.AccessRules["$"] = docRules
	}

	userDefaultRule := fmt.Sprintf("auth.uid=='%s'", ai.Uid)
	if len(docRules.Read) == 0 {
		docRules.Read = append(docRules.Read, userDefaultRule, "auth.worker || auth.uid=='admin'")
	} else {
		docRules.Read = append(docRules.Read, "auth.worker || auth.uid=='admin'")
	}

	if len(docRules.Write) == 0 {
		docRules.Write = append(docRules.Write, userDefaultRule, "auth.worker || auth.uid=='admin'")
	} else {
		docRules.Write = append(docRules.Write, "auth.worker", "auth.uid=='admin'")
	}

	if len(docRules.Delete) == 0 {
		docRules.Delete = append(docRules.Delete, userDefaultRule)
	}

	object.Header.CreatedBy = ai.Uid
	return p.BaseHandler.PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func (p *PolicyHandler) GetObject(ctx context.Context, collection string, id string, opts pb.GetOptions) (*pb.Object, error) {
	err := assetActionAllowedOnObject(&ctx, collection, id, pb.AllowedTo_read, opts.At)
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObject(ctx, collection, id, opts)
}

func (p *PolicyHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts pb.PatchOptions) error {
	err := assetActionAllowedOnObject(&ctx, collection, patch.ObjectId, pb.AllowedTo_delete, "")
	if err != nil {
		return err
	}

	return p.BaseHandler.PatchObject(ctx, collection, patch, opts)
}

func (p *PolicyHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts pb.MoveOptions) error {
	err := assetActionAllowedOnObject(&ctx, collection, objectID, pb.AllowedTo_read, "")
	if err != nil {
		return err
	}

	err = assetActionAllowedOnObject(&ctx, collection, objectID, pb.AllowedTo_delete, "")
	if err != nil {
		return err
	}

	if accessSecurityRules == nil {
		collectionInfo, err := p.next.GetCollection(ctx, targetCollection)
		if err != nil {
			return err
		}
		accessSecurityRules = collectionInfo.DefaultAccessSecurityRules
	}

	return p.next.MoveObject(ctx, collection, objectID, targetCollection, accessSecurityRules, opts)
}

func (p *PolicyHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error) {
	err := assetActionAllowedOnObject(&ctx, collection, id, pb.AllowedTo_read, "")
	if err != nil {
		return nil, err
	}

	return p.BaseHandler.GetObjectHeader(ctx, collection, id)
}

func (p *PolicyHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	err := assetActionAllowedOnObject(&ctx, collection, id, pb.AllowedTo_delete, "")
	if err != nil {
		return err
	}
	return p.BaseHandler.DeleteObject(ctx, collection, id)
}

func (p *PolicyHandler) ListObjects(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error) {
	var err error

	cursor, err := p.BaseHandler.ListObjects(ctx, collection, opts)
	if err != nil {
		return nil, err
	}

	cursorBrowser := cursor.GetBrowser()
	browser := pb.BrowseFunc(func() (*pb.Object, error) {
		for {
			o, err := cursorBrowser.Browse()
			if err != nil {
				return nil, err
			}

			err = assetActionAllowedOnObject(&ctx, collection, o.Header.Id, pb.AllowedTo_read, opts.At)
			if err != nil {
				if err == errors.Unauthorized {
					continue
				}
				return nil, err
			}

			return o, nil
		}
	})

	cursor.SetBrowser(browser)
	return cursor, nil
}

func (p *PolicyHandler) SearchObjects(ctx context.Context, collection string, exp *pb.BooleanExp) (*pb.Cursor, error) {
	if collection == "" || exp == nil {
		return nil, errors.BadInput
	}

	cursor, err := p.BaseHandler.SearchObjects(ctx, collection, exp)
	if err != nil {
		return nil, err
	}

	cursorBrowser := cursor.GetBrowser()
	browser := pb.BrowseFunc(func() (*pb.Object, error) {
		for {
			o, err := cursorBrowser.Browse()
			if err != nil {
				return nil, err
			}

			err = assetActionAllowedOnObject(&ctx, collection, o.Header.Id, pb.AllowedTo_read, "")
			if err != nil {
				if err == errors.Unauthorized {
					continue
				}
				return nil, err
			}

			return o, nil
		}
	})

	cursor.SetBrowser(browser)
	return cursor, nil
}
