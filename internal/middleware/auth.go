package middleware

import (
	"net/http"
	"strings"

	"github.com/rejzzzz/goGate/internal/config"
	"github.com/rejzzzz/goGate/internal/router"
)

// Auth creates a middleware that enforces API Key validation
func Auth(cfg config.AuthConfig) func(http.Handler) http.Handler {
	// Pre-load keys into a map for O(1) lookup
	validKeys := make(map[string]bool)
	for _, key := range cfg.APIKeys {
		validKeys[key] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract route from context (populated by RouteMatch)
			route, ok := r.Context().Value(router.RouteContextKey).(*router.Route)
			if !ok || route == nil {
				// No route matched, let the next handler deal with 404
				next.ServeHTTP(w, r)
				return
			}

			// Determine if auth is required for this specific route
			authRequired := cfg.Enabled // Default to global setting
			if route.Config.AuthRequired != nil {
				authRequired = *route.Config.AuthRequired
			}

			if !authRequired {
				next.ServeHTTP(w, r)
				return
			}

			// Extract API Key from headers
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// Fallback to Authorization: Bearer <key>
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			// Validate key
			if apiKey == "" || !validKeys[apiKey] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Unauthorized: Invalid or missing API Key"}`))
				return
			}

			// Passed auth
			next.ServeHTTP(w, r)
		})
	}
}
