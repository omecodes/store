package router

import (
	"context"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
	"github.com/omecodes/omestore/services/acl"

	"github.com/google/cel-go/cel"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
)

type ctxSettingsDB struct{}
type ctxStore struct{}

type ctxAccessStore struct{}
type ctxCELPolicyEnv struct{}
type ctxCELSearchEnv struct{}

type ctxObjectHeader struct{}
type ctxCELAclPrograms struct{}
type ctxCELSearchPrograms struct{}
type ctxWorkers struct{}
type ctxRouterProvider struct{}

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
func WithAccessStore(store acl.Store) ContextUpdaterFunc {
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
func WithSettings(settings oms.SettingsManager) ContextUpdaterFunc {
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
func WithWorkers(infoDB *bome.JSONMap) ContextUpdaterFunc {
	return func(parent context.Context) context.Context {
		return context.WithValue(parent, ctxWorkers{}, infoDB)
	}
}

// WithRouterProvider updates context by adding a RouterProvider object in its values
func WithRouterProvider(ctx context.Context, p Provider) context.Context {
	return context.WithValue(ctx, ctxRouterProvider{}, p)
}

func CELPolicyEnv(ctx context.Context) *cel.Env {
	o := ctx.Value(ctxCELPolicyEnv{})
	if o == nil {
		return nil
	}
	return o.(*cel.Env)
}

func CELSearchEnv(ctx context.Context) *cel.Env {
	o := ctx.Value(ctxCELSearchEnv{})
	if o == nil {
		return nil
	}
	return o.(*cel.Env)
}

func Objects(ctx context.Context) oms.Objects {
	o := ctx.Value(ctxStore{})
	if o == nil {
		return nil
	}
	return o.(oms.Objects)
}

func Settings(ctx context.Context) oms.SettingsManager {
	o := ctx.Value(ctxSettingsDB{})
	if o == nil {
		return nil
	}
	return o.(oms.SettingsManager)
}

func ACLStore(ctx context.Context) acl.Store {
	o := ctx.Value(ctxAccessStore{})
	if o == nil {
		return nil
	}
	return o.(acl.Store)
}

func GetObjectHeader(ctx *context.Context, objectID string) (*pb.Header, error) {
	var m map[string]*pb.Header
	o := (*ctx).Value(ctxObjectHeader{})
	if o != nil {
		m = o.(map[string]*pb.Header)
		if m != nil {
			header, found := m[objectID]
			if found {
				return header, nil
			}
		}
	}

	if m == nil {
		m = map[string]*pb.Header{}
	}

	route, err := NewRoute(*ctx, SkipParamsCheck(), SkipPoliciesCheck())
	if err != nil {
		return nil, err
	}

	header, err := route.GetObjectHeader(*ctx, objectID)
	if err != nil {
		return nil, err
	}

	m[objectID] = header
	*ctx = context.WithValue(*ctx, ctxObjectHeader{}, m)
	return header, nil
}

func LoadProgramForAccessValidation(ctx *context.Context, expression string) (cel.Program, error) {
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

	env := CELPolicyEnv(*ctx)
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

func LoadProgramForSearch(ctx *context.Context, expression string) (cel.Program, error) {
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

	env := CELSearchEnv(*ctx)
	if env == nil {
		return nil, errors.Internal
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		log.Error("failed to compile expression", log.Err(issues.Err()))
		return nil, errors.BadInput
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	m[expression] = prg
	*ctx = context.WithValue(*ctx, ctxCELAclPrograms{}, m)
	return prg, nil
}

func NewRoute(ctx context.Context, opt ...RouteOption) (Handler, error) {
	o := ctx.Value(ctxRouterProvider{})
	if o == nil {
		return nil, errors.New("no router provider")
	}

	p := o.(Provider)
	router := p.GetRouter(ctx)

	return router.GetRoute(opt...), nil
}
