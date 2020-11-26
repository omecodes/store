package router

type routes struct {
	skipPolicies  bool
	skipParams    bool
	skipExecution bool
}

type RouteOption func(*routes)

func SkipParamsCheck() RouteOption {
	return func(r *routes) {
		r.skipParams = true
	}
}

func SkipPoliciesCheck() RouteOption {
	return func(r *routes) {
		r.skipPolicies = true
	}
}

func SkipExec() RouteOption {
	return func(r *routes) {
		r.skipExecution = true
	}
}

func Route(opts ...RouteOption) (handler Handler) {
	routes := routes{}

	for _, o := range opts {
		o(&routes)
	}

	if !routes.skipExecution {
		handler = &execHandler{}
	} else {
		handler = &dummyHandler{}
	}

	if !routes.skipPolicies {
		handler = &policyHandler{base: base{
			next: handler,
		}}
	}

	if !routes.skipParams {
		handler = &paramsHandler{
			base{next: handler},
		}
	}
	return
}
