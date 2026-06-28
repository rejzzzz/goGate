package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rejzzzz/goGate/internal/config"
)

func TestChecker(t *testing.T) {
	// Create a test server that toggles between healthy and unhealthy
	var isHealthy atomic.Bool
	isHealthy.Store(true)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("expected path /health, got %s", r.URL.Path)
		}
		if isHealthy.Load() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()

	registry := NewRegistry()
	checker := NewChecker(registry)

	group := config.UpstreamGroup{
		Name: "test-group",
		Upstreams: []config.Upstream{
			{URL: server.URL},
		},
		HealthCheck: config.HealthCheckConfig{
			Path:     "/health",
			Interval: 10 * time.Millisecond,
			Timeout:  100 * time.Millisecond,
		},
	}

	// It should start healthy
	if !registry.IsHealthy(server.URL) {
		t.Error("expected upstream to be healthy initially")
	}

	checker.Start([]config.UpstreamGroup{group})
	defer checker.Stop()

	// Wait for the first check to run
	time.Sleep(50 * time.Millisecond)

	if !registry.IsHealthy(server.URL) {
		t.Error("expected upstream to be marked healthy by checker")
	}

	// Toggle to unhealthy
	isHealthy.Store(false)
	time.Sleep(50 * time.Millisecond)

	if registry.IsHealthy(server.URL) {
		t.Error("expected upstream to be marked unhealthy by checker")
	}

	// Toggle back to healthy
	isHealthy.Store(true)
	time.Sleep(50 * time.Millisecond)

	if !registry.IsHealthy(server.URL) {
		t.Error("expected upstream to be marked healthy again by checker")
	}
}
