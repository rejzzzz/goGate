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
	"strconv"
	"time"

	"github.com/rejzzzz/goGate/internal/metrics"
	"github.com/rejzzzz/goGate/internal/router"
)

// Metrics returns a middleware that records Prometheus metrics
func Metrics() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()

			// Extract route from context (set by RouteMatch middleware)
			routePath := "unknown"
			if rt, ok := r.Context().Value(router.RouteContextKey).(*router.Route); ok {
				routePath = rt.Config.Path
			}

			metrics.RecordRequest(
				routePath,
				r.Method,
				strconv.Itoa(rw.statusCode),
				duration,
			)
		})
	}
}
