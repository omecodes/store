package objects

type routesOptions struct {
	skipPolicies   bool
	skipParams     bool
	skipEncryption bool
}

type RouteOption func(*routesOptions)
