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
	"net/http"
	"strconv"
	"strings"

	"github.com/rejzzzz/goGate/internal/config"
	"github.com/rejzzzz/goGate/internal/metrics"
	"github.com/rejzzzz/goGate/internal/ratelimit"
	"github.com/rejzzzz/goGate/internal/router"
)

// RateLimit returns a middleware that enforces rate limits
func RateLimit(store *ratelimit.RedisStore, cfg config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Bypass Check
			if cfg.RateLimitBypassHeader != "" && cfg.RateLimitBypassToken != "" {
				if r.Header.Get(cfg.RateLimitBypassHeader) == cfg.RateLimitBypassToken {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 2. Global Rate Limit Check
			if cfg.GlobalRateLimit.RequestsPerSecond > 0 {
				globalAllowed, globalRemaining, err := store.CheckGlobalRateLimit(
					cfg.GlobalRateLimit.RequestsPerSecond,
					cfg.GlobalRateLimit.Burst,
				)
				if err != nil {
					fmt.Printf("Global Rate limit error (failing open): %v\n", err)
				} else {
					w.Header().Set("X-Global-RateLimit-Limit", strconv.FormatFloat(cfg.GlobalRateLimit.RequestsPerSecond, 'f', 2, 64))
					w.Header().Set("X-Global-RateLimit-Remaining", strconv.Itoa(globalRemaining))

					if !globalAllowed {
						w.Header().Set("Retry-After", "1")
						metrics.RecordRateLimit("global")
						http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
						return
					}
				}
			}

			// 3. Per-IP Route specific Rate Limit Check
			rt, ok := r.Context().Value(router.RouteContextKey).(*router.Route)
			if !ok || rt == nil || rt.Config.RateLimit.RequestsPerSecond <= 0 {
				// No rate limit configured for this route, skip
				next.ServeHTTP(w, r)
				return
			}

			clientIP := getClientIP(r) // defined in logging.go

			// If getClientIP contains port (e.g., 127.0.0.1:54321), strip it
			if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
				// Simple check to not break IPv6 (which has multiple colons) unless it's bracketed
				if !strings.Contains(clientIP, "]") {
					clientIP = clientIP[:idx]
				}
			}

			allowed, remaining, err := store.CheckRateLimit(
				rt.Config.Path,
				clientIP,
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
