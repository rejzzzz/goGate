export interface GatewayStats {
  requestsPerSecond: number;
  p50Latency: number;
  p95Latency: number;
  p99Latency: number;
  errorRate: number;
  rateLimitedCount: number;
  activeCircuitBreakers: number;
}

export interface RouteConfig {
  path: string;
  upstreamGroup: string;
  lbStrategy: string;
  rateLimit: {
    rps: number;
    burst: number;
  };
  stripPrefix: boolean;
}

export interface UpstreamHealth {
  url: string;
  status: 'healthy' | 'degraded' | 'unhealthy';
  activeConnections: number;
  latencyMs: number;
}

export interface UpstreamGroup {
  name: string;
  upstreams: UpstreamHealth[];
}

export interface CircuitBreakerState {
  upstreamUrl: string;
  state: 'closed' | 'open' | 'half-open';
  failureCount: number;
  lastTripTime?: string;
}
