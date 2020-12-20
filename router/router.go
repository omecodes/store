package router

import "context"

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
		handler = &policyHandler{base: base{
			next: handler,
		}}
	}

	if !routes.skipParams {
		handler = &paramsHandler{
			base{next: handler},
		}
	}
	return
}
