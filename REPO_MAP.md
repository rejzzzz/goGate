# Distributed API Gateway - Repository Map

Welcome to the `distributed_api_gateway` project. This document serves as a high-level map of the codebase to help you orient yourself quickly.

## Request Flow

A typical incoming HTTP request follows this lifecycle:

1. **Entry Point (`cmd/gateway/main.go`)**: Initializes configurations, sets up the router, and starts the HTTP server.
2. **Router (`internal/router`)**: Matches the incoming request path/method to a registered route.
3. **Middleware Chain (`internal/middleware`)**: 
   - Logging, Auth, CORS, etc.
   - **Rate Limiter (`internal/ratelimit`)**: Checks if the client has exceeded their request quota.
4. **Proxy (`internal/proxy`)**: Forwards the request to the backend service.
   - **Load Balancer (`internal/loadbalancer`)**: Selects the appropriate backend instance if multiple exist.
   - **Circuit Breaker (`internal/circuitbreaker`)**: Fails fast if the downstream service is unhealthy.
5. **Backends**: The request reaches the actual microservice.

## Repository Structure

```text
.
├── admin-ui/          # Frontend assets and code for the Gateway Admin Dashboard
├── bin/               # Compiled binaries
├── cmd/
│   └── gateway/       # Main entry point for the API Gateway
├── configs/           # Configuration files (gateway.yaml)
├── deploy/            # Deployment files (Docker, Kubernetes)
├── docs/              # Additional documentation
├── examples/
│   └── backends/      # Test mock backend services (service-a, b, c, d)
├── internal/          # Core private gateway code
│   ├── admin/         # Admin API server
│   ├── app/           # Core Gateway application initialization
│   ├── circuitbreaker/# Circuit breaker implementation
│   ├── config/        # Configuration loader
│   ├── discovery/     # Service discovery integrations (Consul, Etcd)
│   ├── healthcheck/   # Active health checking logic
│   ├── loadbalancer/  # Load balancing strategies
│   ├── metrics/       # Prometheus metrics collection
│   ├── middleware/    # HTTP middlewares (auth, ratelimit, logs)
│   ├── proxy/         # Reverse proxy implementations (HTTP/gRPC)
│   ├── ratelimit/     # Redis-backed rate limiter
│   └── router/        # Route matching logic
└── scripts/           # Testing and helper scripts
    ├── k6/            # k6 load testing scripts
    ├── stress_test.go # Native Go stress test script
    └── test_api.ps1   # PowerShell endpoint tester
```

## Important Development Notes
- **Internal vs Public**: All core gateway logic resides in `internal/`. These packages cannot be imported by external Go modules.
- **Config Driven**: The gateway routes and behaviors are primarily driven by the configurations in `configs/`.
