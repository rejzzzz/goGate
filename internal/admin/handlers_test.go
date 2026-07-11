package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/rejzzzz/goGate/internal/config"
	"github.com/rejzzzz/goGate/internal/healthcheck"
	"github.com/rejzzzz/goGate/internal/loadbalancer"
	"github.com/rejzzzz/goGate/internal/router"
)

func TestHandleRoutes(t *testing.T) {
	// 1. Setup Mock State
	cfgRoutes := []config.Route{
		{
			Path:          "/api/v1/test",
			UpstreamGroup: "test-group",
			LoadBalancer:  "round-robin",
			StripPrefix:   true,
			RateLimit: config.RateLimitConfig{
				RequestsPerSecond: 100,
				Burst:             20,
			},
		},
	}

	r := router.New(cfgRoutes)
	uMap := &atomic.Value{}
	uMap.Store(make(map[string][]*loadbalancer.Upstream))
	reg := healthcheck.NewRegistry()

	srv := NewServer(8081, r, uMap, reg, nil)

	// 2. Perform Request
	req, err := http.NewRequest("GET", "/admin/api/routes", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.HandleRoutes)

	handler.ServeHTTP(rr, req)

	// 3. Verify Response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response []map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(response) != 1 {
		t.Fatalf("expected 1 route, got %d", len(response))
	}

	route := response[0]
	if route["path"] != "/api/v1/test" {
		t.Errorf("expected path '/api/v1/test', got %v", route["path"])
	}

	if route["upstreamGroup"] != "test-group" {
		t.Errorf("expected upstreamGroup 'test-group', got %v", route["upstreamGroup"])
	}
}

func TestHandleConfigReload(t *testing.T) {
	r := router.New(nil)
	uMap := &atomic.Value{}
	uMap.Store(make(map[string][]*loadbalancer.Upstream))
	reg := healthcheck.NewRegistry()

	// Create channel with buffer 1 so it doesn't block
	reloadChan := make(chan struct{}, 1)

	srv := NewServer(8081, r, uMap, reg, reloadChan)

	req, err := http.NewRequest("POST", "/admin/api/config/reload", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.HandleConfigReload)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	select {
	case <-reloadChan:
		// success, channel received signal
	default:
		t.Errorf("expected reloadChan to receive a signal but it did not")
	}
}
