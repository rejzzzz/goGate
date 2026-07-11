import type { RouteConfig } from '../api/types';
import { GitMerge, Activity, Server, ArrowRight } from 'lucide-react';

export default function RouteCard({ route }: { route: RouteConfig }) {
  return (
    <div className="card" style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
      <div className="flex-between">
        <h3 style={{ margin: 0, color: '#fff', fontSize: '1.25rem' }}>{route.path}</h3>
        <span className="badge neutral">{route.lbStrategy}</span>
      </div>
      
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
        <GitMerge size={16} />
        <span>Strips Prefix: <strong style={{color: '#fff'}}>{route.stripPrefix ? 'Yes' : 'No'}</strong></span>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
        <Activity size={16} />
        <span>Rate Limit: <strong style={{color: '#fff'}}>{route.rateLimit.rps} req/s</strong> (Burst {route.rateLimit.burst})</span>
      </div>

      <div style={{ 
        marginTop: 'auto', 
        paddingTop: '1.5rem', 
        borderTop: '1px solid var(--border-color)',
        display: 'flex',
        alignItems: 'center',
        gap: '0.5rem',
        color: 'var(--primary-color)'
      }}>
        <Server size={16} />
        <ArrowRight size={14} />
        <span style={{ fontWeight: 600 }}>{route.upstreamGroup}</span>
      </div>
    </div>
  );
}
