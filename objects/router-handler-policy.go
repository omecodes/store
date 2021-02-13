package objects

import (
	"context"
	"fmt"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/store/auth"
	se "github.com/omecodes/store/search-engine"
)

type PolicyHandler struct {
	BaseHandler
}

func (p *PolicyHandler) isAdmin(ctx context.Context) bool {
	user := auth.Get(ctx)
	if user == nil {
		return false
	}
	return user.Name == "admin"
}

func (p *PolicyHandler) CreateCollection(ctx context.Context, collection *Collection) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden
	}
	return p.BaseHandler.CreateCollection(ctx, collection)
}

func (p *PolicyHandler) GetCollection(ctx context.Context, id string) (*Collection, error) {
	user := auth.Get(ctx)
	if user == nil {
		return nil, errors.Forbidden
	}

	if user.Name == "" || user.Name != "admin" && user.Access != "client" {
		return nil, errors.Forbidden
	}

	return p.BaseHandler.GetCollection(ctx, id)
}

func (p *PolicyHandler) ListCollections(ctx context.Context) ([]*Collection, error) {
	user := auth.Get(ctx)
	if user == nil {
		return nil, errors.Forbidden
	}

	if user.Name == "" || user.Name != "admin" && user.Access != "client" {
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

func (p *PolicyHandler) PutObject(ctx context.Context, collection string, object *Object, accessSecurityRules *PathAccessRules, indexes []*se.TextIndex, opts PutOptions) (string, error) {
	user := auth.Get(ctx)
	if user == nil {
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
		docRules = &AccessRules{}
		accessSecurityRules.AccessRules["$"] = docRules
	}

	userDefaultRule := fmt.Sprintf("user.name=='%s'", user.Name)
	readPerm := &auth.Permission{
		Name:        "default-readers",
		Label:       "Readers",
		Description: "In addition of creator, admin and workers are allowed to read every objects",
		Rule:        "user.access=='worker' || user.name=='admin'",
	}

	if len(docRules.Read) == 0 {
		readPerm.Rule = userDefaultRule + " || user.name=='worker' || user.name=='admin'"
		docRules.Read = append(docRules.Read, readPerm)
	} else {
		docRules.Read = append(docRules.Read, readPerm)
	}

	writePerm := &auth.Permission{
		Name:        "default-readers",
		Label:       "Readers",
		Description: "In addition of creator, admin and workers are allowed to write every objects",
		Rule:        "user.access=='worker' || user.name=='admin'",
	}
	if len(docRules.Write) == 0 {
		readPerm.Rule = userDefaultRule + " || user.access=='worker' || user.name=='admin'"
		docRules.Write = append(docRules.Write, writePerm)
	} else {
		docRules.Write = append(docRules.Write, writePerm)
	}

	deletePerm := &auth.Permission{
		Name:        "default-readers",
		Label:       "Readers",
		Description: "In addition of creator, admin and workers are allowed to write every objects",
		Rule:        "user.access=='worker' || user.name=='admin'",
	}
	if len(docRules.Delete) == 0 {
		docRules.Delete = append(docRules.Delete, deletePerm)
	}

	object.Header.CreatedBy = user.Name
	return p.BaseHandler.PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func (p *PolicyHandler) GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*Object, error) {
	err := assetActionAllowedOnObject(&ctx, collection, id, auth.AllowedTo_read, opts.At)
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObject(ctx, collection, id, opts)
}

func (p *PolicyHandler) PatchObject(ctx context.Context, collection string, patch *Patch, opts PatchOptions) error {
	err := assetActionAllowedOnObject(&ctx, collection, patch.ObjectId, auth.AllowedTo_delete, "")
	if err != nil {
		return err
	}
	return p.BaseHandler.PatchObject(ctx, collection, patch, opts)
}

func (p *PolicyHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *PathAccessRules, opts MoveOptions) error {
	err := assetActionAllowedOnObject(&ctx, collection, objectID, auth.AllowedTo_read, "")
	if err != nil {
		return err
	}

	err = assetActionAllowedOnObject(&ctx, collection, objectID, auth.AllowedTo_delete, "")
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

func (p *PolicyHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*Header, error) {
	err := assetActionAllowedOnObject(&ctx, collection, id, auth.AllowedTo_read, "")
	if err != nil {
		return nil, err
	}

	return p.BaseHandler.GetObjectHeader(ctx, collection, id)
}

func (p *PolicyHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	err := assetActionAllowedOnObject(&ctx, collection, id, auth.AllowedTo_delete, "")
	if err != nil {
		return err
	}
	return p.BaseHandler.DeleteObject(ctx, collection, id)
}

func (p *PolicyHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	var err error

	cursor, err := p.BaseHandler.ListObjects(ctx, collection, opts)
	if err != nil {
		return nil, err
	}

	cursorBrowser := cursor.GetBrowser()
	browser := BrowseFunc(func() (*Object, error) {
		for {
			o, err := cursorBrowser.Browse()
			if err != nil {
				return nil, err
			}

			err = assetActionAllowedOnObject(&ctx, collection, o.Header.Id, auth.AllowedTo_read, opts.At)
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

func (p *PolicyHandler) SearchObjects(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error) {
	if collection == "" || query == nil {
		return nil, errors.BadInput
	}

	cursor, err := p.BaseHandler.SearchObjects(ctx, collection, query)
	if err != nil {
		return nil, err
	}

	cursorBrowser := cursor.GetBrowser()
	browser := BrowseFunc(func() (*Object, error) {
		for {
			o, err := cursorBrowser.Browse()
			if err != nil {
				return nil, err
			}

			err = assetActionAllowedOnObject(&ctx, collection, o.Header.Id, auth.AllowedTo_read, "")
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
