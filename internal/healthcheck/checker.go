package healthcheck

// checker.go - Background health check polling
//
// Responsibilities:
// - Poll each upstream's /health endpoint at configured interval
// - Run health checks in background goroutines (one per upstream group)
// - Mark upstream healthy if receives HTTP 200 within timeout
// - Mark upstream unhealthy if request fails or times out
// - Update health registry with results
// - Continue polling unhealthy upstreams to detect recovery
//
// Key Functions:
// - NewChecker(registry *Registry) *Checker: Create health checker
// - Start(upstreamGroups []*UpstreamGroup): Start background goroutines for all upstream groups
// - Stop(): Stop all health check goroutines gracefully
// - checkUpstream(upstream *Upstream, healthPath string, timeout time.Duration): Perform single health check
//
// Implementation Details:
// - Use time.Ticker for interval-based polling
// - Use http.Client with configured timeout
// - Expect HTTP 200 status code on health endpoint
// - Any non-200 or network error marks upstream unhealthy
//
// Inputs:
// - UpstreamGroup configurations with health_check settings
// - Health registry to update
//
// Outputs:
// - Updates to health registry (healthy/unhealthy status per upstream)

type Checker struct {
	registry *Registry
	stopChan chan struct{}
}

// NewChecker creates a new health checker
func NewChecker(registry *Registry) *Checker {
	// TODO: Implement checker initialization
	return &Checker{
		registry: registry,
		stopChan: make(chan struct{}),
	}
}

// Start begins health checking for all upstream groups
func (c *Checker) Start(upstreamGroups interface{}) {
	// TODO: Implement background health checking
}
