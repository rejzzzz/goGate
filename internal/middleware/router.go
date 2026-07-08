package middleware

import (
	"context"
	"net/http"

	"github.com/rejzzzz/goGate/internal/router"
)

// RouteMatch returns a middleware that matches the request path to a configured route
// and adds the Route to the request context. If no route matches, it returns a 404.
func RouteMatch(r *router.Router) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			route, found := r.Match(req.URL.Path)
			if found {
				// Add route to context for subsequent middlewares
				ctx := context.WithValue(req.Context(), router.RouteContextKey, route)
				next.ServeHTTP(w, req.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}
