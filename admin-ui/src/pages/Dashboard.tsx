import { useEffect, useState } from 'react';
import { fetchStats } from '../api/gateway';
import type { GatewayStats } from '../api/types';
import MetricsChart from '../components/MetricsChart';

const generateChartData = () => {
  const data = [];
  const now = new Date();
  for (let i = 30; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 10000);
    data.push({
      time: time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      rps: Math.floor(11000 + Math.random() * 3000),
      latency: +(8 + Math.random() * 6).toFixed(1),
    });
  }
  return data;
};

export default function Dashboard() {
  const [stats, setStats] = useState<GatewayStats | null>(null);
  const [chartData, setChartData] = useState<any[]>([]);

  useEffect(() => {
    fetchStats().then(setStats).catch(console.error);
    setChartData(generateChartData());
    
    // Simulate real-time updates for chart
    const interval = setInterval(() => {
      setChartData(prev => {
        const newData = [...prev.slice(1)];
        const time = new Date();
        newData.push({
          time: time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
          rps: Math.floor(11000 + Math.random() * 3000),
          latency: +(8 + Math.random() * 6).toFixed(1),
        });
        return newData;
      });
    }, 10000);
    return () => clearInterval(interval);
  }, []);

  const [reloading, setReloading] = useState(false);

  const handleReload = async () => {
    setReloading(true);
    try {
      const { reloadConfig } = await import('../api/gateway');
      await reloadConfig();
      alert('Config reloaded successfully!');
    } catch (err) {
      alert('Failed to reload config');
    }
    setReloading(false);
  };

  if (!stats) return <div className="loader">Initializing Gateway Dashboard...</div>;

  return (
    <div>
      <div className="page-header flex-between">
        <h2>Gateway Overview</h2>
        <button 
          className="btn primary" 
          onClick={handleReload}
          disabled={reloading}
        >
          {reloading ? 'Reloading...' : 'Reload Config'}
        </button>
      </div>
      
      <div className="card-grid">
        <div className="card">
          <h3>Requests / Sec</h3>
          <div className="value" style={{ color: 'var(--primary-color)' }}>
            {stats.requestsPerSecond.toLocaleString()}
          </div>
        </div>
        <div className="card">
          <h3>Error Rate</h3>
          <div className="value" style={stats.errorRate > 0.01 ? { color: 'var(--danger-color)' } : { color: 'var(--success-color)' }}>
            {(stats.errorRate * 100).toFixed(2)}%
          </div>
        </div>
        <div className="card">
          <h3>P99 Latency</h3>
          <div className="value">{stats.p99Latency}ms</div>
        </div>
        <div className="card">
          <h3>Rate Limited</h3>
          <div className="value" style={{ color: 'var(--warning-color)' }}>
            {stats.rateLimitedCount.toLocaleString()}
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: '1.5rem' }}>
        <MetricsChart data={chartData} />
      </div>
    </div>
  );
}
