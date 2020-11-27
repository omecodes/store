package router

import (
	"context"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"

	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/oms"
)

type ctxSettingsDB struct{}
type ctxDataDir struct{}
type ctxAdminPassword struct{}
type ctxStore struct{}

type ctxPermissions struct{}
type ctxUserInfo struct{}

type ctxCELPolicyEnv struct{}
type ctxCELSearchEnv struct{}
type ctxSettings struct{}
type ctxUsers struct{}
type ctxAccesses struct{}
type ctxInfo struct{}
type ctxGraftInfo struct{}
type ctxCELAclPrograms struct{}
type ctxCELSearchPrograms struct{}
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
func WithPermissions(perms oms.PermissionsStore) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxPermissions{}, perms)
	}
}

// WithObjectsStore creates a context updater that adds store to a context
func WithObjectsStore(objects oms.Objects) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxStore{}, objects)
	}
}

// WithSettings creates a context updater that adds permissions to a context
func WithSettings(settings bome.JSONMap) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxSettingsDB{}, settings)
	}
}

// WithCelPolicyEnv creates a context updater that adds CEL env used to evaluate acl
func WithCelPolicyEnv(env *cel.Env) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxCELPolicyEnv{}, env)
	}
}

// WithCelEnv creates a context updater that adds CEL env to filter search results
func WithCelSearchEnv(env *cel.Env) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxCELSearchEnv{}, env)
	}
}

// WithWorkers creates a context updater that adds CEL env to a context
func WithWorkers(infoDB bome.JSONMap) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxWorkers{}, infoDB)
	}
}

// ContextWithUserInfo creates a context updater that adds a Auth info to a context
func ContextWithUserInfo(ctx context.Context, a *oms.Auth) context.Context {
	return context.WithValue(ctx, ctxAuthCEL{}, a)
}

func contextWithDataInfo(ctx context.Context, id string, info *oms.Info) context.Context {
	var m map[string]*oms.Info
	o := (ctx).Value(ctxInfo{})
	if o != nil {
		m = o.(map[string]*oms.Info)
	} else {
		m = map[string]*oms.Info{}
		ctx = context.WithValue(ctx, ctxInfo{}, m)
	}
	m[id] = info
	return ctx
}

func celPolicyEnv(ctx context.Context) *cel.Env {
	o := ctx.Value(ctxCELPolicyEnv{})
	if o == nil {
		return nil
	}
	return o.(*cel.Env)
}

func celSearchEnv(ctx context.Context) *cel.Env {
	o := ctx.Value(ctxCELSearchEnv{})
	if o == nil {
		return nil
	}
	return o.(*cel.Env)
}

func storage(ctx context.Context) oms.Objects {
	o := ctx.Value(ctxStore{})
	if o == nil {
		return nil
	}
	return o.(oms.Objects)
}

func settings(ctx context.Context) *bome.JSONMap {
	o := ctx.Value(ctxSettingsDB{})
	if o == nil {
		return nil
	}
	return o.(*bome.JSONMap)
}

func getDataDir(ctx context.Context) string {
	o := ctx.Value(ctxDataDir{})
	if o == nil {
		return ""
	}
	return o.(string)
}

func dataInfo(ctx context.Context, id string) *oms.Info {
	var m map[string]*oms.Info
	o := (ctx).Value(ctxInfo{})
	if o != nil {
		m = o.(map[string]*oms.Info)
		if m != nil {
			info, found := m[id]
			if found {
				return info
			}
		}
	}
	return nil
}

func getPermissionsStore(ctx *context.Context) oms.PermissionsStore {
	o := (*ctx).Value(ctxPermissions{})
	if o == nil {
		return nil
	}
	return o.(oms.PermissionsStore)
}

func getACLProgram(ctx *context.Context, expression string) (cel.Program, error) {
	var m map[string]cel.Program

	o := (*ctx).Value(ctxCELAclPrograms{})
	if o != nil {
		m = o.(map[string]cel.Program)
		if m != nil {
			prg, found := m[expression]
			if found {
				return prg, nil
			}
		}
	}

	if m == nil {
		m = map[string]cel.Program{}
	}

	env := celPolicyEnv(*ctx)
	if env == nil {
		return nil, errors.Internal
	}

	ast, issues := env.Compile(expression)
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
						return types.DefaultTypeAdapter.NativeToValue(&oms.Perm{})
					}

					if perm == nil {
						perm = &oms.Permission{}
					}

					return types.DefaultTypeAdapter.NativeToValue(&oms.Perm{
						Read:  perm.Actions&oms.AllowedTo_read == oms.AllowedTo_read,
						Write: perm.Actions&oms.AllowedTo_write == oms.AllowedTo_write,
					})
				},
			},
		),
	)
	if err != nil {
		return nil, err
	}

	m[expression] = prg
	*ctx = context.WithValue(*ctx, ctxCELAclPrograms{}, m)
	return prg, nil
}

func getSearchProgram(ctx *context.Context, expression string) (cel.Program, error) {
	var m map[string]cel.Program

	o := (*ctx).Value(ctxCELSearchPrograms{})
	if o != nil {
		m = o.(map[string]cel.Program)
		if m != nil {
			prg, found := m[expression]
			if found {
				return prg, nil
			}
		}
	}

	if m == nil {
		m = map[string]cel.Program{}
	}

	env := celSearchEnv(*ctx)
	if env == nil {
		return nil, errors.Internal
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	m[expression] = prg
	*ctx = context.WithValue(*ctx, ctxCELAclPrograms{}, m)
	return prg, nil
}

func authInfo(ctx context.Context) *oms.Auth {
	o := ctx.Value(ctxAuthCEL{})
	if o == nil {
		return nil
	}
	return o.(*oms.Auth)
}

func getWorkerInfoDB(ctx context.Context) *bome.JSONMap {
	o := ctx.Value(ctxWorkers{})
	if o == nil {
		return nil
	}
	return o.(*bome.JSONMap)
}
