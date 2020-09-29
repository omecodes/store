package store

import (
	"context"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/ent"
	"github.com/omecodes/omestore/ent/access"
	"github.com/omecodes/omestore/ent/permission"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/omestore/store/internals"
)

type ctxSettingsDB struct{}
type ctxDataDir struct{}
type ctxAdminPassword struct{}
type ctxCelEnv struct{}
type ctxStore struct{}
type ctxDB struct{}
type ctxSettings struct{}
type ctxUsers struct{}
type ctxAccesses struct{}
type ctxInfo struct{}
type ctxGraftInfo struct{}
type ctxCELPrograms struct{}

type contextUpdaterFunc func(ctx context.Context) context.Context

func getDB(ctx context.Context) *ent.Client {
	o := ctx.Value(ctxDB{})
	if o == nil {
		return nil
	}
	return o.(*ent.Client)
}

func getCelEnv(ctx context.Context) *cel.Env {
	o := ctx.Value(ctxCelEnv{})
	if o == nil {
		return nil
	}
	return o.(*cel.Env)
}

func getAdminPassword(ctx context.Context) (string, bool) {
	o := ctx.Value(ctxAdminPassword{})
	if o == nil {
		return "", false
	}
	return o.(string), true
}

func getStorage(ctx context.Context) pb.Store {
	o := ctx.Value(ctxStore{})
	if o == nil {
		return nil
	}
	return o.(pb.Store)
}

func getAppdata(ctx context.Context) internals.Store {
	o := ctx.Value(ctxSettingsDB{})
	if o == nil {
		return nil
	}
	return o.(internals.Store)
}

func getDataDir(ctx context.Context) string {
	o := ctx.Value(ctxDataDir{})
	if o == nil {
		return ""
	}
	return o.(string)
}

func (s *Server) contextWithDB() contextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxDB{}, s.entDB)
	}
}

func (s *Server) contextWithCelEnv() contextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxCelEnv{}, s.celEnv)
	}
}

func (s *Server) contextWithAdminPassword() contextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxAdminPassword{}, s.adminPassword)
	}
}

func (s *Server) contextWithDataDir() contextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxDataDir{}, s.config.Dir)
	}
}

func (s *Server) contextWithDataDB() contextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxStore{}, s.dataStore)
	}
}

func (s *Server) contextWithAppData() contextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxSettingsDB{}, s.appData)
	}
}

func settings(ctx *context.Context) (*JSON, error) {
	o := (*ctx).Value(ctxSettings{})
	if o != nil {
		return &JSON{Object: o}, nil
	}

	route := getRoute(SkipPoliciesCheck())
	s, err := route.GetSettings(*ctx, pb.SettingsOptions{})
	if err != nil {
		return nil, err
	}

	*ctx = context.WithValue(*ctx, ctxSettings{}, s)
	return s, nil
}

func userInfo(ctx *context.Context, name string) (*ent.User, error) {
	var m map[string]*ent.User

	o := (*ctx).Value(ctxUsers{})
	if o != nil {
		m = o.(map[string]*ent.User)
		if m != nil {
			u, found := m[name]
			if found {
				return u, nil
			}
		}
	}

	if m == nil {
		m = map[string]*ent.User{}
	}

	route := getRoute(SkipPoliciesCheck())
	u, err := route.UserInfo(*ctx, name, pb.UserOptions{})
	if err != nil {
		return nil, err
	}

	m[name] = u
	*ctx = context.WithValue(*ctx, ctxUsers{}, m)
	return u, nil
}

func getAccess(ctx *context.Context, collection string, id string) (*ent.Access, error) {
	var m map[string]*ent.Access

	qid := fmt.Sprintf("%s.%s", collection, id)

	o := (*ctx).Value(ctxAccesses{})
	if o != nil {
		m = o.(map[string]*ent.Access)
		if m != nil {
			u, found := m[qid]
			if found {
				return u, nil
			}
		}
	}

	if m == nil {
		m = map[string]*ent.Access{}
	}

	db := getDB(*ctx)
	if db == nil {
		log.Error("could not find database in context")
		return nil, errors.Internal
	}

	a, err := db.Access.Query().Where(access.ID(qid)).First(*ctx)
	if err != nil && !ent.IsNotFound(err) {
		log.Error("could not get access rules", log.Field("for", qid), log.Err(err))
		return nil, errors.Internal
	}

	if a != nil {
		m[qid] = a
		*ctx = context.WithValue(*ctx, ctxAccesses{}, m)
	}
	return a, nil
}

