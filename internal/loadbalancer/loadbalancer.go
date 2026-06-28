package loadbalancer

import "sync/atomic"

// LoadBalancer selects an upstream from a list of healthy upstreams
type LoadBalancer interface {
	Next(upstreams []*Upstream) *Upstream
}

// Upstream represents a backend service instance
type Upstream struct {
	URL               string
	Healthy           atomic.Bool
	ActiveConnections atomic.Int64 // Used by LeastConnections strategy
}

// NewUpstream creates a new upstream with the given URL and initial health status.
func NewUpstream(url string, healthy bool) *Upstream {
	u := &Upstream{URL: url}
	u.Healthy.Store(healthy)
	return u
}

// New creates a load balancer based on strategy name
func New(strategy string) LoadBalancer {
	switch strategy {
	case "least-connections":
		return NewLeastConnections()
	case "round-robin":
		fallthrough
	default:
		return NewRoundRobin()
	}
}
