package router

import (
	"context"

	"github.com/google/cel-go/cel"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/omestore/oms"
)

type ctxSettingsDB struct{}
type ctxDataDir struct{}
type ctxAdminPassword struct{}
type ctxStore struct{}

type ctxAccessStore struct{}
type ctxCELPolicyEnv struct{}
type ctxCELSearchEnv struct{}
type ctxUserInfo struct{}

type ctxSettings struct{}
type ctxUsers struct{}
type ctxAccesses struct{}
type ctxInfo struct{}
type ctxObjectHeader struct{}
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

// WithAccessStore creates a context updater that adds permissions store to a context
func WithAccessStore(store oms.AccessStore) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxAccessStore{}, store)
	}
}

// WithObjectsStore creates a context updater that adds store to a context
func WithObjectsStore(objects oms.Objects) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxStore{}, objects)
	}
}

// WithSettings creates a context updater that adds permissions to a context
func WithSettings(settings *bome.JSONMap) ContextUpdaterFunc {
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

// WithUserInfo creates a context updater that adds a Auth info to a context
func WithUserInfo(ctx context.Context, a *oms.Auth) context.Context {
	return context.WithValue(ctx, ctxAuthCEL{}, a)
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

func authInfo(ctx context.Context) *oms.Auth {
	o := ctx.Value(ctxAuthCEL{})
	if o == nil {
		return nil
	}
	return o.(*oms.Auth)
}

func workersDB(ctx context.Context) *bome.JSONMap {
	o := ctx.Value(ctxWorkers{})
	if o == nil {
		return nil
	}
	return o.(*bome.JSONMap)
}

func accessStore(ctx context.Context) oms.AccessStore {
	o := ctx.Value(ctxAccessStore{})
	if o == nil {
		return nil
	}
	return o.(oms.AccessStore)
}

func getObjectHeader(ctx *context.Context, objectID string) (*oms.Header, error) {
	var m map[string]*oms.Header
	o := (*ctx).Value(ctxObjectHeader{})
	if o != nil {
		m = o.(map[string]*oms.Header)
		if m != nil {
			header, found := m[objectID]
			if found {
				return header, nil
			}
		}
	}

	if m == nil {
		m = map[string]*oms.Header{}
	}

	route := Route(SkipParamsCheck(), SkipPoliciesCheck())
	header, err := route.GetObjectHeader(*ctx, objectID)
	if err != nil {
		return nil, err
	}

	m[objectID] = header
	*ctx = context.WithValue(*ctx, ctxObjectHeader{}, m)
	return header, nil
}

func loadProgramForAccessValidation(ctx *context.Context, expression string) (cel.Program, error) {
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

	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	m[expression] = prg
	*ctx = context.WithValue(*ctx, ctxCELAclPrograms{}, m)
	return prg, nil
}

func loadProgramForSearch(ctx *context.Context, expression string) (cel.Program, error) {
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
