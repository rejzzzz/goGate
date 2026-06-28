import { useEffect, useState } from 'react';
import { fetchUpstreams } from '../api/gateway';
import type { UpstreamGroup } from '../api/types';

export default function Upstreams() {
  const [groups, setGroups] = useState<UpstreamGroup[]>([]);

  useEffect(() => {
    fetchUpstreams().then(setGroups).catch(console.error);
  }, []);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'healthy': return <span className="badge success">Healthy</span>;
      case 'degraded': return <span className="badge warning">Degraded</span>;
      case 'unhealthy': return <span className="badge danger">Unhealthy</span>;
      default: return <span className="badge neutral">{status}</span>;
    }
  };

  return (
    <div>
      <div className="page-header">
        <h2>Upstream Health</h2>
      </div>
      
      {groups.length === 0 ? <div className="loader">Loading upstreams...</div> : (
        groups.map(group => (
          <div key={group.name} style={{ marginBottom: '2rem' }}>
            <h3 style={{ marginBottom: '1rem' }}>Group: {group.name}</h3>
            <div className="table-container">
              <table>
                <thead>
                  <tr>
                    <th>URL</th>
                    <th>Status</th>
                    <th>Active Connections</th>
                    <th>Latency (ms)</th>
                  </tr>
                </thead>
                <tbody>
                  {group.upstreams.map(u => (
                    <tr key={u.url}>
                      <td>{u.url}</td>
                      <td>{getStatusBadge(u.status)}</td>
                      <td>{u.activeConnections}</td>
                      <td>{u.latencyMs}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))
      )}
    </div>
  );
}
