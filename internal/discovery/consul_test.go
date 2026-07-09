package discovery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockConsulService models the JSON response from Consul's /v1/health/service endpoint
type mockConsulService struct {
	Node struct {
		Address string
	}
	Service struct {
		Address string
		Port    int
	}
}

func TestConsulProvider_GetInstances(t *testing.T) {
	// 1. Create a mock Consul server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/health/service/my-service" {
			t.Errorf("Expected path /v1/health/service/my-service, got %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		mockResp := []mockConsulService{
			{
				Node: struct{ Address string }{Address: "10.0.0.1"},
				Service: struct {
					Address string
					Port    int
				}{Address: "10.0.0.1", Port: 8080},
			},
			{
				Node: struct{ Address string }{Address: "10.0.0.2"},
				Service: struct {
					Address string
					Port    int
				}{Address: "", Port: 9090}, // Should fallback to Node Address
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer ts.Close()

	// 2. Initialize our provider pointing to the mock server
	// Note: ts.URL starts with http://, but consul api client expects an address like 127.0.0.1:8080
	// We'll strip the http:// scheme
	addr := ts.URL[7:]

	provider, err := NewConsulProvider(addr)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// We must override the schema back to http for the mock server,
	// because by default consul api assumes http anyway but sometimes config needs it.
	// We'll just rely on defaults.

	// 3. Test GetInstances
	instances, err := provider.GetInstances("my-service")
	if err != nil {
		t.Fatalf("GetInstances failed: %v", err)
	}

	if len(instances) != 2 {
		t.Fatalf("Expected 2 instances, got %d", len(instances))
	}

	expected1 := "http://10.0.0.1:8080"
	expected2 := "http://10.0.0.2:9090"

	if instances[0] != expected1 {
		t.Errorf("Expected instance 0 to be %s, got %s", expected1, instances[0])
	}
	if instances[1] != expected2 {
		t.Errorf("Expected instance 1 to be %s, got %s", expected2, instances[1])
	}
}

func TestConsulProvider_Watch(t *testing.T) {
	var requestCount int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		w.Header().Set("Content-Type", "application/json")

		if requestCount == 1 {
			// First request: return index 10 and 1 instance
			w.Header().Set("X-Consul-Index", "10")
			mockResp := []mockConsulService{
				{
					Node: struct{ Address string }{Address: "10.0.0.1"},
					Service: struct {
						Address string
						Port    int
					}{Address: "10.0.0.1", Port: 8080},
				},
			}
			json.NewEncoder(w).Encode(mockResp)
		} else if requestCount == 2 {
			// Second request: block for a moment, then return index 11 and 2 instances
			waitIdx := r.URL.Query().Get("index")
			if waitIdx != "10" {
				t.Errorf("Expected wait index 10, got %s", waitIdx)
			}

			time.Sleep(100 * time.Millisecond)

			w.Header().Set("X-Consul-Index", "11")
			mockResp := []mockConsulService{
				{
					Node: struct{ Address string }{Address: "10.0.0.1"},
					Service: struct {
						Address string
						Port    int
					}{Address: "10.0.0.1", Port: 8080},
				},
				{
					Node: struct{ Address string }{Address: "10.0.0.2"},
					Service: struct {
						Address string
						Port    int
					}{Address: "10.0.0.2", Port: 9090},
				},
			}
			json.NewEncoder(w).Encode(mockResp)
		} else {
			// Third request onwards, just block to simulate no more changes
			time.Sleep(1 * time.Second)
			w.Header().Set("X-Consul-Index", "11")
			mockResp := []mockConsulService{} // doesn't matter, we block
			json.NewEncoder(w).Encode(mockResp)
		}
	}))
	defer ts.Close()

	addr := ts.URL[7:]
	provider, err := NewConsulProvider(addr)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	updateCh := make(chan []string)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run watch in background
	go provider.Watch(ctx, "my-service", updateCh)

	// We expect two updates on the channel

	// First update
	select {
	case instances := <-updateCh:
		if len(instances) != 1 {
			t.Fatalf("Expected 1 instance on first update, got %d", len(instances))
		}
		if instances[0] != "http://10.0.0.1:8080" {
			t.Errorf("Unexpected instance: %s", instances[0])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for first update")
	}

	// Second update (after block)
	select {
	case instances := <-updateCh:
		if len(instances) != 2 {
			t.Fatalf("Expected 2 instances on second update, got %d", len(instances))
		}
		if instances[1] != "http://10.0.0.2:9090" {
			t.Errorf("Unexpected instance: %s", instances[1])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for second update")
	}

	// Cancel the watch
	cancel()
}
