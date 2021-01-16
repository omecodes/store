package router

import "context"

const (
	handlerTypeParams = 1
	handlerTypePolicy = 2
	handlerTypeExec   = 3
)

type ObjectsRouter interface {
	// GetRoute returns a sequence of handler
	GetRoute(opts ...RouteOption) ObjectsHandler
}

type ObjectsRouterProvider interface {
	//GetRouter returns a router
	GetRouter(ctx context.Context) ObjectsRouter
}

type ObjectsRouterProvideFunc func(ctx context.Context) ObjectsRouter

func (f ObjectsRouterProvideFunc) GetRouter(ctx context.Context) ObjectsRouter {
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

type ObjectsRouteProviderFunc func(opts ...RouteOption) ObjectsHandler

func (f ObjectsRouteProviderFunc) GetRoute(opts ...RouteOption) ObjectsHandler {
	return f(opts...)
}

func DefaultRouter() ObjectsRouter {
	return ObjectsRouteProviderFunc(getObjectsRoute)
}

func getObjectsRoute(opts ...RouteOption) (handler ObjectsHandler) {
	routes := routesOptions{}

	for _, o := range opts {
		o(&routes)
	}

	if !routes.skipExecution {
		handler = &ObjectsExecHandler{}
	} else {
		handler = &dummyHandler{}
	}

	if !routes.skipPolicies {
		handler = &ObjectsPolicyHandler{ObjectsBaseHandler: ObjectsBaseHandler{
			next: handler,
		}}
	}

	if !routes.skipParams {
		handler = &ObjectsParamsHandler{
			ObjectsBaseHandler{next: handler},
		}
	}
	return
}
