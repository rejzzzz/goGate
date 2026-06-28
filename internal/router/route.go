package router

import "github.com/yourusername/api-gateway/internal/config"

// Route is an internal representation of a configured route
type Route struct {
	Config config.Route
}

// NewRoute creates a router Route from a config Route
func NewRoute(cfg config.Route) *Route {
	return &Route{
		Config: cfg,
	}
}
