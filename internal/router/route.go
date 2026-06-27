package router

// route.go - Route definition structures
//
// Responsibilities:
// - Define Route struct representing a single routing rule
// - Store route matching pattern, upstream group reference, rate limit config
// - Store load balancer strategy selection
// - Provide route comparison for longest-prefix matching
//
// Key Types:
// - Route: Single routing rule with pattern, upstream group, rate limit, LB strategy
// - UpstreamGroup: Collection of Upstream instances with health check and circuit breaker config
// - Upstream: Single backend URL with health and circuit breaker state
//
// Inputs: Configuration data from config package
// Outputs: Structured route and upstream data for router and proxy

type Route struct {
	// TODO: Define route structure
}
