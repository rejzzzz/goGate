package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rejzzzz/goGate/internal/circuitbreaker"
	"github.com/rejzzzz/goGate/internal/loadbalancer"
	"github.com/rejzzzz/goGate/internal/proxy"
)

func TestHTTPProxy_CircuitBreakerIntegration(t *testing.T) {
	// 1. Create a dummy backend server that always returns 500 Internal Server Error
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer backend.Close()

	// 2. Setup the Upstream and Circuit Breaker
	up := loadbalancer.NewUpstream(backend.URL, true)
	up.CircuitBreaker = circuitbreaker.NewBreaker(&circuitbreaker.Config{
		FailureThreshold: 2, // Trip after 2 failures
		Timeout:          1 * time.Second,
		WindowSize:       10 * time.Second,
		BucketCount:      10,
	})

	// 3. Setup the Proxy
	p := proxy.NewHTTPProxy(nil)

	// Helper to send a request through the proxy
	sendRequest := func() int {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		// Manually check breaker like main.go does
		if up.CircuitBreaker != nil && !up.CircuitBreaker.Allow() {
			w.Header().Set("X-Circuit-Breaker", "open")
			http.Error(w, "Service Unavailable: Circuit Breaker Open", http.StatusServiceUnavailable)
			return w.Code
		}

		p.ServeHTTP(w, req, up, "")
		return w.Code
	}

	// Request 1: Backend returns 500, breaker records failure (FailureCount = 1)
	if code := sendRequest(); code != http.StatusInternalServerError && code != http.StatusBadGateway {
		t.Fatalf("Expected 500 or 502 for request 1, got %d", code)
	}

	// Request 2: Backend returns 500, breaker records failure (FailureCount = 2 -> Trips to OPEN)
	if code := sendRequest(); code != http.StatusInternalServerError && code != http.StatusBadGateway {
		t.Fatalf("Expected 500 or 502 for request 2, got %d", code)
	}

	// Request 3: Circuit Breaker is now OPEN. It should short-circuit and return 503.
	if code := sendRequest(); code != http.StatusServiceUnavailable {
		t.Fatalf("Expected 503 Service Unavailable because Circuit Breaker is Open, got %d", code)
	}
}
