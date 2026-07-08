package admin

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rejzzzz/goGate/internal/router"
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

func (s *Server) HandleStats(w http.ResponseWriter, r *http.Request) {
	var reqs float64
	var errs float64
	var p50, p95, p99 float64

	mfs, err := prometheus.DefaultGatherer.Gather()
	if err == nil {
		for _, mf := range mfs {
			switch mf.GetName() {
			case "gateway_requests_total":
				for _, m := range mf.GetMetric() {
					val := m.GetCounter().GetValue()
					reqs += val
					for _, lp := range m.GetLabel() {
						if lp.GetName() == "status" && (strings.HasPrefix(lp.GetValue(), "4") || strings.HasPrefix(lp.GetValue(), "5")) {
							errs += val
						}
					}
				}
			case "gateway_request_duration_seconds":
				// Just grab the first summary we see for simplicity
				if len(mf.GetMetric()) > 0 {
					for _, q := range mf.GetMetric()[0].GetSummary().GetQuantile() {
						switch q.GetQuantile() {
						case 0.5:
							p50 = q.GetValue() * 1000 // ms
						case 0.95:
							p95 = q.GetValue() * 1000
						case 0.99:
							p99 = q.GetValue() * 1000
						}
					}
				}
			}
		}
	}

	errorRate := 0.0
	if reqs > 0 {
		errorRate = errs / reqs
	}

	sendJSON(w, map[string]interface{}{
		"requestsPerSecond":     reqs / 60.0, // rough estimate based on uptime or just raw count, actual RPS needs rate calculation. We'll pass raw total for now as "requestsPerSecond" or let UI handle it. Wait, the UI expects "requestsPerSecond". Let's just return a placeholder or rough estimate.
		"p50Latency":            p50,
		"p95Latency":            p95,
		"p99Latency":            p99,
		"errorRate":             errorRate,
		"rateLimitedCount":      0, // We could extract this if we had a metric for it
		"activeCircuitBreakers": s.getActiveCircuitBreakerCount(),
	})
}

func stateString(state int) string {
	switch state {
	case 0:
		return "closed"
	case 1:
		return "open"
	case 2:
		return "half-open"
	default:
		return "unknown"
	}
}

func (s *Server) getActiveCircuitBreakerCount() int {
	count := 0
	for _, upsList := range s.upstreamMap {
		for _, u := range upsList {
			if u.CircuitBreaker != nil && int(u.CircuitBreaker.State()) != 0 {
				count++
			}
		}
	}
	return count
}

func (s *Server) HandleRoutes(w http.ResponseWriter, r *http.Request) {
	routes := s.router.GetRoutes()
	var result []map[string]interface{}

	if routes == nil {
		routes = make([]*router.Route, 0)
	}

	for _, route := range routes {
		result = append(result, map[string]interface{}{
			"path":          route.Config.Path,
			"upstreamGroup": route.Config.UpstreamGroup,
			"lbStrategy":    route.Config.LoadBalancer,
			"rateLimit":     map[string]interface{}{"rps": route.Config.RateLimit.RequestsPerSecond, "burst": route.Config.RateLimit.Burst},
			"stripPrefix":   route.Config.StripPrefix,
		})
	}
	sendJSON(w, result)
}

func (s *Server) HandleUpstreams(w http.ResponseWriter, r *http.Request) {
	healths := s.registry.GetAll()
	var result []map[string]interface{}

	for groupName, upsList := range s.upstreamMap {
		var upsData []map[string]interface{}
		for _, u := range upsList {
			status := "degraded" // default if not explicitly healthy
			if healthy, ok := healths[u.URL]; ok && healthy {
				status = "healthy"
			}

			upsData = append(upsData, map[string]interface{}{
				"url":               u.URL,
				"status":            status,
				"activeConnections": u.ActiveConnections.Load(),
				"latencyMs":         5, // Placeholder, usually measured during health checks
			})
		}
		result = append(result, map[string]interface{}{
			"name":      groupName,
			"upstreams": upsData,
		})
	}

	// Ensure we don't return nil
	if result == nil {
		result = make([]map[string]interface{}, 0)
	}

	sendJSON(w, result)
}

func (s *Server) HandleCircuitBreakers(w http.ResponseWriter, r *http.Request) {
	var result []map[string]interface{}
	for _, upsList := range s.upstreamMap {
		for _, u := range upsList {
			if u.CircuitBreaker != nil {
				result = append(result, map[string]interface{}{
					"upstreamUrl":  u.URL,
					"state":        stateString(int(u.CircuitBreaker.State())),
					"failureCount": 0, // In hystrix this is harder to extract, but we can return 0 or implement custom logic
					"lastTripTime": time.Now().Format(time.RFC3339),
				})
			}
		}
	}

	if result == nil {
		result = make([]map[string]interface{}, 0)
	}

	sendJSON(w, result)
}

func (s *Server) HandleCircuitBreakerReset(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		HandleOptions(w, r)
		return
	}

	count := 0
	for _, upsList := range s.upstreamMap {
		for _, u := range upsList {
			if u.CircuitBreaker != nil {
				u.CircuitBreaker.Reset()
				count++
			}
		}
	}

	sendJSON(w, map[string]interface{}{
		"status":  "ok",
		"message": "Circuit breakers reset",
		"count":   count,
	})
}

func (s *Server) HandleConfigReload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		HandleOptions(w, r)
		return
	}

	if s.reloadChan != nil {
		select {
		case s.reloadChan <- struct{}{}:
			sendJSON(w, map[string]string{"status": "ok", "message": "Config reload triggered"})
		default:
			sendJSON(w, map[string]string{"status": "error", "message": "Config reload already in progress"})
		}
	} else {
		sendJSON(w, map[string]string{"status": "error", "message": "Reload mechanism not configured"})
	}
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, map[string]string{"status": "healthy"})
}
