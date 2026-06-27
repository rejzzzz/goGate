package middleware

// requestid.go - Request ID generation and propagation
//
// Responsibilities:
// - Generate unique request ID (UUID) if not present in request
// - Propagate existing X-Request-ID header if present
// - Add request ID to request context for downstream handlers
// - Include request ID in all log entries
//
// Key Functions:
// - RequestID() Middleware: Create request ID middleware
// - generateRequestID() string: Generate UUID v4 request ID
//
// Inputs:
// - Incoming HTTP request (may have X-Request-ID header)
//
// Outputs:
// - X-Request-ID header set in request and response
// - Request ID in context for logging

import "net/http"

// RequestID returns a middleware that generates or propagates request IDs
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				// TODO: Generate new UUID
				requestID = "generated-uuid"
			}
			
			// Add to request header and response
			r.Header.Set("X-Request-ID", requestID)
			w.Header().Set("X-Request-ID", requestID)
			
			// TODO: Add to request context for logging
			next.ServeHTTP(w, r)
		})
	}
}
