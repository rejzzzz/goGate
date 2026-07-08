package metrics

// exporter.go - Prometheus metrics HTTP handler
//
// Responsibilities:
// - Provide HTTP handler for /metrics endpoint
// - Export metrics in Prometheus text exposition format
// - Use promhttp.Handler() from Prometheus client library
//
// Key Functions:
// - Handler() http.Handler: Return HTTP handler for /metrics endpoint
//
// Inputs: HTTP GET request to /metrics
// Outputs: Prometheus metrics in text format

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Handler returns the HTTP handler for the /metrics endpoint
func Handler() http.Handler {
	return promhttp.Handler()
}
