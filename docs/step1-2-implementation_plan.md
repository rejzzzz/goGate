Here's the detailed implementation plan for steps 1 (Config Layer) and 2 (Backend Services), broken into small, testable increments.

---

## Step 1: Config Layer

### 1.1 — Define `config/types.go`

Fill in all structs to match `gateway.yaml` exactly:

- `Config` — root struct with fields: `Server`, `Redis`, `Routes`, `UpstreamGroups`, `Metrics`, `Logging`
- `ServerConfig` — `Port`, `AdminPort`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ShutdownTimeout` (all `time.Duration` except ports)
- `RedisConfig` — `Addr`, `Password`, `DB`, `PoolSize`, `MinIdleConns`, `MaxRetries`
- `Route` — `Path`, `StripPrefix`, `UpstreamGroup`, `RateLimit`, `LoadBalancer`, `Type` (for grpc)
- `RateLimitConfig` — `RequestsPerSecond`, `Burst`
- `UpstreamGroup` — `Name`, `Upstreams`, `HealthCheck`, `CircuitBreaker`, `Type`
- `Upstream` — `URL` only (health state lives in the healthcheck registry, not here)
- `HealthCheckConfig` — `Path`, `Interval`, `Timeout`
- `CircuitBreakerConfig` — `FailureThreshold`, `SuccessThreshold`, `Timeout`
- `MetricsConfig` — `Enabled`, `Path`
- `LoggingConfig` — `Level`, `Format`

Use `mapstructure` tags matching the yaml keys (Viper uses mapstructure under the hood). Keep structs flat — no embedding, no pointer receivers on data types.

---

### 1.2 — Implement `config/config.go` — `Load()`

- Use `viper.SetConfigFile(path)` to point at the specific file
- Call `viper.AutomaticEnv()` and `viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))` — maps `redis.password` → `REDIS_PASSWORD`
- Call `viper.ReadInConfig()`, return error on failure
- Unmarshal into `*Config` via `viper.Unmarshal(&cfg)`
- Call `Validate(cfg)` before returning — fail fast on bad config
- Return `*Config, error` — no global state, caller owns the value

---

### 1.3 — Implement `config/config.go` — `Validate()`

Keep validation rules small and explicit:

- `server.port` must be 1–65535
- `server.admin_port` must be 1–65535 and different from `server.port`
- `server.shutdown_timeout` must be > 0
- Each `route.path` must start with `/`
- Each `route.upstream_group` must reference a name that exists in `upstream_groups`
- Each `route.load_balancer` must be `"round-robin"` or `"least-connections"`
- Each `upstream.url` must be a valid URL (use `url.Parse` + check scheme)
- `redis.pool_size` must be > 0
- Rate limit values must be positive if set

Return a descriptive `error` for each violation — wrap them with `fmt.Errorf("route %q: %w", path, err)` so the caller knows which route failed.

---

### 1.4 — Implement `config/config.go` — `WatchForChanges()`

- Use `viper.WatchConfig()` + `viper.OnConfigChange()` — fsnotify is already in go.mod
- On change: call `Load()` again, call `Validate()`, only invoke the callback if validation passes
- Log a warning (don't crash) if the new config is invalid — keep running with the old config
- The callback signature: `func(*Config)` — caller decides what to do (hot-reload)

---

### 1.5 — Verify config layer

Manual check (no test framework needed yet):

- Write a quick `TestLoad` in `config/config_test.go` that loads `gateway.yaml` and asserts a few fields have expected values
- Write `TestValidate_MissingUpstreamGroup` that creates a `Config` with a route referencing a non-existent group and asserts an error is returned
- Run `go test ./internal/config/...` — must pass

---

## Step 2: Backend Services

### 2.1 — Complete `examples/backends/service-a/main.go`

Fill in the two empty handlers:

**`handleUsers`:**

- Set `Content-Type: application/json`
- Write a hardcoded JSON array: `[{"id":"1","name":"Alice"},{"id":"2","name":"Bob"}]`
- Return 200

**`handleUserByID`:**

- Extract the ID from the URL path by trimming `/users/` prefix via `strings.TrimPrefix`
- If ID is `"1"` return `{"id":"1","name":"Alice"}`, if `"2"` return `{"id":"2","name":"Bob"}`
- If unknown ID return 404 with `{"error":"not found"}`

No external dependencies, no routing library. The existing `http.HandleFunc` registrations are fine.

---

### 2.2 — Complete `examples/backends/service-b/main.go`

Service-b is a drop-in slow/flaky version of service-a. The env vars (`DELAY_MS`, `ERROR_RATE`) are already parsed. Wire them into handlers:

- Create a handler factory `func makeHandlers(delayMs int, errorRate float64)` that returns two `http.HandlerFunc` values — one for `/users`, one for `/users/`
- Each handler: call `addDelay(delayMs)`, then `shouldReturnError(errorRate)` — if true return 500 with `{"error":"internal error"}`, otherwise return same JSON as service-a
- Register these handlers in `main()` using the parsed env vars
- `/health` stays fast — no delay, no random errors (health check must not trip the circuit breaker)

---

### 2.3 — Create `examples/backends/service-c/main.go` (gRPC echo service)

This file is completely missing. Create it:

- Run `protoc` to generate Go code from `echo.proto` into `examples/backends/service-c/proto/` — or generate it manually since it's simple enough to write by hand (`echo_grpc.pb.go`, `echo.pb.go`)
- Implement `EchoServiceServer` interface with a single `Echo` method that returns `EchoResponse{Message: req.Message, Timestamp: time.Now().UnixMilli()}`
- In `main()`: read `PORT` env var (default `9000`), create `grpc.NewServer()`, register the service, call `server.Serve(lis)`
- No TLS needed — plain gRPC for local/Docker use

---

### 2.4 — Create `examples/backends/service-d/main.go` (order service)

Also completely missing. Create it as a minimal orders REST service:

- Identical structure to service-a but serves `/orders` and `/orders/` instead of `/users`
- Return hardcoded order JSON: `[{"id":"1","item":"Widget","qty":5},{"id":"2","item":"Gadget","qty":2}]`
- `/health` returns `{"status":"ok"}`
- Read `PORT` from env, default `8083`
- Optional `DELAY_MS` env var for simulating a slow instance (useful for least-connections testing later)

---

### 2.5 — Verify all backends compile and respond

Run each service locally and smoke-test with curl:

```
# Terminal 1
cd examples/backends/service-a && go run main.go
curl http://localhost:8081/health          # {"status":"ok"}
curl http://localhost:8081/users           # [...users...]
curl http://localhost:8081/users/1         # {"id":"1",...}
curl http://localhost:8081/users/99        # 404

