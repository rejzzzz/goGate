package middleware

// ratelimit.go - Rate limiting middleware
//
// Responsibilities:
// - Extract client IP from X-Forwarded-For header or connection remote address
// - Check rate limit via Redis store
// - Reject requests with HTTP 429 if limit exceeded
// - Add X-RateLimit-* headers to all responses
// - Skip rate limiting if route has no rate limit config
//
// Key Functions:
// - RateLimit(store *RedisStore) Middleware: Create rate limiting middleware
// - extractClientIP(r *http.Request) string: Extract client IP from request
//
// Response Headers:
// - X-RateLimit-Limit: {rate} (on all responses)
// - X-RateLimit-Remaining: {remaining} (on all responses)
// - Retry-After: {seconds} (on 429 responses)
//
// Inputs:
// - HTTP request with client IP
// - Route rate limit configuration from context
//
// Outputs:
// - HTTP 429 Too Many Requests if limit exceeded
// - X-RateLimit-* headers on all responses

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/rejzzzz/goGate/internal/config"
	"github.com/rejzzzz/goGate/internal/metrics"
	"github.com/rejzzzz/goGate/internal/ratelimit"
	"github.com/rejzzzz/goGate/internal/router"
)

// RateLimit returns a middleware that enforces rate limits
func RateLimit(store *ratelimit.RedisStore, globalLimiter *ratelimit.GlobalLimiter, cfg config.Config, trustedProxies []*net.IPNet) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Bypass Check
			if cfg.RateLimitBypassHeader != "" && cfg.RateLimitBypassToken != "" {
				if r.Header.Get(cfg.RateLimitBypassHeader) == cfg.RateLimitBypassToken {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 2. Global Rate Limit Check (In-Memory)
			if globalLimiter != nil {
				w.Header().Set("X-Global-RateLimit-Limit", strconv.FormatFloat(cfg.GlobalRateLimit.RequestsPerSecond, 'f', 2, 64))
				w.Header().Set("X-Global-RateLimit-Remaining", strconv.Itoa(int(globalLimiter.Tokens())))

				if !globalLimiter.Allow() {
					w.Header().Set("Retry-After", "1")
					metrics.RecordRateLimit("global")
					http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
					return
				}
			}

			// 3. Route specific Rate Limit Check
			rt, ok := r.Context().Value(router.RouteContextKey).(*router.Route)
			if !ok || rt == nil || rt.Config.RateLimit.RequestsPerSecond <= 0 {
				// No rate limit configured for this route, skip
				next.ServeHTTP(w, r)
				return
			}

			// Determine Identity for Rate Limiting
			// Prioritize API Key if present
			identity := r.Header.Get("X-API-Key")
			if identity == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					identity = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			// Fallback to IP if no API Key is provided
			if identity == "" {
				clientIP := getClientIP(r, trustedProxies) // defined in logging.go
				// If getClientIP contains port (e.g., 127.0.0.1:54321), strip it
				if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
					// Simple check to not break IPv6 (which has multiple colons) unless it's bracketed
					if !strings.Contains(clientIP, "]") {
						clientIP = clientIP[:idx]
					}
				}
				identity = clientIP
			}

			allowed, remaining, err := store.CheckRateLimit(
				r.Context(),
				rt.Config.Path,
				identity,
				rt.Config.RateLimit.RequestsPerSecond,
				rt.Config.RateLimit.Burst,
			)

			if err != nil {
				// Log error, but fail open to not break traffic if Redis is down
				fmt.Printf("Rate limit error (failing open): %v\n", err)
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.FormatFloat(rt.Config.RateLimit.RequestsPerSecond, 'f', 2, 64))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if !allowed {
				w.Header().Set("Retry-After", "1")
				metrics.RecordRateLimit(rt.Config.Path)
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
