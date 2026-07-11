import { useEffect, useState } from 'react';
import { fetchRoutes } from '../api/gateway';
import type { RouteConfig } from '../api/types';
import RouteCard from '../components/RouteCard';

export default function Routes() {
  const [routes, setRoutes] = useState<RouteConfig[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchRoutes()
      .then(setRoutes)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="loader">Loading route configurations...</div>;

  return (
    <div>
      <div className="page-header flex-between">
        <h2>Configured Routes</h2>
        <span className="badge neutral" style={{ background: 'var(--surface-color)' }}>
          {routes.length} Active Routes
        </span>
      </div>
      
      <div className="card-grid">
        {routes.map((route, idx) => (
          <RouteCard key={idx} route={route} />
        ))}
      </div>
    </div>
  );
}
