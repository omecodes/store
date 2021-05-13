package objects

import (
	"context"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common/cenv"
	pb "github.com/omecodes/store/gen/go/proto"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type PolicyHandler struct {
	BaseHandler
}

func (p *PolicyHandler) isAdmin(ctx context.Context) bool {
	user := auth.Get(ctx)
	if user == nil {
		return false
	}

	logs.Info("context user", logs.Details("name", user.Name))
	return user.Name == "admin"
}

type celParams struct {
	auth *pb.User
	data *pb.Header
	app  *pb.ClientApp
}

func (p *PolicyHandler) evaluate(_ context.Context, state *celParams, rule string) (bool, error) {
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
		cel.Types(&pb.User{}, &pb.ClientApp{}, &pb.Header{}),
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

func (p *PolicyHandler) assetActionAllowedOnObject(ctx context.Context, collection string, objectID string, action int32, path string) error {
	return nil
}

func (p *PolicyHandler) CreateCollection(ctx context.Context, collection *pb.Collection) error {
	clientApp := auth.App(ctx)
	if !p.isAdmin(ctx) {
		if clientApp == nil {
			return errors.Forbidden("not allowed to create collections")
		}
	}
	return p.BaseHandler.CreateCollection(ctx, collection)
}

func (p *PolicyHandler) GetCollection(ctx context.Context, id string) (*pb.Collection, error) {
	if !p.isAdmin(ctx) {
		clientApp := auth.App(ctx)
		if clientApp == nil {
			return nil, errors.Forbidden("not allowed to get collection info")
		}
	}
	return p.BaseHandler.GetCollection(ctx, id)
}

func (p *PolicyHandler) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	if !p.isAdmin(ctx) {
		clientApp := auth.App(ctx)
		if clientApp == nil {
			return nil, errors.Forbidden("not allowed to list collections")
		}
	}
	return p.BaseHandler.ListCollections(ctx)
}

func (p *PolicyHandler) DeleteCollection(ctx context.Context, id string) error {
	clientApp := auth.App(ctx)
	if !p.isAdmin(ctx) || clientApp == nil {
		return errors.Forbidden("not allowed to delete collections")
	}
	return p.BaseHandler.DeleteCollection(ctx, id)
}

func (p *PolicyHandler) PutObject(ctx context.Context, collection string, object *pb.Object, authorizedUsers *pb.PathAccessRules, indexes []*pb.TextIndex, opts PutOptions) (string, error) {
	user := auth.Get(ctx)
	if user == nil {
		return "", errors.Forbidden("access forbidden")
	}

	// if no security rules are provided, collection security rules will be used
	if authorizedUsers == nil {
		collectionInfo, err := p.next.GetCollection(ctx, collection)
		if err != nil {
			logs.Error("could not get collection", logs.Err(err))
			return "", err
		}
		authorizedUsers = collectionInfo.DefaultActionAuthorizedUsers
	}
	docRules := authorizedUsers.AccessRules["$"]
	if docRules == nil {
		docRules = &pb.ObjectActionsUsers{}
		authorizedUsers.AccessRules["$"] = docRules
	}

	object.Header.CreatedBy = user.Name

	return p.BaseHandler.PutObject(ctx, collection, object, authorizedUsers, indexes, opts)
}

func (p *PolicyHandler) GetObject(ctx context.Context, collection string, id string, opts GetOptions) (*pb.Object, error) {
	err := p.assetActionAllowedOnObject(ctx, collection, id, 0, opts.At)
	if err != nil {
		return nil, err
	}
	return p.BaseHandler.GetObject(ctx, collection, id, opts)
}

func (p *PolicyHandler) PatchObject(ctx context.Context, collection string, patch *pb.Patch, opts PatchOptions) error {
	err := p.assetActionAllowedOnObject(ctx, collection, patch.ObjectId, 0, "")
	if err != nil {
		return err
	}
	return p.BaseHandler.PatchObject(ctx, collection, patch, opts)
}

func (p *PolicyHandler) MoveObject(ctx context.Context, collection string, objectID string, targetCollection string, authorizedUsers *pb.PathAccessRules, opts MoveOptions) error {
	err := p.assetActionAllowedOnObject(ctx, collection, objectID, 0, "")
	if err != nil {
		return err
	}

	err = p.assetActionAllowedOnObject(ctx, collection, objectID, 0, "")
	if err != nil {
		return err
	}

	if authorizedUsers == nil {
		collectionInfo, err := p.next.GetCollection(ctx, targetCollection)
		if err != nil {
			return err
		}
		authorizedUsers = collectionInfo.DefaultActionAuthorizedUsers
	}

	return p.next.MoveObject(ctx, collection, objectID, targetCollection, authorizedUsers, opts)
}

func (p *PolicyHandler) GetObjectHeader(ctx context.Context, collection string, id string) (*pb.Header, error) {
	err := p.assetActionAllowedOnObject(ctx, collection, id, 0, "")
	if err != nil {
		return nil, err
	}

	return p.BaseHandler.GetObjectHeader(ctx, collection, id)
}

func (p *PolicyHandler) DeleteObject(ctx context.Context, collection string, id string) error {
	err := p.assetActionAllowedOnObject(ctx, collection, id, 0, "")
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
	browser := BrowseFunc(func() (*pb.Object, error) {
		return cursorBrowser.Browse()
	})

	cursor.SetBrowser(browser)
	return cursor, nil
}

func (p *PolicyHandler) SearchObjects(ctx context.Context, collection string, query *pb.SearchQuery) (*Cursor, error) {
	cursor, err := p.BaseHandler.SearchObjects(ctx, collection, query)
	if err != nil {
		return nil, err
	}

	cursorBrowser := cursor.GetBrowser()
	browser := BrowseFunc(func() (*pb.Object, error) {
		for {
			o, err := cursorBrowser.Browse()
			if err != nil {
				return nil, err
			}
			return o, nil
		}
	})

	cursor.SetBrowser(browser)
	return cursor, nil
}
