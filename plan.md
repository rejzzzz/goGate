# Distributed API Gateway — Full Project Plan

---

## 1. Project Overview

A production-grade reverse proxy and API gateway written from scratch in Go. No Kong, no Nginx, no Traefik — you own the core. It handles real traffic, enforces rate limits via Redis, load-balances across multiple upstream services, detects failures via circuit breaking, and exports metrics to a live Grafana dashboard.

**Goal:** Benchmark at 10,000–20,000 req/sec, P99 < 15ms, with full observability.

---

## 2. High-Level Architecture

```
Clients (k6 / wrk / curl)
        │
        ▼
┌──────────────────────────────┐
│        API Gateway (Go)      │
│  ┌──────────────────────┐    │
│  │   Middleware Chain   │    │
│  │  Auth → RateLimit →  │    │
│  │  CircuitBreaker →    │    │
│  │  LoadBalancer →      │    │
│  │  Proxy               │    │
│  └──────────────────────┘    │
│  ┌──────────┐ ┌───────────┐  │
│  │  Router  │ │  Metrics  │  │
│  └──────────┘ └───────────┘  │
└──────────────────────────────┘
        │              │
   ┌────┴────┐    ┌────┴─────┐
   │ Upstreams│    │ Redis    │
   │ A, B, C │    │(RateLimit│
   └─────────┘    └──────────┘
        │
┌───────────────┐
│ Prometheus    │
│ + Grafana     │
└───────────────┘
```

---

## 3. Tech Stack

| Layer            | Choice                     | Why                                            |
| ---------------- | -------------------------- | ---------------------------------------------- |
| Gateway language | Go 1.22+                   | Raw performance, stdlib `net/http`, goroutines |
| Config           | YAML + Viper               | Human-readable, hot-reloadable                 |
| Rate limit store | Redis 7                    | Atomic INCR, TTL, sub-ms ops                   |
| Metrics          | Prometheus client_golang   | Industry standard                              |
| Dashboards       | Grafana                    | Panels, alerting                               |
| gRPC proxy       | grpc-go                    | Required for gRPC upstream support             |
| Admin UI         | React + Vite (lightweight) | Route viewer, circuit state, live stats        |
| Load testing     | k6                         | Scripted, realistic traffic simulation         |
| Containers       | Docker + Docker Compose    | Local dev parity with VPS                      |
| Deployment       | Single VPS (2–4 vCPU)      | Real numbers, not local                        |

---

## 4. Repository Structure

