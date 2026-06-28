package main

import (
	"github.com/yourusername/api-gateway/internal/config"
	"github.com/yourusername/api-gateway/internal/proxy"
	"github.com/yourusername/api-gateway/internal/router"
	"fmt"
	"log"
	"net/http"
	"os"
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

	// 2. Build Upstream Map (Name -> List of URLs)
	upstreamMap := make(map[string][]string)
	for _, group := range cfg.UpstreamGroups {
		var urls []string
		for _, u := range group.Upstreams {
			urls = append(urls, u.URL)
		}
		upstreamMap[group.Name] = urls
	}

	// 3. Initialize Router
	r := router.New(cfg.Routes)
	log.Printf("Loaded %d routes", len(cfg.Routes))

	// 4. Initialize Proxy
	p := proxy.NewHTTPProxy(nil)

	// 5. Build Root HTTP Handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Match Route
		route, found := r.Match(req.URL.Path)
		if !found {
			http.Error(w, "Not Found: No matching route", http.StatusNotFound)
			return
		}

		// Get upstreams
		urls, exists := upstreamMap[route.Config.UpstreamGroup]
		if !exists || len(urls) == 0 {
			http.Error(w, "Bad Gateway: No Upstreams Available", http.StatusBadGateway)
			return
		}

		// Pick first upstream (shortcut for milestone 1)
		targetURL := urls[0]

		var stripPrefix string
		if route.Config.StripPrefix {
			stripPrefix = route.Config.Path
		}

		// Proxy request
		p.ServeHTTP(w, req, targetURL, stripPrefix)
	})

	// 6. Start HTTP Server
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
