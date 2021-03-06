package files

import (
	"context"
)

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

func DefaultFilesRouter() Router {
	return RouteProviderFunc(getRoute)
}

func getRoute(opts ...RouteOption) (handler Handler) {
	routes := routesOptions{}

	for _, o := range opts {
		o(&routes)
	}

	handler = &ExecHandler{}

	if !routes.skipEncryption {
		handler = &EncryptionHandler{BaseHandler: BaseHandler{
			next: handler,
		}}
	}

	if !routes.skipPolicies {
		handler = &ACLHandler{BaseHandler: BaseHandler{
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

type routesOptions struct {
	skipPolicies   bool
	skipParams     bool
	skipEncryption bool
}

type RouteOption func(*routesOptions)

func GetRouteHandler(ctx context.Context, opt ...RouteOption) Handler {
	o := ctx.Value(ctxRouterProvider{})
	if o == nil {
		return DefaultFilesRouter().GetHandler(opt...)
	}

	p := o.(RouterProvider)
	router := p.GetRouter(ctx)

	return router.GetHandler(opt...)
}