```
api-gateway/
├── cmd/
│   └── gateway/
│       └── main.go                  # Entry point + graceful shutdown
│
├── internal/
│   ├── config/
│   │   ├── config.go                # Load & validate YAML config
│   │   └── types.go                 # Config structs
│   │
│   ├── proxy/
│   │   ├── http_proxy.go            # Core HTTP reverse proxy
│   │   ├── grpc_proxy.go            # gRPC reverse proxy (h2c + ForceCodec)
│   │   └── transport.go             # Custom HTTP transport settings
│   │
│   ├── router/
│   │   ├── router.go                # Route matching & dispatch
│   │   └── route.go                 # Route definition struct
│   │
│   ├── middleware/
│   │   ├── chain.go                 # Compose middleware stack
│   │   ├── ratelimit.go             # Token bucket rate limiter
│   │   ├── logging.go               # Structured request logging
│   │   ├── metrics.go               # Prometheus instrumentation
│   │   ├── recovery.go              # Panic recovery
│   │   └── requestid.go             # Inject X-Request-ID header
│   │
│   ├── loadbalancer/
│   │   ├── loadbalancer.go          # LB interface
│   │   ├── roundrobin.go            # Round-robin strategy
│   │   ├── roundrobin_test.go       # Unit tests
│   │   ├── leastconn.go             # Least-connections strategy
│   │   └── leastconn_test.go        # Unit tests
│   │
│   ├── healthcheck/
│   │   ├── checker.go               # Background health poller
│   │   └── registry.go              # Upstream health state store
│   │
│   ├── metrics/
│   │   ├── prometheus.go            # Metric definitions & registration
│   │   └── exporter.go              # /metrics HTTP handler
│   │
│   ├── ratelimit/
│   │   ├── tokenbucket.go           # Token bucket logic
│   │   ├── tokenbucket_test.go      # Unit tests for token math
│   │   └── redis_store.go           # Redis-backed token store (EVALSHA)
│   │
│   ├── circuitbreaker/
│   │   ├── breaker.go               # Closed/Open/Half-Open FSM
│   │   ├── breaker_test.go          # Unit tests for all state transitions
│   │   └── window.go                # Sliding window failure counter
│   │
│   └── admin/
│       ├── server.go                # Admin HTTP server
│       └── handlers.go              # Admin API handlers
│
├── backends/
│   ├── service-a/                   # Fast REST service (port 8081)
│   │   └── main.go
│   ├── service-b/                   # Slow/flaky REST service (port 8082)
│   │   └── main.go
│   ├── service-c/                   # gRPC echo service (port 9000)
│   │   ├── main.go
│   │   └── proto/
│   │       └── echo.proto
│   └── service-d/                   # Second order-service instance (port 8083)
│       └── main.go
│
├── admin-ui/                        # React frontend
│   ├── src/
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Routes.tsx
│   │   │   ├── Upstreams.tsx
│   │   │   └── CircuitBreakers.tsx
│   │   ├── components/
│   │   │   ├── RouteCard.tsx
│   │   │   ├── UpstreamStatus.tsx
│   │   │   ├── CircuitBreakerBadge.tsx
│   │   │   └── MetricsChart.tsx
│   │   └── api/
│   │       └── gateway.ts           # API client for admin endpoints
│   ├── package.json
│   └── vite.config.ts
│
├── configs/
│   └── gateway.yaml                 # Main gateway config file
│
├── deploy/
│   ├── docker-compose.yml
│   ├── prometheus/
│   │   └── prometheus.yml
│   └── grafana/
│       └── provisioning/
│           ├── datasources/
│           │   └── prometheus.yaml  # Auto-connects Prometheus datasource
│           └── dashboards/
│               └── gateway.json     # Pre-built Grafana dashboard JSON
│
├── benchmarks/
│   ├── k6/
│   │   ├── basic_load.js
│   │   ├── rate_limit_test.js
│   │   └── circuit_breaker_test.js
│   └── results/                     # Store benchmark outputs here
│
├── .env.example                     # Template for secrets (redis password, admin token)
├── .gitignore                       # Include .env, bin/, node_modules/
├── Makefile
├── Dockerfile
└── go.mod
```

---

## 5. Gateway Configuration File (gateway.yaml)

```yaml
server:
    port: 8080
    admin_port: 9090
    read_timeout: 10s
    write_timeout: 10s
    idle_timeout: 60s
    shutdown_timeout: 30s

redis:
    addr: "localhost:6379"
    password: ""
    db: 0
    pool_size: 50 # critical at 10k+ req/sec — default of 10 will bottleneck
    min_idle_conns: 10
    max_retries: 3

routes:
    - path: /api/v1/users
      strip_prefix: true
      upstream_group: user-service
      rate_limit:
          requests_per_second: 100
          burst: 20
      load_balancer: round-robin

    - path: /api/v1/orders
      strip_prefix: true
      upstream_group: order-service
      rate_limit:
          requests_per_second: 50
          burst: 10
      load_balancer: least-connections

    - path: /api/v1/grpc
      type: grpc
      upstream_group: grpc-service
      load_balancer: round-robin

upstream_groups:
    - name: user-service
      upstreams:
          - url: http://localhost:8081
          - url: http://localhost:8082
      health_check:
          path: /health
          interval: 5s
          timeout: 2s
      circuit_breaker:
          failure_threshold: 5
          success_threshold: 2
          timeout: 30s

    - name: order-service
      upstreams:
          - url: http://localhost:8083
          - url: http://localhost:8084
      health_check:
          path: /health
          interval: 5s
          timeout: 2s
      circuit_breaker:
          failure_threshold: 3
          success_threshold: 1
          timeout: 15s

    - name: grpc-service
      upstreams:
          - url: localhost:9000
      type: grpc

metrics:
    enabled: true
    path: /metrics

logging:
    level: info
    format: json
```

