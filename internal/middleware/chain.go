package middleware

// chain.go - Middleware composition and chaining
//
// Responsibilities:
// - Provide standard middleware pattern for http.Handler
// - Chain multiple middleware in specified order
// - Ensure middleware executes in correct sequence
//
// Middleware Order (outermost to innermost):
// 1. Recovery - Catch panics and return HTTP 500
// 2. RequestID - Generate/propagate X-Request-ID
// 3. Logging - Log request start and completion
// 4. Metrics - Record request counts and latencies
// 5. RateLimit - Enforce rate limits, reject with 429
// 6. Proxy - Load balance, circuit break, forward to upstream
//
// Key Types:
// - Middleware: func(http.Handler) http.Handler
//
// Key Functions:
// - Chain(handler http.Handler, middlewares ...Middleware) http.Handler: Compose middleware chain
//
// Inputs:
// - Final handler (proxy)
// - Ordered list of middleware functions
//
// Outputs:
// - Wrapped handler with all middleware applied

import "net/http"

// Middleware wraps an http.Handler with additional functionality
type Middleware func(http.Handler) http.Handler

// Chain applies middleware to a handler in the specified order
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Apply middleware in reverse order so outermost is first
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
