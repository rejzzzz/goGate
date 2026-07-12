import type { GatewayStats, RouteConfig, UpstreamGroup, CircuitBreakerState } from './types';

const API_BASE = '/admin/api';

export const fetchStats = async (): Promise<GatewayStats> => {
  const res = await fetch(`${API_BASE}/stats`);
  if (!res.ok) throw new Error('Failed to fetch stats');
  return res.json();
};

export const fetchRoutes = async (): Promise<RouteConfig[]> => {
  const res = await fetch(`${API_BASE}/routes`);
  if (!res.ok) throw new Error('Failed to fetch routes');
  return res.json();
};

export const fetchUpstreams = async (): Promise<UpstreamGroup[]> => {
  const res = await fetch(`${API_BASE}/upstreams`);
  if (!res.ok) throw new Error('Failed to fetch upstreams');
  return res.json();
};

export const fetchCircuitBreakers = async (): Promise<CircuitBreakerState[]> => {
  const res = await fetch(`${API_BASE}/circuit-breakers`);
  if (!res.ok) throw new Error('Failed to fetch circuit breakers');
  return res.json();
};

export const resetCircuitBreaker = async (upstreamUrl: string): Promise<void> => {
  const res = await fetch(`${API_BASE}/circuit-breakers/reset`, { 
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ upstreamUrl })
  });
  if (!res.ok) throw new Error('Failed to reset circuit breakers');
};

export const reloadConfig = async (): Promise<void> => {
  const res = await fetch(`${API_BASE}/config/reload`, { method: 'POST' });
  if (!res.ok) throw new Error('Failed to reload config');
};
