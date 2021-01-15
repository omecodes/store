package router

type CustomObjectsRouter struct {
	paramsHandler *ParamsObjectsHandler
	policyHandler *PolicyObjectsHandler
	execHandler   ObjectsHandler
}

type objectsHandlersOptions struct {
	params *ParamsObjectsHandler
	policy *PolicyObjectsHandler
}

type ObjectsHandlerOption func(*objectsHandlersOptions)

func WithParamsObjectsHandler(handler *ParamsObjectsHandler) ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.params = handler
	}
}

func WithPolicyObjectsHandler(handler *PolicyObjectsHandler) ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.policy = handler
	}
}

func WithDefaultParamsObjectsHandler() ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.params = &ParamsObjectsHandler{}
	}
}

func WithDefaultPoliciesObjectsHandler() ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.policy = &PolicyObjectsHandler{}
	}
}

func (r *CustomObjectsRouter) GetRoute(opts ...RouteOption) ObjectsHandler {
	var handler ObjectsHandler

	options := routesOptions{}
	for _, o := range opts {
		o(&options)
	}

	if !options.skipExecution && r.execHandler != nil {
		handler = r.execHandler
	} else {
		handler = &dummyHandler{}
	}

	if !options.skipPolicies && r.policyHandler != nil {
		r.policyHandler.next = handler
		handler = r.policyHandler
	} else {
		handler = &BaseObjectsHandler{
			next: handler,
		}
	}

	if !options.skipParams && r.paramsHandler != nil {
		r.paramsHandler.next = handler
		handler = r.paramsHandler
	} else {
		handler = &BaseObjectsHandler{
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
