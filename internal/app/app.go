package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/redis/go-redis/v9"
	"github.com/rejzzzz/goGate/internal/admin"
	"github.com/rejzzzz/goGate/internal/circuitbreaker"
	"github.com/rejzzzz/goGate/internal/config"
	"github.com/rejzzzz/goGate/internal/discovery"
	"github.com/rejzzzz/goGate/internal/healthcheck"
	"github.com/rejzzzz/goGate/internal/loadbalancer"
	"github.com/rejzzzz/goGate/internal/metrics"
	"github.com/rejzzzz/goGate/internal/middleware"
	"github.com/rejzzzz/goGate/internal/proxy"
	"github.com/rejzzzz/goGate/internal/ratelimit"
	"github.com/rejzzzz/goGate/internal/router"
	"go.uber.org/zap"
)

// Run initializes and starts the API Gateway.
func Run(configPath string) {
	prodConfig := zap.NewProductionConfig()
	prodConfig.Sampling = &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}
	logger, _ := prodConfig.Build()
	defer logger.Sync()

	logger.Info("Starting Distributed API Gateway...")

	// Initialize Metrics
	metrics.Init()

	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found or error loading it, relying on system environment variables")
	}

	// 1. Load Configuration
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
	initialUpstreamMap := make(map[string][]*loadbalancer.Upstream)
	var upstreamMap atomic.Value
	for _, group := range cfg.UpstreamGroups {
		var ups []*loadbalancer.Upstream
		cbConfig := &circuitbreaker.Config{
			FailureThreshold: group.CircuitBreaker.FailureThreshold,
			SuccessThreshold: group.CircuitBreaker.SuccessThreshold,
			Timeout:          group.CircuitBreaker.Timeout,
		}

		if group.Discovery != nil {
			var provider discovery.Provider
			var err error

			switch group.Discovery.Provider {
			case "consul":
				provider, err = discovery.NewConsulProvider(group.Discovery.Address)
			case "etcd":
				provider, err = discovery.NewEtcdProvider(group.Discovery.Address)
			default:
				err = fmt.Errorf("unsupported discovery provider: %s", group.Discovery.Provider)
			}

			if err != nil {
				logger.Error("Failed to initialize discovery provider", zap.Error(err), zap.String("group", group.Name))
			} else {
				// Initial fetch
				instances, err := provider.GetInstances(group.Discovery.ServiceName)
				if err != nil {
					logger.Error("Failed initial fetch from discovery provider", zap.Error(err), zap.String("group", group.Name))
				} else {
					for _, u := range instances {
						up := loadbalancer.NewUpstream(u, true)
						up.CircuitBreaker = circuitbreaker.NewBreaker(cbConfig)
						ups = append(ups, up)
					}
					logger.Info("Discovered initial upstreams", zap.String("group", group.Name), zap.Int("count", len(ups)))
				}

				// Start watching for updates
				updateCh := make(chan []string)
				go provider.Watch(context.Background(), group.Discovery.ServiceName, updateCh)
				go func(gName string, cb *circuitbreaker.Config) {
					for newInstances := range updateCh {
						logger.Info("Discovery update received", zap.String("group", gName), zap.Int("count", len(newInstances)))

						// Read current map safely
						currentMap, _ := upstreamMap.Load().(map[string][]*loadbalancer.Upstream)

						// Create a deep copy of the map
						newMap := make(map[string][]*loadbalancer.Upstream)
						for k, v := range currentMap {
							newMap[k] = v
						}

						// Build new upstream list for this group
						var newUps []*loadbalancer.Upstream
						for _, u := range newInstances {
							up := loadbalancer.NewUpstream(u, true)
							up.CircuitBreaker = circuitbreaker.NewBreaker(cb)
							newUps = append(newUps, up)
						}

						newMap[gName] = newUps
						upstreamMap.Store(newMap)
					}
				}(group.Name, cbConfig)
			}
		} else {
			// Static upstreams
			for _, u := range group.Upstreams {
				up := loadbalancer.NewUpstream(u.URL, true)
				up.CircuitBreaker = circuitbreaker.NewBreaker(cbConfig)
				ups = append(ups, up)
			}
		}

		initialUpstreamMap[group.Name] = ups
	}

	upstreamMap.Store(initialUpstreamMap)

	// Sync loop: Update load balancer upstreams from health registry without blocking requests
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			healths := registry.GetAll()
			currentMap, _ := upstreamMap.Load().(map[string][]*loadbalancer.Upstream)
			for _, ups := range currentMap {
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

	// 3. Initialize Admin Server
	reloadChan := make(chan struct{}, 1)
	adminPort := cfg.Server.AdminPort
	if adminPort == 0 {
		adminPort = 9090
	}
	adminServer := admin.NewServer(adminPort, r, &upstreamMap, registry, reloadChan, cfg.Metrics.PrometheusURL)

	// Hot Reload goroutine
	go func() {
		for range reloadChan {
			logger.Info("Reloading configuration...")
			newCfg, err := config.Load(configPath)
			if err != nil {
				logger.Error("Failed to hot-reload config", zap.Error(err))
				continue
			}

			// Reload routes
			r.Reload(newCfg.Routes)
			logger.Info("Routes reloaded successfully", zap.Int("routeCount", len(newCfg.Routes)))
		}
	}()
	// 5. Initialize Proxies
	p := proxy.NewHTTPProxy(proxy.NewTransport())
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
		currentMap, _ := upstreamMap.Load().(map[string][]*loadbalancer.Upstream)
		ups, exists := currentMap[route.Config.UpstreamGroup]
		if !exists || len(ups) == 0 {
			http.Error(w, "Bad Gateway: No Upstreams Available", http.StatusBadGateway)
			return
		}

		// Select upstream using route's load balancer
		target := route.LB.Next(ups)
		if target == nil {
			http.Error(w, "Bad Gateway: No Healthy Upstreams Available", http.StatusBadGateway)
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

	// Parse trusted proxies
	trustedProxies := middleware.ParseTrustedProxies(cfg.TrustedProxies, logger)

	// Initialize Global Rate Limiter
	globalLimiter := ratelimit.NewGlobalLimiter(cfg.GlobalRateLimit.RequestsPerSecond, cfg.GlobalRateLimit.Burst)

	// Wrap with Middleware Chain
	finalHandler := middleware.Chain(
		handler,
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.Logging(logger, trustedProxies),
		middleware.RouteMatch(r),
		middleware.Auth(cfg.Auth),
		middleware.Metrics(),
		middleware.RateLimit(redisStore, globalLimiter, *cfg, trustedProxies),
	)

	// 7. Start HTTP Server with h2c for cleartext HTTP/2 (gRPC)
	port := cfg.Server.Port
	if port == 0 {
		port = 8080 // default
	}

	addr := fmt.Sprintf(":%d", port)

	readTimeout := cfg.Server.ReadTimeout
	if readTimeout == 0 {
		readTimeout = 5 * time.Second
	}
	writeTimeout := cfg.Server.WriteTimeout
	if writeTimeout == 0 {
		writeTimeout = 10 * time.Second
	}
	idleTimeout := cfg.Server.IdleTimeout
	if idleTimeout == 0 {
		idleTimeout = 120 * time.Second
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      h2c.NewHandler(finalHandler, &http2.Server{}),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	go func() {
		logger.Info("Gateway listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("Admin API listening", zap.Int("port", adminPort))
		if err := adminServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("Admin server error", zap.Error(err))
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

	if err := adminServer.Stop(ctx); err != nil {
		logger.Error("Admin server forced to shutdown", zap.Error(err))
	}

	logger.Info("Closing Redis connection pool...")
	if err := redisClient.Close(); err != nil {
		logger.Error("Error closing Redis pool", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}
