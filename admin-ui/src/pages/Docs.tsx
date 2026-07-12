import { BookOpen, Terminal, Network, ShieldCheck } from 'lucide-react';

export default function Docs() {
  return (
    <div style={{ maxWidth: '800px' }}>
      <div className="page-header" style={{ marginBottom: '3rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '2rem' }}>
        <h2 style={{ fontSize: '2rem', color: 'var(--text-primary)', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <BookOpen size={28} color="var(--primary-color)" />
          Gateway Documentation
        </h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1.125rem', lineHeight: '1.6' }}>
          Welcome to the official documentation for your Distributed API Gateway. Learn how to configure routing, manage upstreams, and ensure system resiliency.
        </p>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '4rem' }}>
        
        {/* Section 1 */}
        <section>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.5rem' }}>
            <div style={{ padding: '0.5rem', background: 'rgba(99, 102, 241, 0.1)', borderRadius: '8px' }}>
              <Terminal size={20} color="var(--primary-color)" />
            </div>
            <h3 style={{ margin: 0, fontSize: '1.5rem', color: 'var(--text-primary)', fontWeight: 600 }}>Configuration File</h3>
          </div>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1.5rem', fontSize: '1rem', lineHeight: '1.6' }}>
            The gateway is entirely config-driven via the <code>gateway.yaml</code> file located in the <code>configs/</code> directory.
            Whenever you make changes to this file, you can hot-reload the configuration directly from the Dashboard without dropping active connections.
          </p>
          <div style={{ background: 'var(--bg-color)', borderRadius: '8px', border: '1px solid var(--border-color)', overflow: 'hidden' }}>
            <div style={{ padding: '0.75rem 1rem', borderBottom: '1px solid var(--border-color)', background: 'var(--surface-color)', color: 'var(--text-secondary)', fontSize: '0.875rem', fontFamily: 'monospace' }}>
              gateway.yaml
            </div>
            <pre style={{ padding: '1.5rem', margin: 0, overflowX: 'auto', color: 'var(--text-primary)', fontSize: '0.875rem', lineHeight: '1.5' }}>
{`server:
  port: 8080
  admin_port: 9090
  
# Configure your routes below`}
            </pre>
          </div>
        </section>

        {/* Section 2 */}
        <section>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.5rem' }}>
            <div style={{ padding: '0.5rem', background: 'rgba(16, 185, 129, 0.1)', borderRadius: '8px' }}>
              <Network size={20} color="var(--success-color)" />
            </div>
            <h3 style={{ margin: 0, fontSize: '1.5rem', color: 'var(--text-primary)', fontWeight: 600 }}>Routing & Upstreams</h3>
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '1rem', lineHeight: '1.6', marginBottom: '1rem' }}>
            Routes are matched sequentially. You can strip prefixes before forwarding requests to an <code>upstream_group</code>. 
            An upstream group can contain multiple target URLs and specifies load balancing algorithms like <code>round-robin</code> or <code>least-connections</code>.
          </p>
          <ul style={{ listStyleType: 'disc', paddingLeft: '1.5rem', color: 'var(--text-secondary)', lineHeight: '1.8' }}>
            <li><strong>Path Matching:</strong> Supports exact and wildcard prefix matching.</li>
            <li><strong>Rate Limiting:</strong> Each route can enforce a strict requests-per-second (RPS) limit.</li>
            <li><strong>Load Balancing:</strong> Distributes traffic evenly among healthy nodes in an upstream group.</li>
          </ul>
        </section>

        {/* Section 3 */}
        <section>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.5rem' }}>
            <div style={{ padding: '0.5rem', background: 'rgba(245, 158, 11, 0.1)', borderRadius: '8px' }}>
              <ShieldCheck size={20} color="var(--warning-color)" />
            </div>
            <h3 style={{ margin: 0, fontSize: '1.5rem', color: 'var(--text-primary)', fontWeight: 600 }}>Resiliency (Circuit Breakers)</h3>
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '1rem', lineHeight: '1.6' }}>
            The gateway protects your microservices from cascading failures using Circuit Breakers.
            If a service starts failing rapidly, the circuit breaker opens and immediately rejects requests (returning HTTP 503) 
            until the timeout expires. You can monitor and manually reset tripped breakers in the <strong>Circuit Breakers</strong> tab.
          </p>
        </section>

      </div>
    </div>
  );
}
