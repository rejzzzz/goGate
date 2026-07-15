# Distributed API Gateway

A production-grade reverse proxy and API gateway written from scratch in Go. Handles 10,000-20,000 requests per second with P99 latency under 15ms, supporting HTTP/REST and gRPC upstream routing with dynamic load balancing, rate limiting, circuit breaking, and comprehensive observability.

## Features

### Core Gateway Capabilities

- **HTTP & gRPC Reverse Proxying**: Forward requests to multiple upstream services
- **Dynamic Routing**: YAML-based configuration with hot-reload support
- **Load Balancing**: Round-robin and least-connections strategies
- **Circuit Breaker**: Closed/Open/Half-Open state machine for fault isolation
- **Health Checking**: Background polling to detect unhealthy upstreams
- **Rate Limiting**: Redis-backed token bucket with atomic Lua scripts

### Observability

- **Prometheus Metrics**: 10+ metric types for comprehensive monitoring
- **Grafana Dashboard**: Pre-built with 10+ panels for live visualization
- **Structured Logging**: JSON logs with request tracing
- **Admin UI**: React-based interface for system inspection

### Infrastructure

- **Docker Compose**: Full local development environment
- **Graceful Shutdown**: Connection draining with configurable timeout
- **Performance Optimized**: Atomic operations, connection pooling, tuned HTTP transport

## Performance Targets

- **Throughput**: 10,000-20,000 requests per second
- **Latency**: P99 < 15ms (without upstream delay)
- **Rate Limit Overhead**: < 1ms per request (Redis RTT)

## Architecture

```
Clients (k6 / wrk / curl)
        │
        ▼
┌──────────────────────────────┐
│        API Gateway (Go)      │
│  ┌──────────────────────┐    │
│  │   Middleware Chain   │    │
│  │  Recovery → RequestID│    │
│  │  → Logging → Metrics │    │
│  │  → RateLimit → Proxy │    │
│  └──────────────────────┘    │
└──────────────────────────────┘
        │              │
   ┌────┴────┐    ┌────┴─────┐
   │ Upstreams│    │ Redis    │
   │ A, B, C,D│    │(RateLimit│
   └─────────┘    └──────────┘
        │
┌───────────────┐
│ Prometheus    │
│ + Grafana     │
└───────────────┘
```

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make

### Local Development

```bash
# Start all services (gateway + backends + Redis + Prometheus + Grafana)
make run-all

# Gateway: http://localhost:8080
# Admin UI: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9091
```

### Run Individual Components

```bash
# Start only infrastructure (Redis, Prometheus, Grafana)
make run-infra

# Build and run gateway locally
make run-gateway

# Start backend services
make run-backends
```

## Build & Test

```bash
# Build the gateway binary
make build

# Run all tests
make test

# Generate coverage report
make coverage

# Run linter
make lint
```

## Benchmarking

```bash
# Run k6 basic load test (requires gateway running)
make bench-basic

# Run wrk for raw throughput testing
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/users
```

## Project Structure

```
├── cmd/gateway/              # Gateway entry point
├── internal/
│   ├── config/              # Configuration loading & validation
│   ├── proxy/               # HTTP & gRPC reverse proxy
│   ├── router/              # Route matching & dispatch
│   ├── middleware/          # Middleware chain (Recovery, RequestID, Logging, Metrics, RateLimit)
│   ├── loadbalancer/        # Round-robin & least-connections LB
│   ├── healthcheck/         # Background health polling
│   ├── metrics/             # Prometheus instrumentation
│   ├── ratelimit/           # Redis-backed token bucket
│   ├── circuitbreaker/      # State machine for fault isolation
│   └── admin/               # Admin API server & handlers
├── examples/
│   └── backends/            # Example/Mock microservices for testing
├── admin-ui/                 # React admin UI
├── configs/                  # Configuration files
├── scripts/                  # Testing and helper scripts
├── deploy/                   # Deployment files
│   ├── Dockerfile            # Gateway container image
│   ├── docker-compose.yml    # Full stack composition
│   └── prometheus/           # Monitoring configs
└── Makefile                 # Build automation
```

## Configuration

### Environment Variables

Override YAML config values with environment variables:

```bash
GATEWAY_PORT=8080
GATEWAY_ADMIN_PORT=9090
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=secret
GATEWAY_LOG_LEVEL=debug
```

See `.env.example` for all supported variables.

### gateway.yaml

Define routes, upstream groups, rate limits, and load balancing strategies:

```yaml
routes:
    - path: /api/v1/users
      upstream_group: user-service
      load_balancer: round-robin
      rate_limit:
          requests_per_second: 100
          burst: 20

upstream_groups:
    - name: user-service
      upstreams:
          - url: http://service-a:8081
          - url: http://service-b:8082
      health_check:
          path: /health
          interval: 5s
          timeout: 2s
```

## Admin API

The admin server runs on port 9090 and provides:

```
GET  /admin/api/stats                    # Aggregated gateway stats
GET  /admin/api/routes                   # All configured routes
GET  /admin/api/upstreams                # Upstream groups & health
GET  /admin/api/circuit-breakers         # Circuit breaker states
POST /admin/api/circuit-breakers/:id/reset  # Force reset breaker
POST /admin/api/config/reload            # Hot reload configuration
GET  /admin/health                       # Admin server health
```

## Metrics

All metrics exposed at `GET /metrics` in Prometheus format:

- `gateway_requests_total` - Total requests by route/method/status
- `gateway_request_duration_seconds` - Request latency histogram
- `gateway_upstream_request_duration_seconds` - Upstream call latency
- `gateway_rate_limited_total` - Rejected requests
- `gateway_circuit_breaker_state` - Circuit breaker state per upstream
- `gateway_circuit_breaker_trips_total` - Times circuit opened
- `gateway_upstream_health` - Upstream health status
- `gateway_active_connections` - Active connections per upstream
- `gateway_redis_operation_duration_seconds` - Redis operation latency

## Backend Services

Four test services to verify gateway functionality:

- **Service A** (port 8081): Fast REST service (~1ms response)
- **Service B** (port 8082): Slow REST service (50ms delay, 5% errors)
- **Service C** (port 9000): gRPC Echo service
- **Service D** (port 8083): REST service for least-connections testing

## License

MIT
