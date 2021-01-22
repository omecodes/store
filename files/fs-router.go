package files

import (
	"context"
)

type ctxRouterProvider struct{}

type Router interface {
	// GetRoute returns a sequence of handler
	GetRoute(opts ...RouteOption) Handler
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

func (f RouteProviderFunc) GetRoute(opts ...RouteOption) Handler {
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

type routesOptions struct {
	skipPolicies   bool
	skipParams     bool
	skipEncryption bool
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

func SkipEncryption() RouteOption {
	return func(r *routesOptions) {
		r.skipEncryption = true
	}
}

func NewRoute(ctx context.Context, opt ...RouteOption) (Handler, error) {
	o := ctx.Value(ctxRouterProvider{})
	if o == nil {
		return DefaultFilesRouter().GetRoute(opt...), nil
	}

	p := o.(RouterProvider)
	router := p.GetRouter(ctx)

	return router.GetRoute(opt...), nil
}
