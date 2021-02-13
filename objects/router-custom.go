package objects

type CustomObjectsRouter struct {
	paramsHandler *ParamsHandler
	policyHandler *PolicyHandler
	execHandler   Handler
}

type objectsHandlersOptions struct {
	params *ParamsHandler
	policy *PolicyHandler
}

type HandlerOption func(*objectsHandlersOptions)

func WithObjectsParamsHandler(handler *ParamsHandler) HandlerOption {
	return func(options *objectsHandlersOptions) {
		options.params = handler
	}
}

func WithObjectsPolicyHandler(handler *PolicyHandler) HandlerOption {
	return func(options *objectsHandlersOptions) {
		options.policy = handler
	}
}

func WithDefaultObjectsParamsHandler() HandlerOption {
	return func(options *objectsHandlersOptions) {
		options.params = &ParamsHandler{}
	}
}

func WithDefaultObjectsPolicyHandler() HandlerOption {
	return func(options *objectsHandlersOptions) {
		options.policy = &PolicyHandler{}
	}
}

func (r *CustomObjectsRouter) GetHandler(opts ...RouteOption) Handler {
	var handler Handler

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

func NewCustomObjectsRouter(exec Handler, opts ...HandlerOption) *CustomObjectsRouter {
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
