package config

import (
	"testing"
	"time"
)

func TestLoad_GatewayYAML(t *testing.T) {
	cfg, err := Load("../../configs/gateway.yaml")
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("server.port: got %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.AdminPort != 9090 {
		t.Errorf("server.admin_port: got %d, want 9090", cfg.Server.AdminPort)
	}
	if cfg.Server.ShutdownTimeout != 30*time.Second {
		t.Errorf("server.shutdown_timeout: got %v, want 30s", cfg.Server.ShutdownTimeout)
	}
	if cfg.Redis.PoolSize != 50 {
		t.Errorf("redis.pool_size: got %d, want 50", cfg.Redis.PoolSize)
	}
	if len(cfg.Routes) != 4 {
		t.Errorf("routes: got %d, want 4", len(cfg.Routes))
	}
	if len(cfg.UpstreamGroups) != 3 {
		t.Errorf("upstream_groups: got %d, want 3", len(cfg.UpstreamGroups))
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Fatal("Load() expected an error for missing file, got nil")
	}
}

func TestValidate_MissingUpstreamGroup(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080, AdminPort: 9090, ShutdownTimeout: 30 * time.Second},
		Redis:  RedisConfig{PoolSize: 10},
		Routes: []Route{
			{Path: "/api/v1/test", UpstreamGroup: "does-not-exist", LoadBalancer: "round-robin"},
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate() expected error for unknown upstream_group, got nil")
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 0, AdminPort: 9090, ShutdownTimeout: 30 * time.Second},
		Redis:  RedisConfig{PoolSize: 10},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate() expected error for port 0, got nil")
	}
}

func TestValidate_SamePort(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080, AdminPort: 8080, ShutdownTimeout: 30 * time.Second},
		Redis:  RedisConfig{PoolSize: 10},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate() expected error when port == admin_port, got nil")
	}
}

func TestValidate_InvalidLoadBalancer(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080, AdminPort: 9090, ShutdownTimeout: 30 * time.Second},
		Redis:  RedisConfig{PoolSize: 10},
		Routes: []Route{
			{Path: "/api/v1/test", UpstreamGroup: "svc", LoadBalancer: "random"},
		},
		UpstreamGroups: []UpstreamGroup{{Name: "svc"}},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate() expected error for unknown load_balancer, got nil")
	}
}

func TestValidate_RoutePathMissingSlash(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080, AdminPort: 9090, ShutdownTimeout: 30 * time.Second},
		Redis:  RedisConfig{PoolSize: 10},
		Routes: []Route{
			{Path: "api/v1/test", UpstreamGroup: "svc", LoadBalancer: "round-robin"},
		},
		UpstreamGroups: []UpstreamGroup{{Name: "svc"}},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate() expected error for path missing leading '/', got nil")
	}
}

func TestValidate_UpstreamMissingScheme(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080, AdminPort: 9090, ShutdownTimeout: 30 * time.Second},
		Redis:  RedisConfig{PoolSize: 10},
		UpstreamGroups: []UpstreamGroup{
			{
				Name:      "svc",
				Upstreams: []Upstream{{URL: "localhost:8081"}}, // missing http://
			},
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate() expected error for upstream URL without scheme, got nil")
	}
}
