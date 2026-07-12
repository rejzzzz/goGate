import type { RouteConfig } from '../api/types';
import { GitMerge, Activity, Server, ArrowRight } from 'lucide-react';

export default function RouteCard({ route }: { route: RouteConfig }) {
  return (
    <div className="card" style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
      <div className="flex-between">
        <h3 style={{ margin: 0, color: 'var(--text-primary)', fontSize: '1rem', fontWeight: 500 }}>{route.path}</h3>
        <span className="badge neutral">{route.lbStrategy}</span>
      </div>
      
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
        <GitMerge size={14} />
        <span>Strips Prefix: <strong style={{color: 'var(--text-primary)', fontWeight: 500}}>{route.stripPrefix ? 'Yes' : 'No'}</strong></span>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
        <Activity size={14} />
        <span>Rate Limit: <strong style={{color: 'var(--text-primary)', fontWeight: 500}}>{route.rateLimit.rps} req/s</strong> (Burst {route.rateLimit.burst})</span>
      </div>

      <div style={{ 
        marginTop: 'auto', 
        paddingTop: '1rem', 
        borderTop: '1px solid var(--border-color)',
        display: 'flex',
        alignItems: 'center',
        gap: '0.5rem',
        color: 'var(--text-secondary)',
        fontSize: '0.875rem'
      }}>
        <Server size={14} />
        <ArrowRight size={12} />
        <span style={{ fontWeight: 500, color: 'var(--text-primary)' }}>{route.upstreamGroup}</span>
      </div>
    </div>
  );
}
