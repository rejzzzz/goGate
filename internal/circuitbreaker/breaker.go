package circuitbreaker

import "time"

// breaker.go - Circuit breaker state machine
//
// Responsibilities:
// - Maintain one of three states: Closed, Open, or Half-Open
// - Track failures in sliding window when Closed
// - Transition Closed → Open when failures exceed threshold
// - Transition Open → Half-Open after timeout elapses
// - Allow single probe request when Half-Open
// - Transition Half-Open → Closed on probe success
// - Transition Half-Open → Open on probe failure
// - Count upstream call as failed if HTTP 5xx or timeout
//
// Key Functions:
// - NewBreaker(config *Config) *Breaker: Create circuit breaker with configuration
// - Allow() bool: Check if request is allowed (false if Open, true otherwise)
// - RecordSuccess(): Record successful upstream call
// - RecordFailure(): Record failed upstream call
// - State() State: Get current state (Closed, Open, Half-Open)
// - Reset(): Manually reset to Closed state (for admin API)
//
// State Machine:
// - Closed: Normal operation, count failures in window
// - Open: All requests rejected immediately, wait for timeout
// - Half-Open: Allow 1 probe request, others rejected
//
// Inputs:
// - Configuration: failure_threshold, success_threshold, timeout
// - Success/failure events from proxy
//
// Outputs:
// - Allow/reject decision for each request
// - Current state for metrics and admin API

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type Breaker struct {
	state           State
	failureCount    int
	successCount    int
	lastStateChange time.Time
	config          *Config
}

type Config struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
}

// NewBreaker creates a new circuit breaker
func NewBreaker(config *Config) *Breaker {
	// TODO: Implement circuit breaker initialization
	return &Breaker{}
}

// Allow checks if a request is allowed through the circuit breaker
func (b *Breaker) Allow() bool {
	// TODO: Implement circuit breaker logic
	return true
}
