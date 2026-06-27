package middleware

// logging.go - Structured request/response logging
//
// Responsibilities:
// - Log request start with method, path, client IP, user agent
// - Log request completion with status code, duration, bytes sent
// - Include request ID in all log entries
// - Use structured JSON format with zap logger
// - Avoid logging in hot path at high RPS (use async logger or sampling)
//
// Key Functions:
// - Logging(logger *zap.Logger) Middleware: Create logging middleware
//
// Log Fields:
// - Request: method, path, request_id, client_ip, user_agent
// - Response: status_code, duration_ms, bytes_sent, upstream_url
//
// Inputs:
// - HTTP request and response
// - Request ID from context
//
// Outputs:
// - Structured JSON log entries

import (
	"net/http"
	"time"
	"go.uber.org/zap"
)

// Logging returns a middleware that logs requests and responses
func Logging(logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// TODO: Log request start
			logger.Info("request started",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)
			
			// Wrap response writer to capture status code
			// TODO: Implement response writer wrapper
			
			next.ServeHTTP(w, r)
			
			duration := time.Since(start)
			// TODO: Log request completion with status and duration
			logger.Info("request completed",
				zap.Duration("duration", duration),
			)
		})
	}
}