---

## 6. Backend — Implementation Details

### 6.1 Core HTTP Proxy

Use Go's `httputil.ReverseProxy` as a starting point but wrap it with your own transport and director:

```
Request → Director (rewrite URL, add headers) → Transport → Upstream
Response → ModifyResponse (inject response headers, log) → Client
```

Key things to handle:

- Strip/rewrite path prefixes per route
- Inject `X-Forwarded-For`, `X-Request-ID`, `X-Gateway-Version` headers
- Handle upstream timeouts gracefully
- Return `502 Bad Gateway` on upstream errors, not a Go panic

### 6.2 gRPC Proxy

- Use `google.golang.org/grpc` with `grpc.ForceCodec(proxy.Codec())` for transparent proxying (`grpc.WithCodec` is deprecated since grpc-go v1.47+)
- You don't need to know the proto schema — use raw frame forwarding
- Library to consider: `mwitkow/grpc-proxy` (or implement your own bidirectional stream copier)
- Route gRPC based on the `:path` header (gRPC uses HTTP/2 under the hood)
- **h2c (HTTP/2 cleartext):** For local/VPS deployments without TLS, wrap your listener with `golang.org/x/net/http2/h2c` — gRPC will silently fail over plain HTTP/1.1 without it

### 6.3 Router

- On startup, parse `gateway.yaml` and build a route table
- Match incoming requests by prefix (longest match wins)
- Support hot-reload: watch the config file with `fsnotify`, rebuild route table atomically using `sync/atomic` + pointer swap
- Route struct holds: pattern, upstream group reference, rate limit config, LB strategy

### 6.4 Middleware Chain

Build a classic middleware chain using the standard Go pattern:

```go
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}
```

Order of middleware (outermost to innermost):

```
Recovery → RequestID → Logging → Metrics → RateLimit → Proxy
                                                           │
                                                  (inside Proxy handler)
                                                  LB picks upstream
                                                           │
                                                  CircuitBreaker check
                                                  (per chosen upstream)
                                                           │
                                                  Forward request
```

> **Why this order matters:** The circuit breaker operates per-upstream, so it can only be checked after the load balancer selects a specific upstream. Placing CircuitBreaker as a generic middleware before LB selection is architecturally wrong — you wouldn't know which breaker to check. The CB check lives inside the proxy handler, after `Next()` returns an upstream URL.

### 6.5 Rate Limiting — Token Bucket via Redis

Algorithm:

- Each client (by IP or API key) has a bucket in Redis
- Bucket stores: `{tokens, last_refill_timestamp}`
- On each request: calculate tokens added since last refill, add them (capped at burst), deduct 1
- If tokens < 1: reject with `429 Too Many Requests`
- Use a Lua script for atomic read-modify-write in Redis (critical — avoids race conditions)

Redis key format: `ratelimit:{route}:{client_ip}`

Lua script (run via `EVAL`):

```lua
local key = KEYS[1]
local rate = tonumber(ARGV[1])       -- tokens per second
local burst = tonumber(ARGV[2])      -- max bucket size
local now = tonumber(ARGV[3])        -- current unix time in ms

local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1]) or burst
local ts = tonumber(data[2]) or now

local elapsed = (now - ts) / 1000.0
tokens = math.min(burst, tokens + elapsed * rate)

if tokens >= 1 then
    tokens = tokens - 1
    redis.call("HMSET", key, "tokens", tokens, "ts", now)
    redis.call("EXPIRE", key, 60)
    return 1   -- allowed
else
    redis.call("HMSET", key, "tokens", tokens, "ts", now)
    redis.call("EXPIRE", key, 60)
    return 0   -- denied
end
```

Response headers to set on rejection:

- `Retry-After: <seconds>`
- `X-RateLimit-Limit: <rate>`
- `X-RateLimit-Remaining: 0`

Response headers to set on **every** allowed request too:

- `X-RateLimit-Limit: <rate>`
- `X-RateLimit-Remaining: <floor(current_tokens)>`

