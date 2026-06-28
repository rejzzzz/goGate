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

import (
	"log"
	"net/http"
	"time"

	"github.com/rejzzzz/goGate/internal/config"
)

type Checker struct {
	registry *Registry
	stopChan chan struct{}
}

// NewChecker creates a new health checker
func NewChecker(registry *Registry) *Checker {
	return &Checker{
		registry: registry,
		stopChan: make(chan struct{}),
	}
}

// Start begins health checking for all upstream groups
func (c *Checker) Start(upstreamGroups []config.UpstreamGroup) {
	for _, group := range upstreamGroups {
		if group.HealthCheck.Interval <= 0 {
			continue // Skip if health check interval is not configured
		}

		go c.checkGroup(group)
	}
}

// Stop gracefully stops all background health checking
func (c *Checker) Stop() {
	close(c.stopChan)
}

func (c *Checker) checkGroup(group config.UpstreamGroup) {
	ticker := time.NewTicker(group.HealthCheck.Interval)
	defer ticker.Stop()

	// Perform an initial check immediately
	c.runCheck(group)

	for {
		select {
		case <-ticker.C:
			c.runCheck(group)
		case <-c.stopChan:
			return
		}
	}
}

func (c *Checker) runCheck(group config.UpstreamGroup) {
	timeout := group.HealthCheck.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	for _, u := range group.Upstreams {
		url := u.URL + group.HealthCheck.Path
		
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			c.registry.SetHealthy(u.URL)
			if resp.Body != nil {
				resp.Body.Close()
			}
		} else {
			if err != nil {
				log.Printf("[HealthCheck] %s is unhealthy: %v", u.URL, err)
			} else {
				log.Printf("[HealthCheck] %s is unhealthy: HTTP %d", u.URL, resp.StatusCode)
				if resp.Body != nil {
					resp.Body.Close()
				}
			}
			c.registry.SetUnhealthy(u.URL)
		}
	}
}
