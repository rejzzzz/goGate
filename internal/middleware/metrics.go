package middleware

// metrics.go - Prometheus metrics recording middleware
//
// Responsibilities:
// - Record request count by route, method, status code
// - Record request duration histogram by route, method
// - Record rate limit rejections
// - Wrap response writer to capture status code
//
// Key Functions:
// - Metrics() Middleware: Create metrics recording middleware
//
// Metrics Recorded:
// - gateway_requests_total (counter)
// - gateway_request_duration_seconds (histogram)
//
// Inputs:
// - HTTP request and response
// - Route from context (set by router)
//
// Outputs:
// - Prometheus metric updates

import (
	"net/http"
	"time"
)

// Metrics returns a middleware that records Prometheus metrics
func Metrics() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap response writer to capture status code
			// TODO: Implement response writer wrapper
			
			next.ServeHTTP(w, r)
			
			duration := time.Since(start).Seconds()
			// TODO: Record metrics using prometheus package
			_ = duration
		})
	}
}
