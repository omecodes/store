package router

import "context"

const (
	handlerTypeParams = 1
	handlerTypePolicy = 2
	handlerTypeExec   = 3
)

type Router interface {
	// GetRoute returns a sequence of handler
	GetRoute(opts ...RouteOption) ObjectsHandler
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

type RouteProviderFunc func(opts ...RouteOption) ObjectsHandler

func (f RouteProviderFunc) GetRoute(opts ...RouteOption) ObjectsHandler {
	return f(opts...)
}

func DefaultRouter() Router {
	return RouteProviderFunc(getRoute)
}

func getRoute(opts ...RouteOption) (handler ObjectsHandler) {
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