> **Performance tip:** Don't call `EVAL` on every request — that re-parses the Lua script each time. Instead, at gateway startup call `SCRIPT LOAD` once to get the SHA digest, then call `EVALSHA <sha> ...` in the hot path. Falls back to `EVAL` if the script was evicted (Redis can evict scripts under memory pressure).

### 6.6 Load Balancing

**Interface:**

```go
type LoadBalancer interface {
    Next(upstreams []*Upstream) *Upstream
}
```

> `Next` must handle the case where all upstreams are unhealthy and return `nil`. Every caller must nil-check the result and return `503` if no upstream is available.

**Round-Robin:**

- Atomic counter `uint64`, increment and mod by len(upstreams)
- Skip unhealthy upstreams: filter the list to healthy-only first, then apply modulo on that filtered slice (not the full list, or your counter skips indices)

**Least-Connections:**

- Each upstream tracks `activeConnections int64` (atomic)
- On request start: increment; on request end (defer): decrement
- Pick the upstream with the lowest current count
- Break ties with round-robin

### 6.7 Circuit Breaker

Three states: `Closed` → `Open` → `Half-Open` → `Closed`

State machine per upstream:

- **Closed:** Normal operation. Count failures in a sliding window
- **Open:** All requests fail immediately with `503 Service Unavailable`. After `timeout` duration, move to Half-Open
- **Half-Open:** Allow 1 request through as a probe. Success → Closed. Failure → Open again

Sliding window: use a circular buffer of size N tracking success/failure per time slot (10-second buckets, 6 buckets = 1 min window).

Configurable per upstream:

- `failure_threshold`: failures before opening (e.g., 5)
- `success_threshold`: successes in half-open before closing (e.g., 2)
- `timeout`: how long to stay open (e.g., 30s)

Set response header `X-Circuit-Breaker: open` when rejecting.

> **Circuit breaker + health check interaction:** Both systems run independently, which can conflict — health check may mark an upstream "healthy" (its `/health` returns 200) while the circuit breaker is open (it was failing real traffic). Rule: an upstream is only routable if the health check passes **AND** the circuit breaker is not open. Circuit breaker state takes priority. Only once the breaker closes (after a successful half-open probe) does normal routing resume.

### 6.9 Graceful Shutdown

This is missing from most hobby gateway implementations and makes yours stand out. On `SIGTERM` or `SIGINT`:

1. Stop the listener — stop accepting new connections
2. Let in-flight requests drain — use `http.Server.Shutdown(ctx)` with a deadline (e.g., 30s)
3. Flush pending metrics to Prometheus
4. Close the Redis connection pool cleanly
5. Exit with code 0

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    log.Fatal("forced shutdown:", err)
}
```

Add `shutdown_timeout` to the server config block:

```yaml
server:
    shutdown_timeout: 30s
