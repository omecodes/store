package router

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
