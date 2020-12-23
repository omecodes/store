package router

import "context"

const (
	handlerTypeParams = 1
	handlerTypePolicy = 2
	handlerTypeExec   = 3
)

type CustomRouter struct {
	paramsHandler *ParamsHandler
	policyHandler *PolicyHandler
	execHandler   Handler
}

type handlersOptions struct {
	params *ParamsHandler
	policy *PolicyHandler
}

type HandlerOption func(*handlersOptions)

func WithParamsHandler(handler *ParamsHandler) HandlerOption {
	return func(options *handlersOptions) {
		options.params = handler
	}
}

func WithPolicyHandler(handler *PolicyHandler) HandlerOption {
	return func(options *handlersOptions) {
		options.policy = handler
	}
}

func WithDefaultParamsHandler() HandlerOption {
	return func(options *handlersOptions) {
		options.params = &ParamsHandler{}
	}
}

func WithDefaultPoliciesHandler() HandlerOption {
	return func(options *handlersOptions) {
		options.policy = &PolicyHandler{}
	}
}

func (r *CustomRouter) GetRoute(opts ...RouteOption) Handler {
	var handler Handler

	options := routesOptions{}
	for _, o := range opts {
		o(&options)
	}

	if !options.skipExecution {
		if r.execHandler != nil {
			handler = r.execHandler
		} else {
			handler = &BaseHandler{next: &dummyHandler{}}
		}
	} else {
		handler = &dummyHandler{}
	}

	if !options.skipPolicies {
		if r.policyHandler != nil {
			r.policyHandler.next = handler
			handler = r.policyHandler
		}
	}

	if !options.skipParams {
		if r.paramsHandler != nil {
			r.paramsHandler.next = handler
			handler = r.paramsHandler
		}
	}

	return handler
}

func NewCustomRouter(exec Handler, opts ...HandlerOption) *CustomRouter {
	var options handlersOptions
	for _, opt := range opts {
		opt(&options)
	}

	return &CustomRouter{
		paramsHandler: options.params,
		policyHandler: options.policy,
		execHandler:   exec,
	}
}

type Router interface {
	// GetRoute returns a sequence of handler
	GetRoute(opts ...RouteOption) Handler
}

type Provider interface {
	//GetRouter returns a router
	GetRouter(ctx context.Context) Router
}

type ProviderFunc func(ctx context.Context) Router

func (f ProviderFunc) GetRouter(ctx context.Context) Router {
	return f(ctx)
}

type routesOptions struct {
	skipPolicies  bool
	skipParams    bool
	skipExecution bool
}

type RouteOption func(*routesOptions)

func SkipParamsCheck() RouteOption {
	return func(r *routesOptions) {
		r.skipParams = true
	}
}

func SkipPoliciesCheck() RouteOption {
	return func(r *routesOptions) {
		r.skipPolicies = true
	}
}

func SkipExec() RouteOption {
	return func(r *routesOptions) {
		r.skipExecution = true
	}
}

type RouteProviderFunc func(opts ...RouteOption) Handler

func (f RouteProviderFunc) GetRoute(opts ...RouteOption) Handler {
	return f(opts...)
}

func DefaultRouter() Router {
	return RouteProviderFunc(getRoute)
}

func getRoute(opts ...RouteOption) (handler Handler) {
	routes := routesOptions{}

	for _, o := range opts {
		o(&routes)
	}

	if !routes.skipExecution {
		handler = &ExecHandler{}
	} else {
		handler = &dummyHandler{}
	}

	if !routes.skipPolicies {
		handler = &PolicyHandler{BaseHandler: BaseHandler{
			next: handler,
		}}
	}

	if !routes.skipParams {
		handler = &ParamsHandler{
			BaseHandler{next: handler},
		}
	}
	return
}
