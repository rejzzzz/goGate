package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/redis/go-redis/v9"
	"github.com/rejzzzz/goGate/internal/circuitbreaker"
	"github.com/rejzzzz/goGate/internal/config"
	"github.com/rejzzzz/goGate/internal/healthcheck"
	"github.com/rejzzzz/goGate/internal/loadbalancer"
	"github.com/rejzzzz/goGate/internal/metrics"
	"github.com/rejzzzz/goGate/internal/middleware"
	"github.com/rejzzzz/goGate/internal/proxy"
	"github.com/rejzzzz/goGate/internal/ratelimit"
	"github.com/rejzzzz/goGate/internal/router"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting Distributed API Gateway...")

	// Initialize Metrics
	metrics.Init()

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

	// Initialize Redis for Rate Limiting
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
	})
	redisStore := ratelimit.NewRedisStore(redisClient)
	if err := redisStore.LoadScript(context.Background()); err != nil {
		logger.Warn("Failed to preload Redis Lua script", zap.Error(err))
	}

	// 3. Build Upstream Map (Name -> List of *loadbalancer.Upstream)
	upstreamMap := make(map[string][]*loadbalancer.Upstream)
	for _, group := range cfg.UpstreamGroups {
		var ups []*loadbalancer.Upstream
		cbConfig := &circuitbreaker.Config{
			FailureThreshold: group.CircuitBreaker.FailureThreshold,
			SuccessThreshold: group.CircuitBreaker.SuccessThreshold,
			Timeout:          group.CircuitBreaker.Timeout,
		}
		for _, u := range group.Upstreams {
			up := loadbalancer.NewUpstream(u.URL, true)
			up.CircuitBreaker = circuitbreaker.NewBreaker(cbConfig)
			ups = append(ups, up)
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

	// 5. Initialize Proxies
	p := proxy.NewHTTPProxy(nil)
	grpcProxy := proxy.NewGRPCProxy()

	// 6. Build Root HTTP Handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(registry.GetAll())
			return
		}

		if req.URL.Path == "/metrics" {
			metrics.Handler().ServeHTTP(w, req)
			return
		}

		// Match Route from context (set by RouteMatch middleware)
		route, ok := req.Context().Value(router.RouteContextKey).(*router.Route)
		if !ok || route == nil {
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

		if target.CircuitBreaker != nil && !target.CircuitBreaker.Allow() {
			w.Header().Set("X-Circuit-Breaker", "open")
			http.Error(w, "Service Unavailable: Circuit Breaker Open", http.StatusServiceUnavailable)
			return
		}

		target.ActiveConnections.Add(1)
		defer target.ActiveConnections.Add(-1)

		var stripPrefix string
		if route.Config.StripPrefix {
			stripPrefix = route.Config.Path
		}

		// Proxy request
		if strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc") {
			// Inject target for grpc director
			ctx := context.WithValue(req.Context(), proxy.TargetContextKey, target)
			req = req.WithContext(ctx)
			grpcProxy.Server.ServeHTTP(w, req)
		} else {
			p.ServeHTTP(w, req, target, stripPrefix)
		}
	})

	// Wrap with Middleware Chain
	finalHandler := middleware.Chain(
		handler,
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.Logging(logger),
		middleware.RouteMatch(r),
		middleware.Metrics(),
		middleware.RateLimit(redisStore),
	)

	// 7. Start HTTP Server with h2c for cleartext HTTP/2 (gRPC)
	port := cfg.Server.Port
	if port == 0 {
		port = 8080 // default
	}
	
	addr := fmt.Sprintf(":%d", port)
	
	srv := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(finalHandler, &http2.Server{}),
	}

	go func() {
		logger.Info("Gateway listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 30 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	
	logger.Info("Closing Redis connection pool...")
	if err := redisClient.Close(); err != nil {
		logger.Error("Error closing Redis pool", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}
