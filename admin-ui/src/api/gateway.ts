import type { GatewayStats, RouteConfig, UpstreamGroup, CircuitBreakerState } from './types';

// Simple mocked client returning promises to simulate network delay
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

export const fetchStats = async (): Promise<GatewayStats> => {
  await delay(300);
  return {
    requestsPerSecond: 12450,
    p50Latency: 4.2,
    p95Latency: 11.5,
    p99Latency: 14.8,
    errorRate: 0.05,
    rateLimitedCount: 120,
    activeCircuitBreakers: 0,
  };
};

export const fetchRoutes = async (): Promise<RouteConfig[]> => {
  await delay(300);
  return [
    {
      path: '/api/v1/users',
      upstreamGroup: 'user-service',
      lbStrategy: 'round-robin',
      rateLimit: { rps: 100, burst: 20 },
      stripPrefix: true,
    },
    {
      path: '/api/v1/orders',
      upstreamGroup: 'order-service',
      lbStrategy: 'least-connections',
      rateLimit: { rps: 50, burst: 10 },
      stripPrefix: true,
    },
    {
      path: '/api/v1/grpc',
      upstreamGroup: 'grpc-service',
      lbStrategy: 'round-robin',
      rateLimit: { rps: 500, burst: 100 },
      stripPrefix: false,
    },
  ];
};

export const fetchUpstreams = async (): Promise<UpstreamGroup[]> => {
  await delay(300);
  return [
    {
      name: 'user-service',
      upstreams: [
        { url: 'http://localhost:8081', status: 'healthy', activeConnections: 124, latencyMs: 3 },
        { url: 'http://localhost:8082', status: 'degraded', activeConnections: 45, latencyMs: 56 },
      ],
    },
    {
      name: 'order-service',
      upstreams: [
        { url: 'http://localhost:8083', status: 'healthy', activeConnections: 89, latencyMs: 4 },
        { url: 'http://localhost:8084', status: 'healthy', activeConnections: 91, latencyMs: 5 },
      ],
    },
  ];
};

export const fetchCircuitBreakers = async (): Promise<CircuitBreakerState[]> => {
  await delay(300);
  return [
    { upstreamUrl: 'http://localhost:8081', state: 'closed', failureCount: 0 },
    { upstreamUrl: 'http://localhost:8082', state: 'half-open', failureCount: 4, lastTripTime: new Date(Date.now() - 15000).toISOString() },
  ];
};

export const resetCircuitBreaker = async (upstreamUrl: string): Promise<void> => {
  await delay(500);
  console.log(`Reset circuit breaker for ${upstreamUrl}`);
};
