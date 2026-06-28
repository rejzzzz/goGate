package router

import (
	"github.com/rejzzzz/goGate/internal/config"
	"github.com/rejzzzz/goGate/internal/loadbalancer"
)

// Route is an internal representation of a configured route
type Route struct {
	Config config.Route
	LB     loadbalancer.LoadBalancer
}

// NewRoute creates a router Route from a config Route
func NewRoute(cfg config.Route) *Route {
	return &Route{
		Config: cfg,
		LB:     loadbalancer.New(cfg.LoadBalancer),
	}
}
