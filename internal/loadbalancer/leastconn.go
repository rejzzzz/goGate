package loadbalancer

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
// - Acquire(upstream *Upstream): Atomically increment connection count
// - Release(upstream *Upstream): Atomically decrement connection count
//
// Implementation Details:
// - Each Upstream has activeConnections int64 field
// - Use atomic.AddInt64 for increment/decrement (lock-free)
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
	// TODO: Implement least-connections initialization
	return &LeastConnections{}
}

// Next selects the upstream with the fewest active connections
func (lc *LeastConnections) Next(upstreams []*Upstream) *Upstream {
	// TODO: Implement least-connections selection
	return nil
}
