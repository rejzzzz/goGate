import { render, screen } from '@testing-library/react';
import CircuitBreakerBadge from './CircuitBreakerBadge';

describe('CircuitBreakerBadge', () => {
  it('renders correctly in closed state', () => {
    render(<CircuitBreakerBadge state="closed" />);
    expect(screen.getByText(/CLOSED \(HEALTHY\)/i)).toBeDefined();
    expect(screen.getByText(/CLOSED/i).className).toContain('success');
  });

  it('renders correctly in half-open state', () => {
    render(<CircuitBreakerBadge state="half-open" />);
    expect(screen.getByText(/HALF-OPEN \(RECOVERY\)/i)).toBeDefined();
    expect(screen.getByText(/HALF-OPEN/i).className).toContain('warning');
  });

  it('renders correctly in open state', () => {
    render(<CircuitBreakerBadge state="open" />);
    expect(screen.getByText(/OPEN \(FAILING\)/i)).toBeDefined();
    expect(screen.getByText(/OPEN/i).className).toContain('danger');
  });
});
