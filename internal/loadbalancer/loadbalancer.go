package loadbalancer

// loadbalancer.go - Load balancer interface and factory
//
// Responsibilities:
// - Define LoadBalancer interface for upstream selection
// - Provide factory function to create LB instances by strategy name
// - Ensure all implementations handle nil return when no healthy upstreams
//
// Key Interface:
// - LoadBalancer interface:
//   - Next(upstreams []*Upstream) *Upstream: Select next upstream (returns nil if all unhealthy)
//
// Key Functions:
// - New(strategy string) LoadBalancer: Factory to create LB by strategy ("round-robin" or "least-connections")
//
// Inputs:
// - Strategy name from route configuration
// - List of upstreams (filtered for healthy only by caller)
//
// Outputs:
// - Selected upstream URL (or nil if all unhealthy)

// LoadBalancer selects an upstream from a list of healthy upstreams
type LoadBalancer interface {
	Next(upstreams []*Upstream) *Upstream
}

type Upstream struct {
	// TODO: Define upstream structure
}

// New creates a load balancer based on strategy name
func New(strategy string) LoadBalancer {
	// TODO: Implement factory function
	return nil
}
