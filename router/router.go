package router

import "context"

const (
	handlerTypeParams = 1
	handlerTypePolicy = 2
	handlerTypeExec   = 3
)

type CustomRouter struct {
	paramsHandler *BaseHandler
	policyHandler *BaseHandler
	execHandler   Handler
}

type handlersOptions struct {
	params *BaseHandler
	policy *BaseHandler
}

type HandlerOption func(*handlersOptions)

func WithParamsHandler(handler *BaseHandler) HandlerOption {
	return func(options *handlersOptions) {
		options.params = handler
	}
}

func WithPolicyHandler(handler *BaseHandler) HandlerOption {
	return func(options *handlersOptions) {
		options.policy = handler
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
		}
		handler = r.policyHandler
	}

	if !options.skipParams {
		if r.paramsHandler != nil {
			r.paramsHandler.next = handler
		}
		handler = r.paramsHandler
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

type GetRouterFunc func(opts ...RouteOption) Handler

func (f GetRouterFunc) GetRoute(opts ...RouteOption) Handler {
	return f(opts...)
}

func DefaultRouter() Router {
	return GetRouterFunc(getRoute)
}

func getRoute(opts ...RouteOption) (handler Handler) {
	routes := routesOptions{}

	for _, o := range opts {
		o(&routes)
	}

	if !routes.skipExecution {
		handler = &execHandler{}
	} else {
		handler = &dummyHandler{}
	}

	if !routes.skipPolicies {
		handler = &policyHandler{BaseHandler: BaseHandler{
			next: handler,
		}}
	}

	if !routes.skipParams {
		handler = &paramsHandler{
			BaseHandler{next: handler},
		}
	}
	return
}
