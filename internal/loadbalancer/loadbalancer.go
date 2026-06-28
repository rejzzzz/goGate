package loadbalancer

// LoadBalancer selects an upstream from a list of healthy upstreams
type LoadBalancer interface {
	Next(upstreams []*Upstream) *Upstream
}

// Upstream represents a backend service instance
type Upstream struct {
	URL               string
	Healthy           bool
	ActiveConnections int64 // Used by LeastConnections strategy
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
