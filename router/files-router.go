package router

import "context"

type FilesRouter interface {
	// GetRoute returns a sequence of handler
	GetRoute(opts ...RouteOption) FilesHandler
}

type FilesRouterProvider interface {
	//GetRouter returns a router
	GetRouter(ctx context.Context) FilesRouter
}

type FilesRouterProvideFunc func(ctx context.Context) ObjectsRouter

func (f FilesRouterProvideFunc) GetRouter(ctx context.Context) ObjectsRouter {
	return f(ctx)
}

type FilesRouteProviderFunc func(opts ...RouteOption) FilesHandler

func (f FilesRouteProviderFunc) GetRoute(opts ...RouteOption) FilesHandler {
	return f(opts...)
}

func DefaultFilesRouter() FilesRouter {
	return FilesRouteProviderFunc(getFilesRoute)
}

func getFilesRoute(opts ...RouteOption) (handler FilesHandler) {
	routes := routesOptions{}

	for _, o := range opts {
		o(&routes)
	}

	if !routes.skipEncryption {
		handler = &FilesEncryptionHandler{FilesBaseHandler: FilesBaseHandler{
			next: handler,
		}}
	}

	if !routes.skipPolicies {
		handler = &FilesPolicyHandler{FilesBaseHandler: FilesBaseHandler{
			next: handler,
		}}
	}

	if !routes.skipParams {
		handler = &FilesParamsHandler{
			FilesBaseHandler{next: handler},
		}
	}

	return
}
