package objects

type CustomObjectsRouter struct {
	paramsHandler *ParamsHandler
	policyHandler *PolicyHandler
	execHandler   ObjectsHandler
}

type objectsHandlersOptions struct {
	params *ParamsHandler
	policy *PolicyHandler
}

type ObjectsHandlerOption func(*objectsHandlersOptions)

func WithObjectsParamsHandler(handler *ParamsHandler) ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.params = handler
	}
}

func WithObjectsPolicyHandler(handler *PolicyHandler) ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.policy = handler
	}
}

func WithDefaultObjectsParamsHandler() ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.params = &ParamsHandler{}
	}
}

func WithDefaultObjectsPolicyHandler() ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.policy = &PolicyHandler{}
	}
}

func (r *CustomObjectsRouter) GetRoute(opts ...RouteOption) ObjectsHandler {
	var handler ObjectsHandler

	options := routesOptions{}
	for _, o := range opts {
		o(&options)
	}

	handler = r.execHandler

	if !options.skipPolicies && r.policyHandler != nil {
		r.policyHandler.next = handler
		handler = r.policyHandler
	} else {
		handler = &BaseHandler{
			next: handler,
		}
	}

	if !options.skipParams && r.paramsHandler != nil {
		r.paramsHandler.next = handler
		handler = r.paramsHandler
	} else {
		handler = &BaseHandler{
			next: handler,
		}
	}
	return handler
}

func NewCustomObjectsRouter(exec ObjectsHandler, opts ...ObjectsHandlerOption) *CustomObjectsRouter {
	var options objectsHandlersOptions
	for _, opt := range opts {
		opt(&options)
	}

	return &CustomObjectsRouter{
		paramsHandler: options.params,
		policyHandler: options.policy,
		execHandler:   exec,
	}
}
