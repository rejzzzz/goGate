import { useEffect, useState } from 'react';
import { fetchCircuitBreakers, resetCircuitBreaker } from '../api/gateway';
import type { CircuitBreakerState } from '../api/types';
import CircuitBreakerBadge from '../components/CircuitBreakerBadge';

export default function CircuitBreakers() {
  const [breakers, setBreakers] = useState<CircuitBreakerState[]>([]);
  const [loading, setLoading] = useState(true);
  const [resetting, setResetting] = useState<string | null>(null);

  useEffect(() => {
    loadBreakers();
  }, []);

  const loadBreakers = () => {
    fetchCircuitBreakers()
      .then(setBreakers)
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  const handleReset = async (url: string) => {
    setResetting(url);
    await resetCircuitBreaker(url);
    await loadBreakers();
    setResetting(null);
  };

  if (loading) return <div className="loader">Analyzing circuit breaker states...</div>;

  return (
    <div>
      <div className="page-header">
        <h2>Circuit Breakers</h2>
      </div>
      
      <div className="card" style={{ padding: '0', overflow: 'hidden' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
          <thead>
            <tr>
              <th style={{ padding: '1.25rem 1.5rem', background: 'rgba(0,0,0,0.2)', color: 'var(--text-secondary)', fontWeight: 600, borderBottom: '1px solid var(--border-color)' }}>Upstream URL</th>
              <th style={{ padding: '1.25rem 1.5rem', background: 'rgba(0,0,0,0.2)', color: 'var(--text-secondary)', fontWeight: 600, borderBottom: '1px solid var(--border-color)' }}>State</th>
              <th style={{ padding: '1.25rem 1.5rem', background: 'rgba(0,0,0,0.2)', color: 'var(--text-secondary)', fontWeight: 600, borderBottom: '1px solid var(--border-color)' }}>Failures</th>
              <th style={{ padding: '1.25rem 1.5rem', background: 'rgba(0,0,0,0.2)', color: 'var(--text-secondary)', fontWeight: 600, borderBottom: '1px solid var(--border-color)' }}>Last Trip</th>
              <th style={{ padding: '1.25rem 1.5rem', background: 'rgba(0,0,0,0.2)', color: 'var(--text-secondary)', fontWeight: 600, borderBottom: '1px solid var(--border-color)' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {breakers.map((cb, idx) => (
              <tr key={idx} style={{ transition: 'background 0.2s' }}>
                <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-color)', fontWeight: 500, fontFamily: 'monospace' }}>
                  {cb.upstreamUrl.replace('http://', '')}
                </td>
                <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-color)' }}>
                  <CircuitBreakerBadge state={cb.state} />
                </td>
                <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-color)', color: cb.failureCount > 0 ? 'var(--warning-color)' : 'var(--text-secondary)' }}>
                  {cb.failureCount}
                </td>
                <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-color)', color: 'var(--text-secondary)' }}>
                  {cb.lastTripTime ? new Date(cb.lastTripTime).toLocaleTimeString() : '-'}
                </td>
                <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-color)' }}>
                  <button 
                    className="btn" 
                    onClick={() => handleReset(cb.upstreamUrl)}
                    disabled={cb.state === 'closed' || resetting === cb.upstreamUrl}
                    style={{ 
                      opacity: cb.state === 'closed' ? 0.3 : 1, 
                      cursor: cb.state === 'closed' ? 'not-allowed' : 'pointer',
                      border: '1px solid var(--border-color)',
                      background: 'rgba(255,255,255,0.05)'
                    }}
                  >
                    {resetting === cb.upstreamUrl ? 'Resetting...' : 'Force Close'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
