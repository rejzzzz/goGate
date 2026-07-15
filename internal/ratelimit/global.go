package ratelimit

import (
	"math"
	"sync/atomic"
	"time"
)

// GlobalLimiter is an in-memory lock-free token bucket for global gateway traffic shaping.
type GlobalLimiter struct {
	tokens atomic.Int64
	burst  int64
	stop   chan struct{}
}

// NewGlobalLimiter creates a new global in-memory rate limiter using an atomic counter.
func NewGlobalLimiter(rateLimit float64, burst int) *GlobalLimiter {
	if rateLimit <= 0 {
		return nil // Disabled
	}

	gl := &GlobalLimiter{
		burst: int64(burst),
		stop:  make(chan struct{}),
	}
	
	// Start with full bucket
	gl.tokens.Store(int64(burst))

	// Refill loop (e.g., every 10ms for smooth refill)
	ticker := time.NewTicker(10 * time.Millisecond)
	go func() {
		tokensPerTick := rateLimit * 0.01 // 10ms is 0.01s
		var fractionalTokens float64

		for {
			select {
			case <-ticker.C:
				fractionalTokens += tokensPerTick
				wholeTokens := int64(math.Floor(fractionalTokens))
				
				if wholeTokens > 0 {
					fractionalTokens -= float64(wholeTokens)
					
					// Atomically add tokens up to burst
					for {
						current := gl.tokens.Load()
						if current >= gl.burst {
							break // Bucket full
						}
						
						next := current + wholeTokens
						if next > gl.burst {
							next = gl.burst
						}
						
						if gl.tokens.CompareAndSwap(current, next) {
							break
						}
					}
				}
			case <-gl.stop:
				ticker.Stop()
				return
			}
		}
	}()

	return gl
}

// Allow checks if the request is permitted by the global limit.
// Uses a lock-free CompareAndSwap loop.
func (g *GlobalLimiter) Allow() bool {
	if g == nil {
		return true // Disabled
	}
	
	for {
		current := g.tokens.Load()
		if current <= 0 {
			return false
		}
		
		if g.tokens.CompareAndSwap(current, current-1) {
			return true
		}
	}
}

// Tokens returns the approximate current tokens available.
func (g *GlobalLimiter) Tokens() float64 {
	if g == nil {
		return 0
	}
	return float64(g.tokens.Load())
}

// Stop halts the background refill goroutine
func (g *GlobalLimiter) Stop() {
	if g != nil && g.stop != nil {
		close(g.stop)
	}
}
