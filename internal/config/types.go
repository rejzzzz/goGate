package config

import "time"

// Config is the root configuration loaded from gateway.yaml.
type Config struct {
	Server                 ServerConfig    `mapstructure:"server"`
	Redis                  RedisConfig     `mapstructure:"redis"`
	Auth                   AuthConfig      `mapstructure:"auth"`
	Routes                 []Route         `mapstructure:"routes"`
	UpstreamGroups         []UpstreamGroup `mapstructure:"upstream_groups"`
	Metrics                MetricsConfig   `mapstructure:"metrics"`
	Logging                LoggingConfig   `mapstructure:"logging"`
	GlobalRateLimit        RateLimitConfig `mapstructure:"global_rate_limit"`
	RateLimitBypassHeader  string          `mapstructure:"rate_limit_bypass_header"`
	RateLimitBypassToken   string          `mapstructure:"rate_limit_bypass_token"`
}

// AuthConfig holds global API Key authentication settings.
type AuthConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	APIKeys []string `mapstructure:"api_keys"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	AdminPort       int           `mapstructure:"admin_port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// RedisConfig holds Redis connection pool settings.
type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	MaxRetries   int    `mapstructure:"max_retries"`
}

// Route defines a single routing rule.
type Route struct {
	Path          string          `mapstructure:"path"`
	StripPrefix   bool            `mapstructure:"strip_prefix"`
	UpstreamGroup string          `mapstructure:"upstream_group"`
	RateLimit     RateLimitConfig `mapstructure:"rate_limit"`
	LoadBalancer  string          `mapstructure:"load_balancer"`
	Type          string          `mapstructure:"type"` // "grpc" for gRPC routes, empty for HTTP
	AuthRequired  *bool           `mapstructure:"auth_required"`
}

// RateLimitConfig holds per-route rate limiting parameters.
type RateLimitConfig struct {
	RequestsPerSecond float64 `mapstructure:"requests_per_second"`
	Burst             int     `mapstructure:"burst"`
}

// UpstreamGroup is a named collection of upstream backends.
type UpstreamGroup struct {
	Name           string               `mapstructure:"name"`
	Upstreams      []Upstream           `mapstructure:"upstreams"`
	HealthCheck    HealthCheckConfig    `mapstructure:"health_check"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
	Type           string               `mapstructure:"type"` // "grpc" for gRPC upstream groups
	Discovery      *DiscoveryConfig     `mapstructure:"discovery"`
}

// DiscoveryConfig controls dynamic upstream discovery.
type DiscoveryConfig struct {
	Provider    string `mapstructure:"provider"`
	ServiceName string `mapstructure:"service_name"`
	Address     string `mapstructure:"address"`
}

// Upstream is a single backend URL.
// Health state is tracked separately in the healthcheck registry.
type Upstream struct {
	URL string `mapstructure:"url"`
}

// HealthCheckConfig controls background health polling.
type HealthCheckConfig struct {
	Path     string        `mapstructure:"path"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// CircuitBreakerConfig controls the circuit breaker state machine.
type CircuitBreakerConfig struct {
	FailureThreshold int           `mapstructure:"failure_threshold"`
	SuccessThreshold int           `mapstructure:"success_threshold"`
	Timeout          time.Duration `mapstructure:"timeout"`
}

// MetricsConfig controls Prometheus metrics exposure.
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// LoggingConfig controls structured log output.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}
