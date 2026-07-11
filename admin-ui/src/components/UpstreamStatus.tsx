import type { UpstreamGroup } from '../api/types';
import { Server, Activity } from 'lucide-react';

export default function UpstreamStatus({ group }: { group: UpstreamGroup }) {
  return (
    <div className="card" style={{ gridColumn: '1 / -1' }}>
      <div className="flex-between" style={{ marginBottom: '1.5rem' }}>
        <h3 style={{ margin: 0, fontSize: '1.5rem', color: 'var(--text-primary)', textTransform: 'none' }}>
          {group.name}
        </h3>
        <span className="badge neutral" style={{ background: 'rgba(255,255,255,0.05)' }}>
          {group.upstreams.length} Nodes
        </span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1rem' }}>
        {group.upstreams.map((u, i) => (
          <div key={i} style={{ 
            background: 'rgba(0,0,0,0.2)', 
            padding: '1.25rem', 
            borderRadius: '8px',
            border: '1px solid var(--border-color)',
            transition: 'all 0.2s ease',
            cursor: 'default'
          }}
          onMouseEnter={(e) => e.currentTarget.style.borderColor = 'var(--primary-glow)'}
          onMouseLeave={(e) => e.currentTarget.style.borderColor = 'var(--border-color)'}
          >
            <div className="flex-between" style={{ marginBottom: '1rem' }}>
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <span className={`status-dot ${u.status === 'healthy' ? 'healthy' : u.status === 'degraded' ? 'degraded' : 'down'}`}></span>
                <span style={{ fontWeight: 600, fontFamily: 'monospace', color: '#fff' }}>{u.url.replace('http://', '')}</span>
              </div>
              <span className={`badge ${u.status === 'healthy' ? 'success' : u.status === 'degraded' ? 'warning' : 'danger'}`}>
                {u.status}
              </span>
            </div>
            
            <div className="flex-between" style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
              <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <Activity size={14} /> {u.latencyMs}ms
              </span>
              <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <Server size={14} /> {u.activeConnections} conns
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
