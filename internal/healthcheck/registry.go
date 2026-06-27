package healthcheck

// registry.go - Thread-safe health status storage
//
// Responsibilities:
// - Store current health status for each upstream URL
// - Provide thread-safe read/write access to health state
// - Support querying health status by upstream URL
// - Support updating health status from health checker goroutines
//
// Key Functions:
// - NewRegistry() *Registry: Create new empty health registry
// - SetHealthy(upstreamURL string): Mark upstream as healthy
// - SetUnhealthy(upstreamURL string): Mark upstream as unhealthy
// - IsHealthy(upstreamURL string) bool: Query current health status
// - GetAll() map[string]bool: Get all upstream health states (for admin API)
//
// Implementation Details:
// - Use sync.RWMutex to protect health map
// - Health state: true = healthy, false = unhealthy
// - Default to healthy if upstream not yet checked
//
// Inputs: Health check results from checker goroutines
// Outputs: Current health status for load balancer and admin API

import "sync"

type Registry struct {
	mu     sync.RWMutex
	health map[string]bool // upstreamURL -> healthy
}

// NewRegistry creates a new health status registry
func NewRegistry() *Registry {
	// TODO: Implement registry initialization
	return &Registry{
		health: make(map[string]bool),
	}
}

// IsHealthy returns the current health status of an upstream
func (r *Registry) IsHealthy(upstreamURL string) bool {
	// TODO: Implement thread-safe health status read
	return true
}
