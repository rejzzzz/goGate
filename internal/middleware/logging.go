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
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ParseTrustedProxies parses a list of CIDR strings into IPNet structs.
// Invalid CIDRs are logged but ignored to avoid crashing the server.
func ParseTrustedProxies(cidrs []string, logger *zap.Logger) []*net.IPNet {
	var parsed []*net.IPNet
	for _, cidr := range cidrs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			logger.Warn("invalid trusted proxy CIDR", zap.String("cidr", cidr), zap.Error(err))
			continue
		}
		parsed = append(parsed, ipnet)
	}
	return parsed
}

// isTrustedProxy checks if the given IP string matches any of the trusted proxy CIDRs.
func isTrustedProxy(ipStr string, trustedProxies []*net.IPNet) bool {
	if len(trustedProxies) == 0 {
		return false
	}
	
	// Strip port if present
	host := ipStr
	if idx := strings.LastIndex(ipStr, ":"); idx != -1 {
		if !strings.Contains(ipStr, "]") {
			host = ipStr[:idx]
		}
	}
	
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	
	for _, network := range trustedProxies {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// getClientIP returns the real client IP.
// It only trusts X-Forwarded-For if the request comes from a trusted proxy.
func getClientIP(r *http.Request, trustedProxies []*net.IPNet) string {
	remoteIP := r.RemoteAddr
	
	// Strip port for consistency
	if idx := strings.LastIndex(remoteIP, ":"); idx != -1 && !strings.Contains(remoteIP, "]") {
		remoteIP = remoteIP[:idx]
	}

	if isTrustedProxy(remoteIP, trustedProxies) {
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP != "" {
			// X-Forwarded-For can contain multiple IPs, take the first one
			ips := strings.Split(clientIP, ",")
			return strings.TrimSpace(ips[0])
		}
	}
	
	return remoteIP
}

// Logging returns a middleware that logs requests and responses
func Logging(logger *zap.Logger, trustedProxies []*net.IPNet) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID, _ := r.Context().Value(RequestIDKey).(string)
			clientIP := getClientIP(r, trustedProxies)

			// Wrap response writer to capture status code and bytes
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			logger.Info("request completed",
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("client_ip", clientIP),
				zap.Int("status_code", rw.statusCode),
				zap.Duration("duration", duration),
				zap.Int("bytes_sent", rw.bytesWritten),
			)
		})
	}
}