```

- Background goroutine per upstream group polls each upstream's `/health` endpoint every N seconds
- Maintains an `UpstreamRegistry` with current health status per upstream URL
- Unhealthy upstreams are skipped by the load balancer
- If ALL upstreams in a group are unhealthy, return `503` with `X-Gateway-Error: no-healthy-upstreams`
- Expose aggregated health at `GET /health` on the gateway itself

Health response format:

```json
{
    "status": "healthy",
    "upstreams": {
        "user-service": {
            "http://localhost:8081": "healthy",
            "http://localhost:8082": "degraded"
        }
    },
    "circuit_breakers": {
        "http://localhost:8081": "closed",
        "http://localhost:8082": "open"
    }
}
```

---

## 7. Dummy Backend Services

### Service A — Fast REST (port 8081)

- `GET /users` → returns a static JSON list
- `GET /users/:id` → returns a single user object
- `GET /health` → returns `{"status": "ok"}`
- Simulates a healthy, fast service (~1ms response)

### Service B — Slow REST (port 8082)

- Same endpoints as Service A but adds `time.Sleep(50ms)` to simulate a slower upstream
- Randomly returns `500` for 5% of requests to trigger circuit breaker testing

### Service C — gRPC Echo (port 9000)

- Proto: `rpc Echo(EchoRequest) returns (EchoResponse)`
- Just echoes back the message with a timestamp
- Used to verify gRPC proxy works end to end

### Service D — Second Order-Service instance (port 8083)

- Identical to Service A but serves `/orders` endpoints
- Exists specifically so `order-service` has two upstreams — **least-connections load balancing is meaningless with a single upstream**
- Simulate different load by throttling one instance with a `time.Sleep` flag

All four services are standalone Go binaries in `/backends/`.

---

## 8. Prometheus Metrics to Expose

Define all metrics in `internal/metrics/prometheus.go`:

| Metric                                      | Type      | Labels                           | Description                             |
| ------------------------------------------- | --------- | -------------------------------- | --------------------------------------- |
| `gateway_requests_total`                    | Counter   | `route`, `method`, `status_code` | Total requests handled                  |
| `gateway_request_duration_seconds`          | Histogram | `route`, `method`                | Request latency buckets                 |
| `gateway_upstream_request_duration_seconds` | Histogram | `upstream`, `route`              | Upstream call latency                   |
| `gateway_rate_limited_total`                | Counter   | `route`                          | Requests rejected by rate limiter       |
| `gateway_circuit_breaker_state`             | Gauge     | `upstream`                       | 0=closed, 1=open, 2=half-open           |
| `gateway_circuit_breaker_trips_total`       | Counter   | `upstream`                       | Total times circuit opened              |
| `gateway_upstream_health`                   | Gauge     | `upstream`                       | 1=healthy, 0=unhealthy                  |
| `gateway_active_connections`                | Gauge     | `upstream`                       | Current active connections per upstream |
| `gateway_redis_operation_duration_seconds`  | Histogram | `operation`                      | Redis latency for rate limiting ops     |

Expose all at `GET /metrics` (standard Prometheus format).

Histogram buckets for latency: `[0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0]`

---

## 9. Admin UI — Pages & Features

A lightweight React app served by the gateway's admin server on port `9090`.

### Pages

#### `/` — Dashboard

- Total requests/sec (live, last 60s)
- P50 / P95 / P99 latency
- Rate limited requests count
- Active circuit breakers (how many are open)
- Error rate percentage
- All numbers pulled from `GET /admin/api/stats`

#### `/routes` — Route Explorer

- Table of all configured routes
- Columns: Path, Upstream Group, LB Strategy, Rate Limit, Status
- Click a route to expand and see per-upstream health
- Source: `GET /admin/api/routes`

#### `/upstreams` — Upstream Health

- Card per upstream group
- Per-upstream: URL, current health (green/red dot), active connections, response time (last check)
- Source: `GET /admin/api/upstreams`

#### `/circuit-breakers` — Circuit Breaker State

- Table: Upstream URL, Current State (badge), Failure Count, Last Trip Time, Next Retry In
- Manual "Force Close" button per breaker for testing
- Source: `GET /admin/api/circuit-breakers`, `POST /admin/api/circuit-breakers/:upstream/reset`

### Admin API Routes (served on port 9090)

```
GET  /admin/api/stats                          # Aggregated gateway stats
GET  /admin/api/routes                         # All configured routes
GET  /admin/api/upstreams                      # All upstream groups + health
GET  /admin/api/circuit-breakers               # All circuit breaker states
POST /admin/api/circuit-breakers/:id/reset     # Force reset a breaker
POST /admin/api/config/reload                  # Trigger hot config reload
GET  /admin/health                             # Admin server liveness check
```

> `/health` and `/metrics` are **not** on the admin port. They live on the main gateway port (8080) so external load balancers and Prometheus can reach them without touching the admin surface.

---

## 10. Gateway HTTP Routes (public-facing, port 8080)

```
ANY  /api/v1/users/*        → user-service (round-robin: 8081, 8082)
ANY  /api/v1/orders/*       → order-service (least-conn: 8083)
ANY  /api/v1/grpc/*         → grpc-service (round-robin: 9000)
GET  /health                → gateway health check
GET  /metrics               → Prometheus metrics
```

---

## 11. Infrastructure — Docker Compose

```yaml
networks:
    gateway-net:
        driver: bridge

services:
    gateway:
        build: .
        ports: ["8080:8080", "9090:9090"]
        depends_on:
            redis:
                condition: service_healthy
        networks: [gateway-net]
        restart: unless-stopped
        env_file: .env

    service-a:
        build: ./backends/service-a
        ports: ["8081:8081"]
        networks: [gateway-net]
        restart: unless-stopped

    service-b:
        build: ./backends/service-b
        ports: ["8082:8082"]
        networks: [gateway-net]
        restart: unless-stopped

    service-c:
        build: ./backends/service-c
        ports: ["9000:9000"]
        networks: [gateway-net]
        restart: unless-stopped

    service-d:
        build: ./backends/service-d
        ports: ["8083:8083"]
        networks: [gateway-net]
        restart: unless-stopped

    redis:
        image: redis:7-alpine
        ports: ["6379:6379"]
        networks: [gateway-net]
        restart: unless-stopped
        healthcheck:
            test: ["CMD", "redis-cli", "ping"]
            interval: 5s
            timeout: 3s
            retries: 5
        volumes:
            - redis-data:/data

    prometheus:
        image: prom/prometheus
        volumes:
            - ./deploy/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
        ports: ["9091:9090"]
        networks: [gateway-net]
        restart: unless-stopped

    grafana:
        image: grafana/grafana
        ports: ["3000:3000"]
        networks: [gateway-net]
        restart: unless-stopped
        volumes:
            - ./deploy/grafana/provisioning:/etc/grafana/provisioning
            - grafana-data:/var/lib/grafana

volumes:
    redis-data:
    grafana-data:
```

> The `grafana/provisioning` directory must contain **both** `datasources/prometheus.yaml` and `dashboards/gateway.json`. Without the datasource file, Grafana starts but the dashboard has no data source connected and all panels show "No data".

`deploy/grafana/provisioning/datasources/prometheus.yaml`:

```yaml
apiVersion: 1
datasources:
    - name: Prometheus
      type: prometheus
      url: http://prometheus:9090
      isDefault: true
```

---

## 12. Grafana Dashboard Panels

Pre-build a dashboard JSON in `deploy/grafana/dashboards/gateway.json` with these panels:

1. **Requests/sec** — `rate(gateway_requests_total[1m])`
2. **P99 Latency** — `histogram_quantile(0.99, rate(gateway_request_duration_seconds_bucket[1m]))`
3. **P95 Latency** — same with 0.95
4. **Error Rate %** — `rate(gateway_requests_total{status_code=~"5.."}[1m]) / rate(gateway_requests_total[1m]) * 100`
5. **Rate Limited Requests/sec** — `rate(gateway_rate_limited_total[1m])`
6. **Circuit Breaker States** — `gateway_circuit_breaker_state` by upstream
7. **Active Connections per Upstream** — `gateway_active_connections`
8. **Upstream Health** — `gateway_upstream_health` as a stat panel
9. **Redis Operation Latency** — `histogram_quantile(0.99, rate(gateway_redis_operation_duration_seconds_bucket[1m]))`
10. **Requests by Route** — stacked bar: `sum by(route)(rate(gateway_requests_total[1m]))`

---

## 13. Benchmarking Setup

### k6 Scripts (in `/benchmarks/k6/`)

**basic_load.js** — ramp to 1000 VUs, sustain for 3 minutes:

```js
export const options = {
    stages: [
        { duration: "30s", target: 200 },
        { duration: "2m", target: 1000 },
        { duration: "30s", target: 0 },
    ],
    thresholds: {
        http_req_duration: ["p(99)<15"], // P99 under 15ms
        http_req_failed: ["rate<0.01"], // Error rate under 1%
    },
};
```

**rate_limit_test.js** — fire from 1000+ unique IPs to test rate limiter at scale

**circuit_breaker_test.js** — kill service-b, verify 503s from gateway, restart it, verify recovery

### wrk for raw throughput:

```bash
wrk -t12 -c400 -d30s http://your-vps:8080/api/v1/users
```

> **Critical:** Run k6 and wrk from a **separate machine** (not the same VPS as the gateway). If the load generator and gateway compete for the same CPU cores, your benchmark numbers are meaningless. Use a second cheap VPS, your local machine, or a CI runner pointed at the VPS IP. The 10k–20k req/sec target assumes the gateway has full CPU access.

### Target numbers:

- Throughput: 10,000–20,000 req/sec
- P99 latency: < 15ms (without upstream slowness)
- Rate limit overhead: < 1ms per request (Redis RTT is the bottleneck)

---

## 14. Build Order (Step-by-Step)

Follow this sequence — each step is independently testable:

1. **Project scaffold** — `go mod init`, set up directory structure, write `Makefile`, create `.env.example`
2. **Config loader** — Parse `gateway.yaml` with Viper, validate required fields, support env var overrides via `viper.AutomaticEnv()`
3. **Dummy backends** — Write Service A, B, C, D. Docker Compose them up
4. **Basic HTTP proxy** — Hardcode one route, proxy to Service A, confirm it works with `curl`
5. **Router** — Dynamic route matching from config, multiple routes working
6. **Round-robin LB** — Distribute load across Service A and B; write unit tests for Next() logic
7. **Health checker** — Background polling, mark Service B unhealthy when killed
8. **Least-connections LB** — Add as alternate strategy, test with slow Service B; write unit tests
9. **Circuit breaker FSM** — Implement state machine; write unit tests for all 3 state transitions
10. **Prometheus metrics** — Instrument everything, verify `/metrics` output
11. **Redis rate limiting** — Add Lua script with EVALSHA, test with `curl` loops, verify 429s; unit test token math
12. **Graceful shutdown** — `signal.Notify` + `server.Shutdown(ctx)` with drain timeout
13. **gRPC proxy** — Wire up Service C, test with `grpcurl`
14. **Admin API** — REST endpoints for status/config
15. **Admin UI** — React app consuming admin API
16. **Grafana dashboard** — Import JSON dashboard with provisioned datasource, verify all panels populate
17. **Hot config reload** — `fsnotify` watcher, test live route changes
18. **VPS deployment** — Docker Compose on VPS, point k6 at it **from a separate machine**
19. **Benchmarks** — Run k6/wrk, capture screenshots, record numbers

---

## 15. Performance Considerations

- **Avoid allocations in hot path** — reuse buffers with `sync.Pool`
- **Use `atomic` not `sync.Mutex`** for counters (active connections, round-robin index)
- **Connection pooling** — tune `http.Transport` with `MaxIdleConnsPerHost`, `IdleConnTimeout`
- **Avoid logging in hot path** — use async log queue or sample logs at high RPS
- **Redis connection pool** — set `pool_size` to 50 in config; default of 10 is a bottleneck at 10k req/sec
- **Goroutine per request** — Go handles this natively; don't fight it
- **HTTP/2 for gRPC** — ensure TLS or `h2c` (HTTP/2 cleartext) is configured properly
- **GOMAXPROCS** — set explicitly to the number of VPS cores (`runtime.GOMAXPROCS(runtime.NumCPU())`), or use `uber-go/automaxprocs` which reads cgroup limits correctly in Docker
- **GC tuning** — at sustained high RPS, Go's GC can cause latency spikes. Set `GOGC=200` (less frequent GC) or use `GOMEMLIMIT` (Go 1.19+) to cap memory and let the runtime balance GC frequency itself: `GOMEMLIMIT=512MiB ./gateway`
- **Disable Nagle's algorithm** — set `TCP_NODELAY` on upstream connections to reduce latency on small responses

---

## 16. Things to Be Careful About

| Area                | Pitfall                                       | Solution                                                        |
| ------------------- | --------------------------------------------- | --------------------------------------------------------------- |
| Rate limiting       | Race condition on token bucket                | Use Redis Lua script (atomic)                                   |
| Circuit breaker     | Thundering herd on recovery                   | Half-open state only lets 1 probe through                       |
| Load balancer       | Routing to unhealthy upstreams                | Always filter by health registry before picking                 |
| gRPC proxy          | Framing errors if not using HTTP/2            | Ensure `h2c` transport for cleartext gRPC                       |
| Config reload       | Race condition during reload                  | Atomic pointer swap for route table                             |
| Metrics             | High cardinality labels (e.g., per-IP)        | Never label by raw client IP in Prometheus                      |
| Redis failure       | Gateway crashes if Redis is down              | Fail open on Redis errors (allow the request), log the error    |
| Panics              | Upstream returning unexpected data            | Recovery middleware on every request                            |
| Header forwarding   | Hop-by-hop headers leaking                    | Strip `Connection`, `Transfer-Encoding`, etc. before forwarding |
| Timeout propagation | Client disconnects but upstream still running | Use `context` cancellation on every upstream call               |

---

## 17. Makefile Targets

```makefile
run-gateway:      go run ./cmd/gateway
run-backends:     docker compose up service-a service-b service-c service-d
run-infra:        docker compose up redis prometheus grafana
run-all:          docker compose up
build:            go build -o bin/gateway ./cmd/gateway
test:             go test ./...
coverage:         go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
bench-basic:      k6 run benchmarks/k6/basic_load.js
bench-ratelimit:  k6 run benchmarks/k6/rate_limit_test.js
bench-cb:         k6 run benchmarks/k6/circuit_breaker_test.js
lint:             golangci-lint run
proto:            protoc --go_out=. --go-grpc_out=. backends/service-c/proto/echo.proto
docker-build:     docker build -t api-gateway .
clean:            rm -rf bin/ coverage.out
deploy:           rsync -avz --exclude='.git' --exclude='node_modules' --exclude='bin' \
                    --exclude='admin-ui/node_modules' . user@vps:/opt/api-gateway && \
                    ssh user@vps "cd /opt/api-gateway && docker compose up -d --build"
```

---

## 18. Go Dependencies (go.mod)

```
github.com/spf13/viper              # Config loading + env var override
github.com/redis/go-redis/v9        # Redis client
github.com/prometheus/client_golang # Prometheus metrics
github.com/fsnotify/fsnotify        # Config file watcher
go.uber.org/zap                     # Structured logging
google.golang.org/grpc              # gRPC
google.golang.org/protobuf          # Protobuf
github.com/mwitkow/grpc-proxy       # gRPC transparent proxy
golang.org/x/net                    # h2c (HTTP/2 cleartext) — required for gRPC without TLS
go.uber.org/automaxprocs            # Auto-sets GOMAXPROCS from cgroup limits in Docker
```

Create a `.env.example` (and real `.env` gitignored on VPS) for secrets:

```
REDIS_PASSWORD=
GATEWAY_ADMIN_TOKEN=changeme
```

Call `viper.AutomaticEnv()` and `viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))` so `redis.password` maps to `REDIS_PASSWORD`.

---

## 19. Resume Bullets (Fill in After Benchmarks)

Once you run k6 and capture real numbers, write these:

- Engineered a custom API gateway in Go processing **[X,XXX]+ requests/second** with P99 latency under **[X]ms**, supporting both HTTP/REST and gRPC upstream routing with dynamic load balancing
- Implemented Redis-backed token bucket rate limiting enforced across **1,000+ concurrent simulated clients** with sub-millisecond overhead per request using atomic Lua scripts
- Built a circuit breaker state machine (Closed/Open/Half-Open) with configurable failure thresholds, automatically isolating degraded upstreams and self-healing after recovery windows
- Deployed on a VPS with a Prometheus + Grafana observability stack tracking latency percentiles, error rates, circuit breaker states, and per-route throughput across all routes in real time

---

## 20. Stretch Goals (After Core is Working)

- **JWT authentication middleware** — validate Bearer tokens before proxying
- **Request/response transformation** — rewrite headers or body on the fly
- **Weighted load balancing** — give some upstreams more traffic share
- **TLS termination** — accept HTTPS, proxy HTTP to upstreams
- **Admin UI live updates** — WebSocket stream for real-time dashboard without polling
- **Distributed rate limiting** — multi-instance gateway sharing Redis rate limit state
- **A/B routing** — route X% of traffic to a canary upstream by header or percentage
