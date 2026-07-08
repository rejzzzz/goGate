package circuitbreaker

// window.go - Sliding window failure counter
//
// Responsibilities:
// - Track success/failure counts in a sliding time window
// - Divide window into fixed-size time buckets (e.g., 10-second buckets)
// - Expire old buckets as time advances
// - Calculate total failures in current window
// - Support concurrent access from multiple request goroutines
//
// Key Functions:
// - NewWindow(windowSize time.Duration, bucketCount int) *Window: Create sliding window
// - RecordSuccess(): Record successful request in current bucket
// - RecordFailure(): Record failed request in current bucket
// - FailureCount() int: Get total failures in current window
// - Reset(): Clear all buckets
//
// Implementation Details:
// - Use circular buffer of size N for buckets
// - Each bucket tracks: success count, failure count, start timestamp
// - When accessing, first expire buckets older than window size
// - Use sync.Mutex to protect bucket operations
//
// Inputs: Success/failure events from circuit breaker
// Outputs: Total failure count in sliding window

import (
	"sync"
	"time"
)

type Window struct {
	mu           sync.Mutex
	buckets      []bucket
	size         time.Duration
	bucketLength time.Duration
}

type bucket struct {
	failures  int
	successes int
	timestamp time.Time
}

// NewWindow creates a new sliding window for failure tracking
func NewWindow(windowSize time.Duration, bucketCount int) *Window {
	if bucketCount <= 0 {
		bucketCount = 10
	}
	return &Window{
		buckets:      make([]bucket, bucketCount),
		size:         windowSize,
		bucketLength: windowSize / time.Duration(bucketCount),
	}
}

func (w *Window) getCurrentBucket() *bucket {
	now := time.Now()
	// Find index of current bucket
	index := (now.UnixNano() / int64(w.bucketLength)) % int64(len(w.buckets))
	b := &w.buckets[index]

	// If bucket is too old, reset it
	if now.Sub(b.timestamp) > w.size {
		b.failures = 0
		b.successes = 0
		b.timestamp = now
	}
	return b
}

func (w *Window) RecordSuccess() {
	w.mu.Lock()
	defer w.mu.Unlock()
	b := w.getCurrentBucket()
	b.successes++
}

func (w *Window) RecordFailure() {
	w.mu.Lock()
	defer w.mu.Unlock()
	b := w.getCurrentBucket()
	b.failures++
}

func (w *Window) FailureCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	totalFailures := 0

	for i := range w.buckets {
		b := &w.buckets[i]
		if now.Sub(b.timestamp) <= w.size {
			totalFailures += b.failures
		}
	}
	return totalFailures
}

func (w *Window) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i := range w.buckets {
		w.buckets[i] = bucket{}
	}
}
