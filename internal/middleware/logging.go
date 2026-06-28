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
	"strings"
	"time"

	"go.uber.org/zap"
)

func getClientIP(r *http.Request) string {
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(clientIP, ",")
		return strings.TrimSpace(ips[0])
	}
	return r.RemoteAddr
}

// Logging returns a middleware that logs requests and responses
func Logging(logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			requestID, _ := r.Context().Value(RequestIDKey).(string)
			clientIP := getClientIP(r)

			logger.Info("request started",
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("client_ip", clientIP),
				zap.String("user_agent", r.UserAgent()),
			)
			
			// Wrap response writer to capture status code and bytes
			rw := newResponseWriter(w)
			
			next.ServeHTTP(rw, r)
			
			duration := time.Since(start)
			logger.Info("request completed",
				zap.String("request_id", requestID),
				zap.Int("status_code", rw.statusCode),
				zap.Duration("duration", duration),
				zap.Int("bytes_sent", rw.bytesWritten),
			)
		})
	}
}
