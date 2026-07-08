package ratelimit

// tokenbucket.go - Token bucket rate limiting algorithm
//
// Responsibilities:
// - Implement token bucket algorithm with burst support
// - Calculate tokens added since last refill based on elapsed time
// - Cap tokens at burst limit (max bucket size)
// - Deduct 1 token per allowed request
// - Reject request if tokens < 1
//
// Key Functions:
// - NewTokenBucket(rate float64, burst int) *TokenBucket: Create token bucket
// - Allow(now time.Time) (allowed bool, remaining int): Check if request allowed
// - calculateTokens(lastRefill time.Time, now time.Time, currentTokens float64) float64: Calculate new token count
//
// Algorithm:
// 1. Get current token count and last refill timestamp
// 2. Calculate elapsed time since last refill
// 3. Add tokens: elapsed_seconds * rate (capped at burst)
// 4. If tokens >= 1: deduct 1 and allow request
// 5. If tokens < 1: reject request
//
// Inputs:
// - Rate: tokens per second (e.g., 100)
// - Burst: max bucket size (e.g., 20)
// - Current timestamp
//
// Outputs:
// - Allowed/rejected decision
// - Remaining tokens (for X-RateLimit-Remaining header)

import "time"

type TokenBucket struct {
	rate       float64 // Tokens per second
	burst      int     // Max bucket size
	tokens     float64
	lastRefill time.Time
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(rate float64, burst int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed under rate limit
func (tb *TokenBucket) Allow(now time.Time) (allowed bool, remaining int) {
	elapsed := now.Sub(tb.lastRefill).Seconds()
	
	// Add tokens based on time elapsed
	tb.tokens += elapsed * tb.rate
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}
	tb.lastRefill = now

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true, int(tb.tokens)
	}

	return false, 0
}
