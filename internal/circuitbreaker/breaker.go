package circuitbreaker

import (
	"sync"
	"time"
)

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
	mu                  sync.Mutex
	state               State
	window              *Window
	consecutiveSuccesses int
	lastStateChange     time.Time
	config              *Config
}

type Config struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
	WindowSize       time.Duration
	BucketCount      int
}

// NewBreaker creates a new circuit breaker
func NewBreaker(config *Config) *Breaker {
	if config.WindowSize == 0 {
		config.WindowSize = 10 * time.Second
	}
	if config.BucketCount == 0 {
		config.BucketCount = 10
	}
	return &Breaker{
		state:  StateClosed,
		window: NewWindow(config.WindowSize, config.BucketCount),
		config: config,
	}
}

// Allow checks if a request is allowed through the circuit breaker
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(b.lastStateChange) >= b.config.Timeout {
			b.state = StateHalfOpen
			b.lastStateChange = time.Now()
			return true // Allow one probe request
		}
		return false
	case StateHalfOpen:
		// We only allow one request to probe, if it's already in HalfOpen, 
		// another request comes in, we reject it until the probe resolves.
		return false
	}
	return false
}

// RecordSuccess records a successful upstream call
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		b.window.RecordSuccess()
	case StateHalfOpen:
		b.consecutiveSuccesses++
		if b.consecutiveSuccesses >= b.config.SuccessThreshold {
			b.state = StateClosed
			b.lastStateChange = time.Now()
			b.consecutiveSuccesses = 0
			b.window.Reset()
		}
	}
}

// RecordFailure records a failed upstream call
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		b.window.RecordFailure()
		if b.window.FailureCount() >= b.config.FailureThreshold {
			b.state = StateOpen
			b.lastStateChange = time.Now()
			b.window.Reset()
		}
	case StateHalfOpen:
		b.state = StateOpen
		b.lastStateChange = time.Now()
		b.consecutiveSuccesses = 0
	}
}

// State returns the current state of the circuit breaker
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Reset manually resets the circuit breaker to Closed state
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = StateClosed
	b.lastStateChange = time.Now()
	b.consecutiveSuccesses = 0
	b.window.Reset()
}