# Terminal 2
cd examples/backends/service-b && go run main.go
curl http://localhost:8082/users           # sometimes 500, ~95% returns data

# Terminal 3
cd examples/backends/service-c && go run main.go
# grpcurl -plaintext localhost:9000 echo.EchoService/Echo  (if grpcurl installed)

# Terminal 4
cd examples/backends/service-d && go run main.go
curl http://localhost:8083/health
curl http://localhost:8083/orders
```

---

### 2.6 — Verify Docker builds

Each backend has a `Dockerfile` already. Build and confirm they compile in Docker:

```
docker build -t service-a ./examples/backends/service-a
docker build -t service-b ./examples/backends/service-b
docker build -t service-c ./examples/backends/service-c
docker build -t service-d ./examples/backends/service-d
```

Fix any build errors (likely just missing generated proto files for service-c).

---

### 2.7 — Bring up backends via Docker Compose

```
docker compose up service-a service-b service-d redis
```

Confirm all containers start healthy and respond on their ports. Service-c can be skipped until the gRPC proxy step.

---

## Dependency Graph

```
1.1 types.go
  └─ 1.2 Load()
       └─ 1.3 Validate()
            └─ 1.4 WatchForChanges()
                 └─ 1.5 config tests ✓

2.1 service-a ──┐
2.2 service-b ──┤
2.3 service-c ──┼── 2.5 smoke test ── 2.6 Docker build ── 2.7 compose up ✓
2.4 service-d ──┘
```

Steps 1 and 2 are independent of each other and can be done in either order, but config needs to be done before `main.go` and the proxy layer (step 3+).

---

Ready to start implementing? I'd suggest kicking off with `config/types.go` since every other file depends on it.
