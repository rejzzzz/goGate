import { useEffect, useState } from 'react';
import { fetchStats } from '../api/gateway';
import type { GatewayStats } from '../api/types';
import MetricsChart from '../components/MetricsChart';

const generateChartData = (timeWindow: string) => {
  const data = [];
  const now = new Date();
  
  let points = 30;
  let intervalMs = 10000;
  
  switch(timeWindow) {
    case '15m':
      points = 30;
      intervalMs = 30000;
      break;
    case '30m':
      points = 30;
      intervalMs = 60000;
      break;
    case '1h':
      points = 60;
      intervalMs = 60000;
      break;
    case '24h':
      points = 48;
      intervalMs = 1800000; // 30 min intervals
      break;
    case '5m':
    default:
      points = 30;
      intervalMs = 10000;
  }

  for (let i = points; i >= 0; i--) {
    const time = new Date(now.getTime() - i * intervalMs);
    data.push({
      time: timeWindow === '24h' 
        ? time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) 
        : time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
      rps: Math.floor(11000 + Math.random() * 3000),
      latency: +(8 + Math.random() * 6).toFixed(1),
    });
  }
  return data;
};

export default function Dashboard() {
  const [stats, setStats] = useState<GatewayStats | null>(null);
  const [chartData, setChartData] = useState<any[]>([]);
  const [timeWindow, setTimeWindow] = useState('5m');

  useEffect(() => {
    fetchStats().then(setStats).catch((err) => {
      console.error('Failed to fetch stats, using mock data for preview:', err);
      setStats({
        requestsPerSecond: 12500,
        p50Latency: 5.2,
        p95Latency: 12.4,
        p99Latency: 18.7,
        errorRate: 0.001,
        rateLimitedCount: 45,
        activeCircuitBreakers: 0
      });
    });
  }, []);

  useEffect(() => {
    setChartData(generateChartData(timeWindow));
    
    let intervalMs = 10000;
    if (timeWindow === '15m') intervalMs = 30000;
    else if (timeWindow === '30m' || timeWindow === '1h') intervalMs = 60000;
    else if (timeWindow === '24h') intervalMs = 1800000;

    // Simulate real-time updates for chart
    const interval = setInterval(() => {
      setChartData(prev => {
        const newData = [...prev.slice(1)];
        const time = new Date();
        newData.push({
          time: timeWindow === '24h' 
            ? time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) 
            : time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
          rps: Math.floor(11000 + Math.random() * 3000),
          latency: +(8 + Math.random() * 6).toFixed(1),
        });
        return newData;
      });
    }, intervalMs);
    return () => clearInterval(interval);
  }, [timeWindow]);

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
        <MetricsChart 
          data={chartData} 
          timeWindow={timeWindow} 
          onTimeWindowChange={setTimeWindow} 
        />
      </div>
    </div>
  );
}
