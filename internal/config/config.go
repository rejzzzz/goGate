package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Load parses and validates the configuration file at path.
// Environment variables override file values: redis.password → REDIS_PASSWORD.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, func(config *mapstructure.DecoderConfig) {
		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
		)
	}); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := Validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate checks all configuration constraints.
// Returns a descriptive error identifying the offending field or route.
func Validate(cfg *Config) error {
	if err := validateServer(cfg.Server); err != nil {
		return err
	}
	if err := validateRedis(cfg.Redis); err != nil {
		return err
	}

	// Build a set of upstream group names for cross-reference checks.
	groupNames := make(map[string]struct{}, len(cfg.UpstreamGroups))
	for _, g := range cfg.UpstreamGroups {
		groupNames[g.Name] = struct{}{}
	}

	for _, r := range cfg.Routes {
		if err := validateRoute(r, groupNames); err != nil {
			return err
		}
	}

	for _, g := range cfg.UpstreamGroups {
		if err := validateUpstreamGroup(g); err != nil {
			return err
		}
	}

	return nil
}

// WatchForChanges monitors path for edits and calls onChange with a freshly
// loaded Config whenever the file changes and the new config is valid.
// Invalid configs are silently ignored so the gateway keeps running.
func WatchForChanges(path string, onChange func(*Config)) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Ignore the initial read error here — caller already loaded successfully.
	_ = v.ReadInConfig()

	v.WatchConfig()
	v.OnConfigChange(func(_ fsnotify.Event) {
		newCfg, err := Load(path)
		if err != nil {
			// Keep running with the current config; caller is not notified.
			return
		}
		onChange(newCfg)
	})
}

// --- private helpers ---------------------------------------------------------

func validateServer(s ServerConfig) error {
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("server.port %d is out of range 1-65535", s.Port)
	}
	if s.AdminPort < 1 || s.AdminPort > 65535 {
		return fmt.Errorf("server.admin_port %d is out of range 1-65535", s.AdminPort)
	}
	if s.AdminPort == s.Port {
		return fmt.Errorf("server.admin_port must differ from server.port (%d)", s.Port)
	}
	if s.ShutdownTimeout <= 0 {
		return fmt.Errorf("server.shutdown_timeout must be greater than zero")
	}
	return nil
}

func validateRedis(r RedisConfig) error {
	if r.PoolSize <= 0 {
		return fmt.Errorf("redis.pool_size must be greater than zero")
	}
	return nil
}

func validateRoute(r Route, groupNames map[string]struct{}) error {
	if !strings.HasPrefix(r.Path, "/") {
		return fmt.Errorf("route %q: path must start with '/'", r.Path)
	}

	if r.Async != nil {
		if r.Async.Exchange == "" || r.Async.RoutingKey == "" {
			return fmt.Errorf("route %q: async config must specify exchange and routing_key", r.Path)
		}
	} else {
		if _, ok := groupNames[r.UpstreamGroup]; !ok {
			return fmt.Errorf("route %q: upstream_group %q not found in upstream_groups", r.Path, r.UpstreamGroup)
		}

		// gRPC routes do not use a load-balancer field in the same way; skip LB
		// validation for them to avoid false negatives when the field is empty.
		if r.Type != "grpc" {
			valid := map[string]bool{"round-robin": true, "least-connections": true}
			if !valid[r.LoadBalancer] {
				return fmt.Errorf("route %q: load_balancer %q must be \"round-robin\" or \"least-connections\"", r.Path, r.LoadBalancer)
			}
		}
	}

	if r.RateLimit.RequestsPerSecond < 0 {
		return fmt.Errorf("route %q: rate_limit.requests_per_second must be non-negative", r.Path)
	}
	if r.RateLimit.Burst < 0 {
		return fmt.Errorf("route %q: rate_limit.burst must be non-negative", r.Path)
	}

	return nil
}

func validateUpstreamGroup(g UpstreamGroup) error {
	if g.Name == "" {
		return fmt.Errorf("upstream group is missing a name")
	}
	if (g.Upstreams == nil || len(g.Upstreams) == 0) && g.Discovery == nil {
		return fmt.Errorf("upstream group %s has no upstreams and no discovery config", g.Name)
	}
	for _, u := range g.Upstreams {
		parsed, err := url.Parse(u.URL)
		if err != nil {
			return fmt.Errorf("upstream_group %q: invalid url %q: %w", g.Name, u.URL, err)
		}
		// gRPC upstream URLs are host:port without a scheme.
		// url.Parse is lenient: "localhost:8081" parses with scheme "localhost",
		// so we check for a recognised transport scheme explicitly.
		if g.Type != "grpc" {
			if parsed.Scheme != "http" && parsed.Scheme != "https" {
				return fmt.Errorf("upstream_group %q: url %q must use http:// or https://", g.Name, u.URL)
			}
		}
	}
	return nil
}
