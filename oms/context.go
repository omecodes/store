package oms

import (
	"context"
	"fmt"
	"github.com/omecodes/bome"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"

	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
)

type ctxSettingsDB struct{}
type ctxDataDir struct{}
type ctxAdminPassword struct{}
type ctxStore struct{}

type ctxPermissions struct{}
type ctxUserInfo struct{}

type ctxCELEnv struct{}
type ctxSettings struct{}
type ctxUsers struct{}
type ctxAccesses struct{}
type ctxInfo struct{}
type ctxGraftInfo struct{}
type ctxCELPrograms struct{}
type ctxAuthCEL struct{}
type ctxWorkers struct{}

// ContextUpdater is a convenience for context enriching object
// It take a Context object and return a new one with that contains
// at least the same info as the passed one.
type ContextUpdater interface {
	UpdateContext(ctx context.Context) context.Context
}

type ContextUpdaterFunc func(ctx context.Context) context.Context

func (u ContextUpdaterFunc) UpdateContext(ctx context.Context) context.Context {
	return u(ctx)
}

// WithPermissions creates a context updater that adds permissions store to a context
func WithPermissions(perms PermissionsStore) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxPermissions{}, perms)
	}
}

// WithStore creates a context updater that adds store to a context
func WithStore(store Store) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxStore{}, store)
	}
}

// WithSettings creates a context updater that adds permissions to a context
func WithSettings(settings bome.JSONMap) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxSettingsDB{}, settings)
	}
}

// WithCelEnv creates a context updater that adds CEL env to a context
func WithCelEnv(env *cel.Env) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxCELEnv{}, env)
	}
}

// WithWorkers creates a context updater that adds CEL env to a context
func WithWorkers(infoDB bome.JSONMap) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxWorkers{}, infoDB)
	}
}

// ContextWithUserInfo creates a context updater that adds a Auth info to a context
func ContextWithUserInfo(ctx context.Context, a *Auth) context.Context {
	return context.WithValue(ctx, ctxAuthCEL{}, a)
}

func contextWithDataInfo(ctx context.Context, collection string, id string, info *Info) context.Context {
	var m map[string]*Info
	qid := fmt.Sprintf("%s.%s", collection, id)
	o := (ctx).Value(ctxInfo{})
	if o != nil {
		m = o.(map[string]*Info)
	} else {
		m = map[string]*Info{}
		ctx = context.WithValue(ctx, ctxInfo{}, m)
	}
	m[qid] = info
	return ctx
}

func celEnv(ctx context.Context) *cel.Env {
	o := ctx.Value(ctxCELEnv{})
	if o == nil {
		return nil
	}
	return o.(*cel.Env)
}

func storage(ctx context.Context) Store {
	o := ctx.Value(ctxStore{})
	if o == nil {
		return nil
	}
	return o.(Store)
}

func settings(ctx context.Context) bome.JSONMap {
	o := ctx.Value(ctxSettingsDB{})
	if o == nil {
		return nil
	}
	return o.(bome.JSONMap)
}

func getDataDir(ctx context.Context) string {
	o := ctx.Value(ctxDataDir{})
	if o == nil {
		return ""
	}
	return o.(string)
}

func dataInfo(ctx context.Context, collection string, id string) *Info {
	var m map[string]*Info
	qid := fmt.Sprintf("%s.%s", collection, id)
	o := (ctx).Value(ctxInfo{})
	if o != nil {
		m = o.(map[string]*Info)
		if m != nil {
			info, found := m[qid]
			if found {
				return info
			}
		}
	}
	return nil
}

func graftInfo(ctx context.Context, collection string, dataID string, id string) *GraftInfo {
	var m map[string]*GraftInfo
	gid := fmt.Sprintf("%s.%s.%s", collection, dataID, id)
	o := (ctx).Value(ctxGraftInfo{})
	if o != nil {
		m = o.(map[string]*GraftInfo)
		if m != nil {
			info, found := m[gid]
			if found {
				return info
			}
		}
	}
	return nil
}

func contextWithGraftInfo(ctx context.Context, collection string, dataID string, id string, info *GraftInfo) context.Context {
	var m map[string]*GraftInfo
	qid := fmt.Sprintf("%s.%s.%s", collection, dataID, id)
	o := (ctx).Value(ctxGraftInfo{})
	if o != nil {
		m = o.(map[string]*GraftInfo)
	} else {
		m = map[string]*GraftInfo{}
		ctx = context.WithValue(ctx, ctxGraftInfo{}, m)
	}
	m[qid] = info
	return ctx
}

func getPermissionsStore(ctx *context.Context) PermissionsStore {
	o := (*ctx).Value(ctxPermissions{})
	if o == nil {
		return nil
	}
	return o.(PermissionsStore)
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

	env := celEnv(*ctx)
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

					permissions := getPermissionsStore(ctx)
					auth := authInfo(*ctx)

					perm, err := permissions.Get(uid, uri, auth.Uid)
					if err != nil {
						log.Error("could not load user permission", log.Err(err))
						return types.DefaultTypeAdapter.NativeToValue(&Perm{})
					}

					if perm == nil {
						perm = &Permission{}
					}

					return types.DefaultTypeAdapter.NativeToValue(&Perm{
						Read:  perm.Actions&AllowedTo_read == AllowedTo_read,
						Write: perm.Actions&AllowedTo_write == AllowedTo_write,
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

func authInfo(ctx context.Context) *Auth {
	o := ctx.Value(ctxAuthCEL{})
	if o == nil {
		return nil
	}
	return o.(*Auth)
}

func getWorkerInfoDB(ctx context.Context) bome.JSONMap {
	o := ctx.Value(ctxWorkers{})
	if o == nil {
		return nil
	}
	return o.(bome.JSONMap)
}
