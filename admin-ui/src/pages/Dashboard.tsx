import { useEffect, useState } from 'react';
import { fetchStats } from '../api/gateway';
import type { GatewayStats } from '../api/types';

export default function Dashboard() {
  const [stats, setStats] = useState<GatewayStats | null>(null);

  useEffect(() => {
    fetchStats().then(setStats).catch(console.error);
  }, []);

  if (!stats) return <div className="loader">Loading dashboard...</div>;

  return (
    <div>
      <div className="page-header">
        <h2>Gateway Overview</h2>
      </div>
      
      <div className="card-grid">
        <div className="card">
          <h3>Requests / Sec</h3>
          <div className="value">{stats.requestsPerSecond.toLocaleString()}</div>
        </div>
        <div className="card">
          <h3>Error Rate</h3>
          <div className="value">{(stats.errorRate * 100).toFixed(2)}%</div>
        </div>
        <div className="card">
          <h3>P99 Latency</h3>
          <div className="value">{stats.p99Latency}ms</div>
        </div>
        <div className="card">
          <h3>P95 Latency</h3>
          <div className="value">{stats.p95Latency}ms</div>
        </div>
        <div className="card">
          <h3>P50 Latency</h3>
          <div className="value">{stats.p50Latency}ms</div>
        </div>
        <div className="card">
          <h3>Rate Limited (Last min)</h3>
          <div className="value">{stats.rateLimitedCount.toLocaleString()}</div>
        </div>
        <div className="card">
          <h3>Open Circuit Breakers</h3>
          <div className="value">{stats.activeCircuitBreakers}</div>
        </div>
      </div>
    </div>
  );
}
