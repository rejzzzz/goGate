package loadbalancer

import "sync/atomic"

// leastconn.go - Least-connections load balancing strategy
//
// Responsibilities:
// - Select upstream with fewest active connections
// - Track active connection count per upstream (atomic int64)
// - Increment connection count when request starts
// - Decrement connection count when request completes (defer in caller)
// - Break ties using round-robin when multiple upstreams have same connection count
// - Return nil if no healthy upstreams available
//
// Key Functions:
// - NewLeastConnections() *LeastConnections: Create new least-connections LB
// - Next(upstreams []*Upstream) *Upstream: Select upstream with lowest connection count
//
// Implementation Details:
// - Each Upstream has ActiveConnections atomic.Int64 field
// - Scan all healthy upstreams to find minimum connection count
// - If tie, use round-robin counter to break tie
//
// Inputs: List of upstreams with current connection counts
// Outputs: Upstream with lowest connection count (or nil if all unhealthy)

type LeastConnections struct {
	tieBreaker uint64 // Round-robin counter for tie-breaking
}

// NewLeastConnections creates a new least-connections load balancer
func NewLeastConnections() *LeastConnections {
	return &LeastConnections{}
}

// Next selects the upstream with the fewest active connections
func (lc *LeastConnections) Next(upstreams []*Upstream) *Upstream {
	var healthy []*Upstream
	for _, u := range upstreams {
		if u.Healthy.Load() {
			healthy = append(healthy, u)
		}
	}

	if len(healthy) == 0 {
		return nil
	}

	var tied []*Upstream
	minConns := int64(-1)

	for _, u := range healthy {
		conns := u.ActiveConnections.Load()
		if minConns == -1 || conns < minConns {
			minConns = conns
			tied = []*Upstream{u}
		} else if conns == minConns {
			tied = append(tied, u)
		}
	}

	if len(tied) == 1 {
		return tied[0]
	}

	// Break ties using round-robin
	idx := atomic.AddUint64(&lc.tieBreaker, 1)
	return tied[(idx-1)%uint64(len(tied))]
}
