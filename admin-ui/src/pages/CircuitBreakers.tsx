import { useEffect, useState } from 'react';
import { fetchCircuitBreakers, resetCircuitBreaker } from '../api/gateway';
import type { CircuitBreakerState } from '../api/types';

export default function CircuitBreakers() {
  const [breakers, setBreakers] = useState<CircuitBreakerState[]>([]);

  const load = () => {
    fetchCircuitBreakers().then(setBreakers).catch(console.error);
  };

  useEffect(() => {
    load();
  }, []);

  const handleReset = async (url: string) => {
    try {
      await resetCircuitBreaker(url);
      load(); // refresh after reset
    } catch (e) {
      console.error(e);
    }
  };

  const getStatusBadge = (state: string) => {
    switch (state) {
      case 'closed': return <span className="badge success">Closed</span>;
      case 'half-open': return <span className="badge warning">Half-Open</span>;
      case 'open': return <span className="badge danger">Open</span>;
      default: return <span className="badge neutral">{state}</span>;
    }
  };

  return (
    <div>
      <div className="page-header">
        <h2>Circuit Breakers</h2>
      </div>
      
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Upstream URL</th>
              <th>State</th>
              <th>Failures</th>
              <th>Last Trip Time</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {breakers.length === 0 ? (
              <tr><td colSpan={5} className="loader">Loading circuit breakers...</td></tr>
            ) : (
              breakers.map(b => (
                <tr key={b.upstreamUrl}>
                  <td><strong>{b.upstreamUrl}</strong></td>
                  <td>{getStatusBadge(b.state)}</td>
                  <td>{b.failureCount}</td>
                  <td>{b.lastTripTime ? new Date(b.lastTripTime).toLocaleString() : '-'}</td>
                  <td>
                    {b.state !== 'closed' && (
                      <button className="btn" onClick={() => handleReset(b.upstreamUrl)}>
                        Force Reset
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
