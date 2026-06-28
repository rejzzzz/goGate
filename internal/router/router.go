package router

import (
	"github.com/yourusername/api-gateway/internal/config"
	"strings"
	"sync/atomic"
)

// Router handles HTTP request routing
type Router struct {
	// We use atomic.Value to allow lock-free, concurrent reads while 
	// supporting hot-reloads of the route table.
	routes atomic.Value 
}

// New creates a new router initialized with the given config routes
func New(configRoutes []config.Route) *Router {
	r := &Router{}
	r.Reload(configRoutes)
	return r
}

// Reload atomically updates the route table
func (r *Router) Reload(configRoutes []config.Route) {
	newRoutes := make([]*Route, len(configRoutes))
	for i, cfg := range configRoutes {
		newRoutes[i] = NewRoute(cfg)
	}
	r.routes.Store(newRoutes)
}

// Match finds the route with the longest matching prefix for the given path
func (r *Router) Match(path string) (*Route, bool) {
	routes := r.routes.Load().([]*Route)
	
	var bestMatch *Route
	var bestLen int

	for _, route := range routes {
		prefix := route.Config.Path
		
		// Exact match or prefix match (checking if path continues properly)
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			if len(prefix) > bestLen {
				bestMatch = route
				bestLen = len(prefix)
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, true
	}
	
	// Fallback check for exact prefix matching (e.g. prefix is "/" or "/api")
	for _, route := range routes {
		prefix := route.Config.Path
		if strings.HasPrefix(path, prefix) {
			if len(prefix) > bestLen {
				bestMatch = route
				bestLen = len(prefix)
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, true
	}

	return nil, false
}
