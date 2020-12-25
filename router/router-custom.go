package router

type CustomRouter struct {
	paramsHandler *ParamsHandler
	policyHandler *PolicyHandler
	execHandler   Handler
}

type handlersOptions struct {
	params *ParamsHandler
	policy *PolicyHandler
}

type HandlerOption func(*handlersOptions)

func WithParamsHandler(handler *ParamsHandler) HandlerOption {
	return func(options *handlersOptions) {
		options.params = handler
	}
}

func WithPolicyHandler(handler *PolicyHandler) HandlerOption {
	return func(options *handlersOptions) {
		options.policy = handler
	}
}

func WithDefaultParamsHandler() HandlerOption {
	return func(options *handlersOptions) {
		options.params = &ParamsHandler{}
	}
}

func WithDefaultPoliciesHandler() HandlerOption {
	return func(options *handlersOptions) {
		options.policy = &PolicyHandler{}
	}
}

func (r *CustomRouter) GetRoute(opts ...RouteOption) Handler {
	var handler Handler

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

func NewCustomRouter(exec Handler, opts ...HandlerOption) *CustomRouter {
	var options handlersOptions
	for _, opt := range opts {
		opt(&options)
	}

	return &CustomRouter{
		paramsHandler: options.params,
		policyHandler: options.policy,
		execHandler:   exec,
	}
}
