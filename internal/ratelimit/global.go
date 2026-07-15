package ratelimit

import (
	"golang.org/x/time/rate"
)

// GlobalLimiter is an in-memory token bucket for global gateway traffic shaping.
// It replaces a Redis-based global rate limit to avoid "hot key" bottlenecks.
type GlobalLimiter struct {
	limiter *rate.Limiter
}

// NewGlobalLimiter creates a new global in-memory rate limiter.
// rateLimit is the maximum requests per second.
// burst is the maximum tokens that can be consumed at once.
func NewGlobalLimiter(rateLimit float64, burst int) *GlobalLimiter {
	if rateLimit <= 0 {
		return nil // Disabled
	}
	return &GlobalLimiter{
		limiter: rate.NewLimiter(rate.Limit(rateLimit), burst),
	}
}

// Allow checks if the request is permitted by the global limit.
// Returns true if allowed, and false if denied.
func (g *GlobalLimiter) Allow() bool {
	if g == nil {
		return true // Disabled
	}
	return g.limiter.Allow()
}

// Tokens returns the approximate current tokens available.
func (g *GlobalLimiter) Tokens() float64 {
	if g == nil {
		return 0
	}
	return g.limiter.Tokens()
}