func getDataInfo(ctx context.Context, collection string, id string) *pb.Info {
	var m map[string]*pb.Info
	qid := fmt.Sprintf("%s.%s", collection, id)
	o := (ctx).Value(ctxInfo{})
	if o != nil {
		m = o.(map[string]*pb.Info)
		if m != nil {
			info, found := m[qid]
			if found {
				return info
			}
		}
	}
	return nil
}

func getGraftInfo(ctx context.Context, collection string, dataID string, id string) *pb.GraftInfo {
	var m map[string]*pb.GraftInfo
	gid := fmt.Sprintf("%s.%s.%s", collection, dataID, id)
	o := (ctx).Value(ctxGraftInfo{})
	if o != nil {
		m = o.(map[string]*pb.GraftInfo)
		if m != nil {
			info, found := m[gid]
			if found {
				return info
			}
		}
	}
	return nil
}

func contextWithDataInfo(ctx context.Context, collection string, id string, info *pb.Info) context.Context {
	var m map[string]*pb.Info
	qid := fmt.Sprintf("%s.%s", collection, id)
	o := (ctx).Value(ctxInfo{})
	if o != nil {
		m = o.(map[string]*pb.Info)
	} else {
		m = map[string]*pb.Info{}
		ctx = context.WithValue(ctx, ctxInfo{}, m)
	}
	m[qid] = info
	return ctx
}

func contextWithGraftInfo(ctx context.Context, collection string, dataID string, id string, info *pb.GraftInfo) context.Context {
	var m map[string]*pb.GraftInfo
	qid := fmt.Sprintf("%s.%s.%s", collection, dataID, id)
	o := (ctx).Value(ctxGraftInfo{})
	if o != nil {
		m = o.(map[string]*pb.GraftInfo)
	} else {
		m = map[string]*pb.GraftInfo{}
		ctx = context.WithValue(ctx, ctxGraftInfo{}, m)
	}
	m[qid] = info
	return ctx
}

func getProgram(ctx *context.Context, rule string) (cel.Program, error) {
	var m map[string]cel.Program

	o := (*ctx).Value(ctxCELPrograms{})
	if o != nil {
		m = o.(map[string]cel.Program)
		if m != nil {
			prg, found := m[rule]
			if found {
				return prg, nil
			}
		}
	}

	if m == nil {
		m = map[string]cel.Program{}
	}

	env := getCelEnv(*ctx)
	if env == nil {
		return nil, errors.Internal
	}

	ast, issues := env.Compile(rule)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	prg, err := env.Program(
		ast,
		cel.Functions(
			&functions.Overload{
				Operator: "acl",
				Binary: func(l ref.Val, r ref.Val) ref.Val {
					if types.StringType != l.Type() {
						return types.ValOrErr(l, "expect first argument to be string")
					}
					if types.StringType != r.Type() {
						return types.ValOrErr(r, "expect second argument to be string")
					}

					uid := l.Value().(string)
					uri := r.Value().(string)

					db := getDB(*ctx)
					perm, err := db.Permission.Query().Where(permission.User(uid), permission.Data(uri)).First(*ctx)
					if err != nil && !ent.IsNotFound(err) {
						log.Error("could not load user permission", log.Err(err))
						return types.DefaultTypeAdapter.NativeToValue(&pb.PermCEL{})
					}

					if perm == nil {
						perm = &ent.Permission{}
					}

					return types.DefaultTypeAdapter.NativeToValue(&pb.PermCEL{
						Read:  perm.ActionRead,
						Write: perm.ActionWrite,
					})
				},
			},
		),
	)
	if err != nil {
		return nil, err
	}

	m[rule] = prg
	*ctx = context.WithValue(*ctx, ctxCELPrograms{}, m)
	return prg, nil
}
