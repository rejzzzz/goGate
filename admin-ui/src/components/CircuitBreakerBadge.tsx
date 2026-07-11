import { ShieldAlert, ShieldCheck, Shield } from 'lucide-react';

export default function CircuitBreakerBadge({ state }: { state: 'closed' | 'open' | 'half-open' }) {
  if (state === 'closed') {
    return (
      <span className="badge success" style={{ display: 'inline-flex', gap: '6px', padding: '0.5rem 1rem', fontSize: '0.8rem' }}>
        <ShieldCheck size={16} /> CLOSED (HEALTHY)
      </span>
    );
  }
  
  if (state === 'half-open') {
    return (
      <span className="badge warning" style={{ display: 'inline-flex', gap: '6px', padding: '0.5rem 1rem', fontSize: '0.8rem' }}>
        <Shield size={16} /> HALF-OPEN (RECOVERY)
      </span>
    );
  }

  return (
    <span className="badge danger" style={{ 
      display: 'inline-flex', 
      gap: '6px', 
      padding: '0.5rem 1rem', 
      fontSize: '0.8rem',
      animation: 'pulse-glow 2s infinite' 
    }}>
      <ShieldAlert size={16} /> OPEN (FAILING)
    </span>
  );
}
