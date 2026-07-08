package circuitbreaker

// breaker_test.go - Unit tests for circuit breaker state machine
//
// Test Cases:
// - TestBreaker_ClosedToOpen: Verify transitions to Open after failure_threshold failures
// - TestBreaker_OpenToHalfOpen: Verify transitions to Half-Open after timeout
// - TestBreaker_HalfOpenToClosed: Verify transitions to Closed on probe success
// - TestBreaker_HalfOpenToOpen: Verify transitions to Open on probe failure
// - TestBreaker_RejectsWhenOpen: Verify Allow() returns false when Open
// - TestBreaker_AllowsProbeWhenHalfOpen: Verify Allow() returns true once when Half-Open
// - TestBreaker_SlidingWindow: Verify uses sliding window, not fixed window
// - TestBreaker_ManualReset: Verify Reset() transitions to Closed immediately
//
// Inputs: Mock upstream success/failure events
// Outputs: Assertions on state transitions and Allow() return values

import (
	"testing"
	"time"
)

func TestBreaker_ClosedToOpen(t *testing.T) {
	cb := NewBreaker(&Config{
		FailureThreshold: 3,
		Timeout:          1 * time.Second,
	})

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.State())
	}
	if cb.Allow() {
		t.Fatal("expected Allow() to return false when Open")
	}
}

func TestBreaker_OpenToHalfOpen(t *testing.T) {
	cb := NewBreaker(&Config{
		FailureThreshold: 1,
		Timeout:          10 * time.Millisecond,
	})

	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.State())
	}

	// wait for timeout
	time.Sleep(15 * time.Millisecond)

	if !cb.Allow() {
		t.Fatal("expected Allow() to return true when transitioning to Half-Open")
	}

	if cb.State() != StateHalfOpen {
		t.Fatalf("expected state Half-Open, got %v", cb.State())
	}

	// second request should be rejected (only 1 probe allowed)
	if cb.Allow() {
		t.Fatal("expected Allow() to return false for second request in Half-Open")
	}
}

func TestBreaker_HalfOpenToClosed(t *testing.T) {
	cb := NewBreaker(&Config{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		Timeout:          10 * time.Millisecond,
	})

	cb.RecordFailure()
	time.Sleep(15 * time.Millisecond)
	cb.Allow() // Trigger Half-Open

	cb.RecordSuccess()
	if cb.State() == StateClosed {
		t.Fatal("expected state to still be Half-Open after 1 success")
	}

	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Fatalf("expected state Closed, got %v", cb.State())
	}
}

func TestBreaker_HalfOpenToOpen(t *testing.T) {
	cb := NewBreaker(&Config{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		Timeout:          10 * time.Millisecond,
	})

	cb.RecordFailure()
	time.Sleep(15 * time.Millisecond)
	cb.Allow() // Trigger Half-Open

	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.State())
	}
}

func TestBreaker_SlidingWindow(t *testing.T) {
	cb := NewBreaker(&Config{
		FailureThreshold: 2,
		WindowSize:       50 * time.Millisecond,
		BucketCount:      5, // 10ms buckets
	})

	cb.RecordFailure()
	// Wait for window to slide past the first failure
	time.Sleep(60 * time.Millisecond)

	cb.RecordFailure()

	// Should still be closed because the first failure expired
	if cb.State() != StateClosed {
		t.Fatalf("expected state Closed due to sliding window expiration, got %v", cb.State())
	}

	// Two failures quickly
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.State())
	}
}

func TestBreaker_ManualReset(t *testing.T) {
	cb := NewBreaker(&Config{
		FailureThreshold: 1,
		Timeout:          1 * time.Second,
	})

	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.State())
	}

	cb.Reset()
	if cb.State() != StateClosed {
		t.Fatalf("expected state Closed after reset, got %v", cb.State())
	}
}
