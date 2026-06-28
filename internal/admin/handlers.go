package admin

import (
	"encoding/json"
	"net/http"
	"time"
)

// Helper to set CORS headers and return JSON
func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(data)
}

func HandleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
}

func HandleStats(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, map[string]interface{}{
		"requestsPerSecond":     12450,
		"p50Latency":            4.2,
		"p95Latency":            11.5,
		"p99Latency":            14.8,
		"errorRate":             0.05,
		"rateLimitedCount":      120,
		"activeCircuitBreakers": 0,
	})
}

func HandleRoutes(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, []map[string]interface{}{
		{
			"path":          "/api/v1/users",
			"upstreamGroup": "user-service",
			"lbStrategy":    "round-robin",
			"rateLimit":     map[string]interface{}{"rps": 100, "burst": 20},
			"stripPrefix":   true,
		},
		{
			"path":          "/api/v1/orders",
			"upstreamGroup": "order-service",
			"lbStrategy":    "least-connections",
			"rateLimit":     map[string]interface{}{"rps": 50, "burst": 10},
			"stripPrefix":   true,
		},
		{
			"path":          "/api/v1/grpc",
			"upstreamGroup": "grpc-service",
			"lbStrategy":    "round-robin",
			"rateLimit":     map[string]interface{}{"rps": 500, "burst": 100},
			"stripPrefix":   false,
		},
	})
}

func HandleUpstreams(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, []map[string]interface{}{
		{
			"name": "user-service",
			"upstreams": []map[string]interface{}{
				{"url": "http://localhost:8081", "status": "healthy", "activeConnections": 124, "latencyMs": 3},
				{"url": "http://localhost:8082", "status": "degraded", "activeConnections": 45, "latencyMs": 56},
			},
		},
		{
			"name": "order-service",
			"upstreams": []map[string]interface{}{
				{"url": "http://localhost:8083", "status": "healthy", "activeConnections": 89, "latencyMs": 4},
				{"url": "http://localhost:8084", "status": "healthy", "activeConnections": 91, "latencyMs": 5},
			},
		},
	})
}

func HandleCircuitBreakers(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, []map[string]interface{}{
		{"upstreamUrl": "http://localhost:8081", "state": "closed", "failureCount": 0},
		{"upstreamUrl": "http://localhost:8082", "state": "half-open", "failureCount": 4, "lastTripTime": time.Now().Add(-15 * time.Second).Format(time.RFC3339)},
	})
}

func HandleCircuitBreakerReset(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		HandleOptions(w, r)
		return
	}
	sendJSON(w, map[string]string{"status": "ok", "message": "Circuit breaker reset"})
}

func HandleConfigReload(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, map[string]string{"status": "ok", "message": "Config reloaded"})
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, map[string]string{"status": "healthy"})
}
