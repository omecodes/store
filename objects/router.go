package objects

import "context"

type Router interface {
	// GetHandler returns a sequence of handler
	GetHandler(opts ...RouteOption) Handler
}

type RouterProvider interface {
	//GetRouter returns a router
	GetRouter(ctx context.Context) Router
}

type RouterProvideFunc func(ctx context.Context) Router

func (f RouterProvideFunc) GetRouter(ctx context.Context) Router {
	return f(ctx)
}

type RouteProviderFunc func(opts ...RouteOption) Handler

func (f RouteProviderFunc) GetHandler(opts ...RouteOption) Handler {
	return f(opts...)
}

func DefaultRouter() Router {
	return RouteProviderFunc(getHandler)
}

func getHandler(opts ...RouteOption) (handler Handler) {
	routes := routesOptions{}

	for _, o := range opts {
		o(&routes)
	}

	handler = &ExecHandler{}

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
