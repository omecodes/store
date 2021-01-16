package router

import (
	"context"
	"fmt"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/pb"
)

type ObjectsPolicyHandler struct {
	ObjectsBaseHandler
}

func (p *ObjectsPolicyHandler) isAdmin(ctx context.Context) bool {
	authCEL := auth.Get(ctx)
	if authCEL == nil {
		return false
	}
	return authCEL.Uid == "admin"
}

func (p *ObjectsPolicyHandler) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.ObjectsBaseHandler.CreateCollection(ctx, collection)
}

func (p *ObjectsPolicyHandler) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	if !p.isAdmin(ctx) {
		return nil, errors.Forbidden
	}
	return p.ObjectsBaseHandler.GetCollection(ctx, id)
}

func (p *ObjectsPolicyHandler) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	if !p.isAdmin(ctx) {
		return nil, errors.Forbidden
	}
	return p.ObjectsBaseHandler.ListCollections(ctx)
}

func (p *ObjectsPolicyHandler) DeleteCollection(ctx context.Context, id string) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.ObjectsBaseHandler.DeleteCollection(ctx, id)
}

func (p *ObjectsPolicyHandler) PutObject(ctx context.Context, collection string, object *pb.Object, accessSecurityRules *pb.PathAccessRules, indexes []*pb.TextIndex, opts pb.PutOptions) (string, error) {
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
	return p.ObjectsBaseHandler.PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func (p *ObjectsPolicyHandler) GetObject(ctx context.Context, collection string, id string, opts pb.GetOptions) (*pb.Object, error) {
	err := assetActionAllowedOnObject(&ctx, collection, id, pb.AllowedTo_read, opts.At)
	if err != nil {
		return nil, err
	}
	return p.ObjectsBaseHandler.GetObject(ctx, collection, id, opts)
}

func (p *ObjectsPolicyHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts pb.PatchOptions) error {
	err := assetActionAllowedOnObject(&ctx, collection, patch.ObjectId, pb.AllowedTo_delete, "")
	if err != nil {
		return err
	}

	return p.ObjectsBaseHandler.PatchObject(ctx, collection, patch, opts)
}

func (p *ObjectsPolicyHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *pb.PathAccessRules, opts pb.MoveOptions) error {
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

func (p *ObjectsPolicyHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error) {
	err := assetActionAllowedOnObject(&ctx, collection, id, pb.AllowedTo_read, "")
	if err != nil {
		return nil, err
	}

	return p.ObjectsBaseHandler.GetObjectHeader(ctx, collection, id)
}

func (p *ObjectsPolicyHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	err := assetActionAllowedOnObject(&ctx, collection, id, pb.AllowedTo_delete, "")
	if err != nil {
		return err
	}
	return p.ObjectsBaseHandler.DeleteObject(ctx, collection, id)
}

func (p *ObjectsPolicyHandler) ListObjects(ctx context.Context, collection string, opts pb.ListOptions) (*pb.Cursor, error) {
	var err error

	cursor, err := p.ObjectsBaseHandler.ListObjects(ctx, collection, opts)
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

func (p *ObjectsPolicyHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*pb.Cursor, error) {
	if collection == "" || query == nil {
		return nil, errors.BadInput
	}

	cursor, err := p.ObjectsBaseHandler.SearchObjects(ctx, collection, query)
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
