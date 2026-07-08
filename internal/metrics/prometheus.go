package metrics

// prometheus.go - Prometheus metric definitions and registration
//
// Responsibilities:
// - Define all gateway metrics (counters, histograms, gauges)
// - Register metrics with Prometheus client
// - Provide helper functions to record metric values
//
// Metrics:
// - gateway_requests_total: Counter with labels route, method, status_code
// - gateway_request_duration_seconds: Histogram with labels route, method
// - gateway_upstream_request_duration_seconds: Histogram with labels upstream, route
// - gateway_rate_limited_total: Counter with label route
// - gateway_circuit_breaker_state: Gauge with label upstream (0=Closed, 1=Open, 2=Half-Open)
// - gateway_circuit_breaker_trips_total: Counter with label upstream
// - gateway_upstream_health: Gauge with label upstream (1=healthy, 0=unhealthy)
// - gateway_active_connections: Gauge with label upstream
// - gateway_redis_operation_duration_seconds: Histogram with label operation
//
// Key Functions:
// - Init(): Initialize and register all metrics
// - RecordRequest(route, method, statusCode string, duration float64): Record request metrics
// - RecordRateLimit(route string): Increment rate limit counter
// - SetCircuitBreakerState(upstream string, state int): Update circuit breaker state gauge
// - SetUpstreamHealth(upstream string, healthy bool): Update upstream health gauge
//
// Inputs: Metric values from middleware and gateway components
// Outputs: Registered Prometheus metrics available at /metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total requests handled",
		},
		[]string{"route", "method", "status_code"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request latency buckets",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"route", "method"},
	)

	upstreamRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_upstream_request_duration_seconds",
			Help:    "Upstream call latency",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"upstream", "route"},
	)

	rateLimitedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_rate_limited_total",
			Help: "Requests rejected by rate limiter",
		},
		[]string{"route"},
	)

	circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_circuit_breaker_state",
			Help: "0=closed, 1=open, 2=half-open",
		},
		[]string{"upstream"},
	)

	circuitBreakerTripsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_circuit_breaker_trips_total",
			Help: "Total times circuit opened",
		},
		[]string{"upstream"},
	)

	upstreamHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_upstream_health",
			Help: "1=healthy, 0=unhealthy",
		},
		[]string{"upstream"},
	)

	activeConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_active_connections",
			Help: "Current active connections per upstream",
		},
		[]string{"upstream"},
	)

	redisOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_redis_operation_duration_seconds",
			Help:    "Redis latency for rate limiting ops",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"operation"},
	)
)

// Init initializes and registers all Prometheus metrics
func Init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(upstreamRequestDuration)
	prometheus.MustRegister(rateLimitedTotal)
	prometheus.MustRegister(circuitBreakerState)
	prometheus.MustRegister(circuitBreakerTripsTotal)
	prometheus.MustRegister(upstreamHealth)
	prometheus.MustRegister(activeConnections)
	prometheus.MustRegister(redisOperationDuration)
}

// RecordRequest records a completed request
func RecordRequest(route, method, statusCode string, duration float64) {
	requestsTotal.WithLabelValues(route, method, statusCode).Inc()
	requestDuration.WithLabelValues(route, method).Observe(duration)
}

// RecordRateLimit records a rate limited request
func RecordRateLimit(route string) {
	rateLimitedTotal.WithLabelValues(route).Inc()
}

// SetCircuitBreakerState sets the state of the circuit breaker for an upstream
func SetCircuitBreakerState(upstream string, state int) {
	circuitBreakerState.WithLabelValues(upstream).Set(float64(state))
}

// RecordCircuitBreakerTrip records a circuit breaker trip
func RecordCircuitBreakerTrip(upstream string) {
	circuitBreakerTripsTotal.WithLabelValues(upstream).Inc()
}

// SetUpstreamHealth sets the health of an upstream
func SetUpstreamHealth(upstream string, healthy bool) {
	val := 0.0
	if healthy {
		val = 1.0
	}
	upstreamHealth.WithLabelValues(upstream).Set(val)
}

// SetActiveConnections sets the active connection count for an upstream
func SetActiveConnections(upstream string, connections int64) {
	activeConnections.WithLabelValues(upstream).Set(float64(connections))
}

// RecordRedisOperation records the duration of a Redis operation
func RecordRedisOperation(operation string, duration float64) {
	redisOperationDuration.WithLabelValues(operation).Observe(duration)
}
