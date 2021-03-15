package objects

import (
	"context"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"strings"

	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common/cenv"
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

type celParams struct {
	auth *auth.User
	data *Header
	app  *auth.ClientApp
}

func (p *PolicyHandler) evaluate(ctx context.Context, state *celParams, rule string) (bool, error) {
	if rule == "" || rule == "false" || rule == "(false)" {
		return false, nil
	}

	if rule == "true" || rule == "(true)" {
		return true, nil
	}

	prg, err := cenv.GetProgram(rule,
		cel.Declarations(
			decls.NewVar("user", decls.NewObjectType("User")),
			decls.NewVar("app", decls.NewObjectType("ClientApp")),
			decls.NewVar("data", decls.NewObjectType("Header")),
			decls.NewFunction("now",
				decls.NewOverload(
					"now_uint",
					[]*expr.Type{}, decls.Uint,
				),
			),
		),
		cel.Types(&auth.User{}, &auth.ClientApp{}, &Header{}),
	)
	if err != nil {
		return false, err
	}

	vars := map[string]interface{}{
		"user": state.auth,
		"app":  state.app,
		"data": state.data,
	}

	out, details, err := prg.Eval(vars)
	if err != nil {
		logs.Error("cel execution", logs.Details("details", details))
		return false, err
	}

	return out.Value().(bool), nil
}

func (p *PolicyHandler) getAccessRule(ctx context.Context, collection string, objectID string, action auth.AllowedTo, path string) (string, error) {
	accessStore := GetACLStore(ctx)
	if accessStore == nil {
		logs.Error("ACL-Read-Check: missing access store in context")
		return "", errors.Internal("missing ACL store")
	}

	ruleCollection, err := accessStore.GetRules(ctx, collection, objectID)
	if err != nil {
		return "", err
	}

	if path == "" {
		path = "$"
	}

	rules, found := ruleCollection.AccessRules[path]
	if !found {
		logs.Error("ACL: could not find access security rule", logs.Details("object", objectID), logs.Details("path", path))
		rules, found = ruleCollection.AccessRules["$"]
		if !found {
			return "", errors.Forbidden("no security rules found")
		}
	}

	var actionRules []*auth.Permission
	switch action {
	case auth.AllowedTo_read:
		actionRules = rules.Read
	case auth.AllowedTo_write:
		actionRules = rules.Write
	case auth.AllowedTo_delete:
		actionRules = rules.Delete

	default:
		logs.Error("ACL: no rule for this action", logs.Details("action", action.String()))
		return "", errors.Unsupported("unsupported ACL action")
	}

	var formattedRules []string
	for _, exp := range actionRules {
		formattedRules = append(formattedRules, "("+exp.Rule+")")
	}
	rule := strings.Join(formattedRules, " || ")

	return rule, nil
}

func (p *PolicyHandler) assetActionAllowedOnObject(ctx context.Context, collection string, objectID string, action auth.AllowedTo, path string) error {
	header, err := p.next.GetObjectHeader(ctx, collection, objectID)
	if err != nil {
		return err
	}

	authCEL := auth.Get(ctx)
	if authCEL == nil {
		authCEL = &auth.User{}
	}

	s := &celParams{
		data: header,
		auth: authCEL,
		app:  auth.App(ctx),
	}

	rule, err := p.getAccessRule(ctx, collection, objectID, action, path)
	if err != nil {
		return err
	}

	allowed, err := p.evaluate(ctx, s, rule)
	if err != nil {
		logs.Error("failed to evaluate access rule", logs.Err(err))
		return errors.Internal("unable to evaluate access permission")
	}

	if !allowed {
		return errors.Unauthorized("not authorized")
	}

	return nil
}

func (p *PolicyHandler) CreateCollection(ctx context.Context, collection *Collection) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden("access forbidden")
	}
	return p.BaseHandler.CreateCollection(ctx, collection)
}

func (p *PolicyHandler) GetCollection(ctx context.Context, id string) (*Collection, error) {
	user := auth.Get(ctx)
	if user == nil {
		return nil, errors.Forbidden("no user provided")
	}

	app := auth.App(ctx)
	if user.Name != "admin" && app == nil {
		return nil, errors.Forbidden("collections are only readable within a registered application")
	}

	if user.Name == "" {
		return nil, errors.Forbidden("Resource access refused", errors.Details{Key: "user", Value: user})
	}

	return p.BaseHandler.GetCollection(ctx, id)
}

func (p *PolicyHandler) ListCollections(ctx context.Context) ([]*Collection, error) {
	user := auth.Get(ctx)
	if user == nil {
		return nil, errors.Forbidden("no user provided")
	}

	app := auth.App(ctx)
	if user.Name != "admin" && app == nil {
		return nil, errors.Forbidden("collections are only readable within a registered application")
	}

	if user.Name == "" {
		return nil, errors.Forbidden("Resource access refused", errors.Details{Key: "user", Value: user})
	}

	return p.BaseHandler.ListCollections(ctx)
}

func (p *PolicyHandler) DeleteCollection(ctx context.Context, id string) error {
	if !p.isAdmin(ctx) {
		return errors.Forbidden("access forbidden")
	}
	return p.BaseHandler.DeleteCollection(ctx, id)
}

