package loadbalancer

import (
	"sync"
	"testing"
)

func TestRoundRobin_EvenDistribution(t *testing.T) {
	lb := NewRoundRobin()
	upstreams := []*Upstream{
		NewUpstream("A", true),
		NewUpstream("B", true),
		NewUpstream("C", true),
	}

	if u := lb.Next(upstreams); u == nil || u.URL != "A" {
		t.Errorf("Expected A, got %v", u)
	}
	if u := lb.Next(upstreams); u == nil || u.URL != "B" {
		t.Errorf("Expected B, got %v", u)
	}
	if u := lb.Next(upstreams); u == nil || u.URL != "C" {
		t.Errorf("Expected C, got %v", u)
	}
	if u := lb.Next(upstreams); u == nil || u.URL != "A" {
		t.Errorf("Expected A again, got %v", u)
	}
}

func TestRoundRobin_SkipsUnhealthyUpstreams(t *testing.T) {
	lb := NewRoundRobin()
	upstreams := []*Upstream{
		NewUpstream("A", true),
		NewUpstream("B", false),
		NewUpstream("C", true),
	}

	if u := lb.Next(upstreams); u == nil || u.URL != "A" {
		t.Errorf("Expected A, got %v", u)
	}
	if u := lb.Next(upstreams); u == nil || u.URL != "C" {
		t.Errorf("Expected C, got %v", u)
	}
}

func TestRoundRobin_AllUnhealthy(t *testing.T) {
	lb := NewRoundRobin()
	upstreams := []*Upstream{
		NewUpstream("A", false),
		NewUpstream("B", false),
	}

	if u := lb.Next(upstreams); u != nil {
		t.Errorf("Expected nil, got %v", u)
	}
}

func TestRoundRobin_Concurrent(t *testing.T) {
	lb := NewRoundRobin()
	upstreams := []*Upstream{
		NewUpstream("A", true),
		NewUpstream("B", true),
	}

	var wg sync.WaitGroup
	// Run 1000 concurrent requests to test thread-safety
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u := lb.Next(upstreams)
			if u == nil {
				t.Error("Expected an upstream, got nil")
			}
		}()
	}
	wg.Wait()
}
