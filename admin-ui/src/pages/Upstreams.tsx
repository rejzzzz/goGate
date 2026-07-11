import { useEffect, useState } from 'react';
import { fetchUpstreams } from '../api/gateway';
import type { UpstreamGroup as UpstreamGroupType } from '../api/types';
import UpstreamStatus from '../components/UpstreamStatus';

export default function Upstreams() {
  const [groups, setGroups] = useState<UpstreamGroupType[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchUpstreams()
      .then(setGroups)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="loader">Discovering upstreams...</div>;

  return (
    <div>
      <div className="page-header">
        <h2>Upstream Health</h2>
        <p style={{ color: 'var(--text-secondary)', marginTop: '0.5rem' }}>
          Real-time status of all configured backend service groups.
        </p>
      </div>
      
      <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
        {groups.map((group, idx) => (
          <UpstreamStatus key={idx} group={group} />
        ))}
      </div>
    </div>
  );
}