func (p *PolicyHandler) PutObject(ctx context.Context, collection string, object *Object, accessSecurityRules *PathAccessRules, indexes []*se.TextIndex, opts PutOptions) (string, error) {
	user := auth.Get(ctx)
	if user == nil {
		return "", errors.Forbidden("access forbidden")
	}

	// if no security rules are provided, collection security rules will be used
	if accessSecurityRules == nil {
		collectionInfo, err := p.next.GetCollection(ctx, collection)
		if err != nil {
			logs.Error("could not get collection", logs.Err(err))
			return "", err
		}
		accessSecurityRules = collectionInfo.DefaultAccessSecurityRules
	}
	docRules := accessSecurityRules.AccessRules["$"]
	if docRules == nil {
		docRules = &AccessRules{}
		accessSecurityRules.AccessRules["$"] = docRules
	}

	creatorRule := fmt.Sprintf("user.name=='%s' && user.access=='client'", user.Name)

	readPerm := &auth.Permission{
		Name:        "default-readers",
		Label:       "Readers",
		Description: "In addition of creator, admin and workers are allowed to read the object",
		Rule:        "user.name=='admin'",
	}
	if len(docRules.Read) == 0 {
		readPerm.Rule = creatorRule + " || user.name=='worker' || user.name=='admin'"
	}
	docRules.Read = append(docRules.Read, readPerm)

	writePerm := &auth.Permission{
		Name:        "default-readers",
		Label:       "Readers",
		Description: "In addition of creator, admin and workers are allowed to edit the object",
		Rule:        "user.name=='admin'",
	}
	if len(docRules.Write) == 0 {
		writePerm.Rule = creatorRule + " || user.access=='worker' || user.name=='admin'"
	}
	docRules.Write = append(docRules.Write, writePerm)

	deletePerm := &auth.Permission{
		Name:        "default-readers",
		Label:       "Readers",
		Description: "In addition of creator, admin and workers are allowed to write the object",
		Rule:        "user.name=='admin'",
	}
	if len(docRules.Delete) == 0 {
		deletePerm.Rule = creatorRule + " || user.name=='admin'"
	}
	docRules.Delete = append(docRules.Delete, deletePerm)

	object.Header.CreatedBy = user.Name

	return p.BaseHandler.PutObject(ctx, collection, object, accessSecurityRules, indexes, opts)
}

func (p *PolicyHandler) GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*Object, error) {
	err := p.assetActionAllowedOnObject(ctx, collection, id, auth.AllowedTo_read, opts.At)
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObject(ctx, collection, id, opts)
}

func (p *PolicyHandler) PatchObject(ctx context.Context, collection string, patch *Patch, opts PatchOptions) error {
	err := p.assetActionAllowedOnObject(ctx, collection, patch.ObjectId, auth.AllowedTo_write, "")
	if err != nil {
		return err
	}
	return p.BaseHandler.PatchObject(ctx, collection, patch, opts)
}

func (p *PolicyHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, accessSecurityRules *PathAccessRules, opts MoveOptions) error {
	err := p.assetActionAllowedOnObject(ctx, collection, objectID, auth.AllowedTo_read, "")
	if err != nil {
		return err
	}

	err = p.assetActionAllowedOnObject(ctx, collection, objectID, auth.AllowedTo_delete, "")
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
	err := p.assetActionAllowedOnObject(ctx, collection, id, auth.AllowedTo_read, "")
	if err != nil {
		return nil, err
	}

	return p.BaseHandler.GetObjectHeader(ctx, collection, id)
}

func (p *PolicyHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	err := p.assetActionAllowedOnObject(ctx, collection, id, auth.AllowedTo_delete, "")
	if err != nil {
		return err
	}
	return p.BaseHandler.DeleteObject(ctx, collection, id)
}

func (p *PolicyHandler) ListObjects(ctx context.Context, collection string, opts ListOptions) (*Cursor, error) {
	var err error

	cursor, err := p.next.ListObjects(ctx, collection, opts)
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

			authCEL := auth.Get(ctx)
			if authCEL == nil {
				authCEL = &auth.User{}
			}

			s := &celParams{
				data: o.Header,
				auth: authCEL,
			}

			rule, err := p.getAccessRule(ctx, collection, o.Header.Id, auth.AllowedTo_read, opts.At)
			if err != nil {
				return nil, err
			}

			allowed, err := p.evaluate(ctx, s, rule)
			if err != nil {
				return nil, errors.Internal("unable to evaluate access permission")
			}

			if !allowed {
				continue
			}
			return o, nil
		}
	})

	cursor.SetBrowser(browser)
	return cursor, nil
}

func (p *PolicyHandler) SearchObjects(ctx context.Context, collection string, query *se.SearchQuery) (*Cursor, error) {
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

			authCEL := auth.Get(ctx)
			if authCEL == nil {
				authCEL = &auth.User{}
			}

			s := &celParams{
				data: o.Header,
				auth: authCEL,
			}

			rule, err := p.getAccessRule(ctx, collection, o.Header.Id, auth.AllowedTo_read, "")
			if err != nil {
				return nil, err
			}

			allowed, err := p.evaluate(ctx, s, rule)
			if err != nil {
				logs.Error("failed to evaluate access rule", logs.Err(err))
				return nil, errors.Internal("unable to evaluate access permission")
			}

			if !allowed {
				continue
			}
			return o, nil
		}
	})

	cursor.SetBrowser(browser)
	return cursor, nil
}
