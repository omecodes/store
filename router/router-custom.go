package router

type CustomObjectsRouter struct {
	paramsHandler *ObjectsParamsHandler
	policyHandler *ObjectsPolicyHandler
	execHandler   ObjectsHandler
}

type objectsHandlersOptions struct {
	params *ObjectsParamsHandler
	policy *ObjectsPolicyHandler
}

type ObjectsHandlerOption func(*objectsHandlersOptions)

func WithObjectsParamsHandler(handler *ObjectsParamsHandler) ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.params = handler
	}
}

func WithObjectsPolicyHandler(handler *ObjectsPolicyHandler) ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.policy = handler
	}
}

func WithDefaultObjectsParamsHandler() ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.params = &ObjectsParamsHandler{}
	}
}

func WithDefaultObjectsPolicyHandler() ObjectsHandlerOption {
	return func(options *objectsHandlersOptions) {
		options.policy = &ObjectsPolicyHandler{}
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
		handler = &objectsDummyHandler{}
	}

	if !options.skipPolicies && r.policyHandler != nil {
		r.policyHandler.next = handler
		handler = r.policyHandler
	} else {
		handler = &ObjectsBaseHandler{
			next: handler,
		}
	}

	if !options.skipParams && r.paramsHandler != nil {
		r.paramsHandler.next = handler
		handler = r.paramsHandler
	} else {
		handler = &ObjectsBaseHandler{
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
