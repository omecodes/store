package objects

type CustomRouter struct {
	paramsHandler *ParamsHandler
	policyHandler *ACLHandler
	execHandler   Handler
}

type objectsHandlersOptions struct {
	params *ParamsHandler
	policy *ACLHandler
}

type HandlerOption func(*objectsHandlersOptions)

func WithDefaultParamsHandler() HandlerOption {
	return func(options *objectsHandlersOptions) {
		options.params = &ParamsHandler{}
	}
}

func WithDefaultPolicyHandler() HandlerOption {
	return func(options *objectsHandlersOptions) {
		options.policy = &ACLHandler{}
	}
}

func (r *CustomRouter) GetHandler(opts ...RouteOption) Handler {
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

func NewCustomRouter(exec Handler, opts ...HandlerOption) *CustomRouter {
	var options objectsHandlersOptions
	for _, opt := range opts {
		opt(&options)
	}
	return &CustomRouter{
		paramsHandler: options.params,
		policyHandler: options.policy,
		execHandler:   exec,
	}
}
