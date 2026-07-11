import { render, screen } from '@testing-library/react';
import UpstreamStatus from './UpstreamStatus';
import type { UpstreamGroup } from '../api/types';

const mockGroup: UpstreamGroup = {
  name: 'test-service',
  upstreams: [
    { url: 'http://localhost:8080', status: 'healthy', activeConnections: 10, latencyMs: 5 },
    { url: 'http://localhost:8081', status: 'degraded', activeConnections: 50, latencyMs: 150 },
  ],
};

describe('UpstreamStatus', () => {
  it('renders the group name and node count', () => {
    render(<UpstreamStatus group={mockGroup} />);
    expect(screen.getByText('test-service')).toBeDefined();
    expect(screen.getByText('2 Nodes')).toBeDefined();
  });

  it('renders healthy upstreams correctly', () => {
    render(<UpstreamStatus group={mockGroup} />);
    expect(screen.getByText('localhost:8080')).toBeDefined();
    const healthyBadge = screen.getByText('healthy');
    expect(healthyBadge).toBeDefined();
    expect(healthyBadge.className).toContain('success');
  });

  it('renders degraded upstreams correctly', () => {
    render(<UpstreamStatus group={mockGroup} />);
    expect(screen.getByText('localhost:8081')).toBeDefined();
    const degradedBadge = screen.getByText('degraded');
    expect(degradedBadge).toBeDefined();
    expect(degradedBadge.className).toContain('warning');
  });
});
