package middleware

// recovery.go - Panic recovery middleware
//
// Responsibilities:
// - Catch panics from downstream handlers
// - Log panic with stack trace
// - Return HTTP 500 Internal Server Error to client
// - Prevent entire gateway process from crashing
// - Allow subsequent requests to be processed normally
//
// Key Functions:
// - Recovery(logger *zap.Logger) Middleware: Create recovery middleware with logger
//
// Inputs:
// - Zap logger for structured logging
// - Panics from downstream handlers
//
// Outputs:
// - HTTP 500 response on panic
// - Log entry with panic message and stack trace

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// Recovery returns a middleware that recovers from panics
func Recovery(logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID, _ := r.Context().Value(RequestIDKey).(string)

					logger.Error("panic recovered",
						zap.String("request_id", requestID),
						zap.Any("error", err),
						zap.ByteString("stacktrace", debug.Stack()),
					)

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
