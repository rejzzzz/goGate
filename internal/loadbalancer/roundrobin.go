package loadbalancer

// roundrobin.go - Round-robin load balancing strategy
//
// Responsibilities:
// - Distribute requests evenly across all healthy upstreams
// - Maintain atomic counter to track next upstream index
// - Filter for healthy upstreams only before selecting
// - Return nil if no healthy upstreams available
// - Use modulo arithmetic to wrap around to start of list
//
// Key Functions:
// - NewRoundRobin() *RoundRobin: Create new round-robin LB with counter at 0
// - Next(upstreams []*Upstream) *Upstream: Select next upstream using atomic increment
//
// Implementation Details:
// - Use atomic.AddUint64 to increment counter (lock-free)
// - Filter upstreams to healthy-only list first
// - Apply modulo on counter with len(healthyUpstreams)
// - Thread-safe: multiple goroutines can call Next() concurrently
//
// Inputs: List of all upstreams (healthy and unhealthy)
// Outputs: Next upstream in round-robin rotation (or nil if all unhealthy)

import "sync/atomic"

type RoundRobin struct {
	counter uint64 // Atomic counter for round-robin selection
}

// NewRoundRobin creates a new round-robin load balancer
func NewRoundRobin() *RoundRobin {
	// TODO: Implement round-robin initialization
	return &RoundRobin{}
}

// Next selects the next upstream in round-robin order
func (rr *RoundRobin) Next(upstreams []*Upstream) *Upstream {
	// TODO: Implement round-robin selection with atomic counter
	_ = atomic.AddUint64(&rr.counter, 1) // Placeholder for atomic increment
	return nil
}
