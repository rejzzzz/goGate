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

	srv := NewServer(8081, r, uMap, reg, nil, "")

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

	srv := NewServer(8081, r, uMap, reg, reloadChan, "")

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

func TestHandleMetricsHistory(t *testing.T) {
	// 1. Create a mock Prometheus server
	mockPrometheus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/query_range" {
			w.Header().Set("Content-Type", "application/json")
			query := r.URL.Query().Get("query")
			
			if query == "sum(rate(gateway_requests_total[1m]))" {
				// Mock RPS response
				w.Write([]byte(`{
					"status": "success",
					"data": {
						"resultType": "matrix",
						"result": [
							{
								"metric": {},
								"values": [
									[1715000000, "15.5"],
									[1715000010, "20.1"]
								]
							}
						]
					}
				}`))
			} else {
				// Mock Latency response
				w.Write([]byte(`{
					"status": "success",
					"data": {
						"resultType": "matrix",
						"result": [
							{
								"metric": {},
								"values": [
									[1715000000, "0.012"],
									[1715000010, "0.015"]
								]
							}
						]
					}
				}`))
			}
		} else {
			http.NotFound(w, r)
		}
	}))
	defer mockPrometheus.Close()

	// 2. Setup Server with mock Prometheus URL
	r := router.New(nil)
	uMap := &atomic.Value{}
	uMap.Store(make(map[string][]*loadbalancer.Upstream))
	reg := healthcheck.NewRegistry()

	srv := NewServer(8081, r, uMap, reg, nil, mockPrometheus.URL)

	// 3. Perform Request
	req, err := http.NewRequest("GET", "/admin/api/metrics/history?window=5m", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.HandleMetricsHistory)
	handler.ServeHTTP(rr, req)

	// 4. Verify Response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response []HistoryData
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v, body: %s", err, rr.Body.String())
	}

	if len(response) != 2 {
		t.Fatalf("expected 2 data points, got %d", len(response))
	}

	// Verify the parsed data
	// Note: The time string depends on the timezone of the system running the test,
	// so we only check RPS and Latency which should be mapped correctly.
	
	// First point
	if response[0].RPS != 15.5 {
		t.Errorf("expected RPS 15.5, got %v", response[0].RPS)
	}
	if response[0].Latency != 12.0 { // 0.012 * 1000
		t.Errorf("expected Latency 12.0, got %v", response[0].Latency)
	}

	// Second point
	if response[1].RPS != 20.1 {
		t.Errorf("expected RPS 20.1, got %v", response[1].RPS)
	}
	if response[1].Latency != 15.0 { // 0.015 * 1000
		t.Errorf("expected Latency 15.0, got %v", response[1].Latency)
	}
}
