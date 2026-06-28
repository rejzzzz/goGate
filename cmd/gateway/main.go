package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rejzzzz/goGate/internal/config"
	"github.com/rejzzzz/goGate/internal/healthcheck"
	"github.com/rejzzzz/goGate/internal/loadbalancer"
	"github.com/rejzzzz/goGate/internal/proxy"
	"github.com/rejzzzz/goGate/internal/router"
)

func main() {
	log.Println("Starting Distributed API Gateway...")

	// 1. Load Configuration
	configPath := "configs/gateway.yaml"
	if envPath := os.Getenv("GATEWAY_CONFIG"); envPath != "" {
		configPath = envPath
	}
	
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Health Checker
	registry := healthcheck.NewRegistry()
	checker := healthcheck.NewChecker(registry)
	checker.Start(cfg.UpstreamGroups)

	// 3. Build Upstream Map (Name -> List of *loadbalancer.Upstream)
	upstreamMap := make(map[string][]*loadbalancer.Upstream)
	for _, group := range cfg.UpstreamGroups {
		var ups []*loadbalancer.Upstream
		for _, u := range group.Upstreams {
			ups = append(ups, loadbalancer.NewUpstream(u.URL, true))
		}
		upstreamMap[group.Name] = ups
	}

	// Sync loop: Update load balancer upstreams from health registry without blocking requests
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			healths := registry.GetAll()
			for _, ups := range upstreamMap {
				for _, u := range ups {
					healthy, exists := healths[u.URL]
					if !exists {
						healthy = true
					}
					u.Healthy.Store(healthy)
				}
			}
		}
	}()

	// 4. Initialize Router
	r := router.New(cfg.Routes)
	log.Printf("Loaded %d routes", len(cfg.Routes))

	// 5. Initialize Proxy
	p := proxy.NewHTTPProxy(nil)

	// 6. Build Root HTTP Handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(registry.GetAll())
			return
		}

		// Match Route
		route, found := r.Match(req.URL.Path)
		if !found {
			http.Error(w, "Not Found: No matching route", http.StatusNotFound)
			return
		}

		// Get upstreams
		ups, exists := upstreamMap[route.Config.UpstreamGroup]
		if !exists || len(ups) == 0 {
			http.Error(w, "Bad Gateway: No Upstreams Available", http.StatusBadGateway)
			return
		}

		// Select upstream using route's load balancer
		target := route.LB.Next(ups)
		if target == nil {
			http.Error(w, "Bad Gateway: No Healthy Upstreams", http.StatusBadGateway)
			return
		}

		var stripPrefix string
		if route.Config.StripPrefix {
			stripPrefix = route.Config.Path
		}

		// Proxy request
		p.ServeHTTP(w, req, target.URL, stripPrefix)
	})

	// 7. Start HTTP Server
	port := cfg.Server.Port
	if port == 0 {
		port = 8080 // default
	}
	
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Gateway listening on %s", addr)
	
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
