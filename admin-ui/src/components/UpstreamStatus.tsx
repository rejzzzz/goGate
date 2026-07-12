import type { UpstreamGroup } from '../api/types';
import { Server, Activity } from 'lucide-react';

export default function UpstreamStatus({ group }: { group: UpstreamGroup }) {
  return (
    <div className="card" style={{ gridColumn: '1 / -1', padding: '1.5rem' }}>
      <div className="flex-between" style={{ marginBottom: '1.5rem' }}>
        <h3 style={{ margin: 0, fontSize: '1.25rem', color: 'var(--text-primary)' }}>
          {group.name}
        </h3>
        <span className="badge neutral">
          {group.upstreams.length} Nodes
        </span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1rem' }}>
        {group.upstreams.map((u, i) => (
          <div key={i} style={{ 
            background: 'var(--bg-color)', 
            padding: '1rem', 
            borderRadius: 'var(--border-radius)',
            border: '1px solid var(--border-color)',
            transition: 'border-color 0.15s ease'
          }}>
            <div className="flex-between" style={{ marginBottom: '0.75rem' }}>
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <span className={`status-dot ${u.status === 'healthy' ? 'healthy' : u.status === 'degraded' ? 'degraded' : 'down'}`}></span>
                <span style={{ fontWeight: 500, fontSize: '0.875rem', fontFamily: 'monospace', color: 'var(--text-primary)' }}>{u.url.replace('http://', '')}</span>
              </div>
              <span className={`badge ${u.status === 'healthy' ? 'success' : u.status === 'degraded' ? 'warning' : 'danger'}`}>
                {u.status}
              </span>
            </div>
            
            <div className="flex-between" style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
              <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                <Activity size={12} /> {u.latencyMs}ms
              </span>
              <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                <Server size={12} /> {u.activeConnections} conns
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
