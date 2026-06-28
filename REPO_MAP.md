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
├── backends/          # Example/Mock microservices (service-a, service-b) for testing
├── benchmarks/        # Load testing and performance benchmarking scripts (e.g., hey, wrk)
├── cmd/
│   └── gateway/       # Main executable for the API Gateway (`go run ./cmd/gateway`)
├── configs/           # Configuration files (YAML, JSON) defining routes and backend endpoints
├── deploy/            # Deployment manifests (Docker, Kubernetes)
├── docs/              # Additional documentation
└── internal/          # Core gateway logic (private to this module)
    ├── admin/         # HTTP handlers for the Admin API (used by admin-ui)
    ├── circuitbreaker/# Implementation of circuit breaking patterns (e.g., using hystrix-go or custom)
    ├── config/        # Configuration parsing and state management
    ├── healthcheck/   # Active/Passive health checking for backend services
    ├── loadbalancer/  # Load balancing algorithms (Round Robin, Least Connections)
    ├── metrics/       # Prometheus metrics exposition (requests, latency, errors)
    ├── middleware/    # HTTP middlewares (logging, auth, recovery)
    ├── proxy/         # Reverse proxy implementation (e.g., httputil.ReverseProxy wrappers)
    ├── ratelimit/     # Rate limiting logic (token bucket, leaky bucket)
    └── router/        # Route registration and matching
```

## Important Development Notes
- **Internal vs Public**: All core gateway logic resides in `internal/`. These packages cannot be imported by external Go modules.
- **Config Driven**: The gateway routes and behaviors are primarily driven by the configurations in `configs/`.
