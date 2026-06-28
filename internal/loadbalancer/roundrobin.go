package loadbalancer

import "sync/atomic"

// RoundRobin implements round-robin load balancing strategy
type RoundRobin struct {
	counter uint64 // Atomic counter for round-robin selection
}

// NewRoundRobin creates a new round-robin load balancer
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

// Next selects the next upstream in round-robin order
func (rr *RoundRobin) Next(upstreams []*Upstream) *Upstream {
	// Filter for healthy upstreams
	var healthy []*Upstream
	for _, u := range upstreams {
		if u.Healthy {
			healthy = append(healthy, u)
		}
	}

	if len(healthy) == 0 {
		return nil
	}

	// Atomically increment counter
	idx := atomic.AddUint64(&rr.counter, 1)

	// Modulo against healthy slice (subtract 1 so first call hits index 0)
	return healthy[(idx-1)%uint64(len(healthy))]
}
