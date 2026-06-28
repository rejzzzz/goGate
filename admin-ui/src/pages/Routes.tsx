import { useEffect, useState } from 'react';
import { fetchRoutes } from '../api/gateway';
import type { RouteConfig } from '../api/types';

export default function GatewayRoutes() {
  const [routes, setRoutes] = useState<RouteConfig[]>([]);

  useEffect(() => {
    fetchRoutes().then(setRoutes).catch(console.error);
  }, []);

  return (
    <div>
      <div className="page-header">
        <h2>Active Routes</h2>
      </div>
      
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Path Prefix</th>
              <th>Upstream Group</th>
              <th>Load Balancer</th>
              <th>Rate Limit (RPS)</th>
              <th>Strip Prefix</th>
            </tr>
          </thead>
          <tbody>
            {routes.length === 0 ? (
              <tr><td colSpan={5} className="loader">Loading routes...</td></tr>
            ) : (
              routes.map(r => (
                <tr key={r.path}>
                  <td><strong>{r.path}</strong></td>
                  <td>{r.upstreamGroup}</td>
                  <td>{r.lbStrategy}</td>
                  <td>{r.rateLimit.rps} (Burst: {r.rateLimit.burst})</td>
                  <td>{r.stripPrefix ? 'Yes' : 'No'}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
