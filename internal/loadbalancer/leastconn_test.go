package loadbalancer

import (
	"sync"
	"testing"
)

func TestLeastConnections_PicksLowest(t *testing.T) {
	lc := NewLeastConnections()
	
	upstreams := []*Upstream{
		NewUpstream("A", true),
		NewUpstream("B", true),
		NewUpstream("C", true),
	}

	// Set active connections
	upstreams[0].ActiveConnections.Store(10)
	upstreams[1].ActiveConnections.Store(2)
	upstreams[2].ActiveConnections.Store(5)

	if u := lc.Next(upstreams); u == nil || u.URL != "B" {
		t.Errorf("Expected B (lowest), got %v", u)
	}
}

func TestLeastConnections_SkipsUnhealthy(t *testing.T) {
	lc := NewLeastConnections()
	
	upstreams := []*Upstream{
		NewUpstream("A", false),
		NewUpstream("B", true),
		NewUpstream("C", true),
	}

	// Set active connections (A is lowest, but unhealthy)
	upstreams[0].ActiveConnections.Store(0)
	upstreams[1].ActiveConnections.Store(10)
	upstreams[2].ActiveConnections.Store(20)

	if u := lc.Next(upstreams); u == nil || u.URL != "B" {
		t.Errorf("Expected B (lowest healthy), got %v", u)
	}
}

func TestLeastConnections_TieBreaking(t *testing.T) {
	lc := NewLeastConnections()
	
	upstreams := []*Upstream{
		NewUpstream("A", true),
		NewUpstream("B", true),
		NewUpstream("C", true),
	}

	// Set active connections (A and C tied for lowest)
	upstreams[0].ActiveConnections.Store(5)
	upstreams[1].ActiveConnections.Store(10)
	upstreams[2].ActiveConnections.Store(5)

	// First pick should be A
	if u := lc.Next(upstreams); u == nil || u.URL != "A" {
		t.Errorf("Expected A, got %v", u)
	}
	
	// Second pick should be C
	if u := lc.Next(upstreams); u == nil || u.URL != "C" {
		t.Errorf("Expected C, got %v", u)
	}

	// Third pick should be A again
	if u := lc.Next(upstreams); u == nil || u.URL != "A" {
		t.Errorf("Expected A, got %v", u)
	}
}

func TestLeastConnections_Concurrent(t *testing.T) {
	lc := NewLeastConnections()
	upstreams := []*Upstream{
		NewUpstream("A", true),
		NewUpstream("B", true),
	}
	upstreams[0].ActiveConnections.Store(5)
	upstreams[1].ActiveConnections.Store(5)

	var wg sync.WaitGroup
	// Run 1000 concurrent requests to test tie-breaker thread-safety
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u := lc.Next(upstreams)
			if u == nil {
				t.Error("Expected an upstream, got nil")
			}
		}()
	}
	wg.Wait()
}
