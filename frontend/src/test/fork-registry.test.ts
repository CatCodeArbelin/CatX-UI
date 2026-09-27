import { describe, expect, it } from 'vitest';

import { forkApiSections, forkNavigationItems, forkRoutes } from '@/forkext/registry';

describe('fork registries', () => {
  it('registers the client activity page through the fork boundary', () => {
    expect(forkRoutes).toHaveLength(4);
    expect(forkNavigationItems).toEqual([
      { key: 'client-activity', label: 'Client activity', path: '/activity' },
      { key: 'policy-engine', label: 'Policy engine', path: '/policies' },
      { key: 'audit', label: 'Audit', path: '/audit' },
      { key: 'webhooks', label: 'Webhooks', path: '/webhooks' },
    ]);
    expect(forkApiSections).toHaveLength(2);
    expect(forkApiSections[0].id).toBe('risk-intelligence');
    expect(forkApiSections[0].endpoints).toHaveLength(7);
    expect(
      forkApiSections[0].endpoints.map((endpoint) => `${endpoint.method} ${endpoint.path}`),
    ).toContain('GET /panel/api/risk/clients/:email');
    expect(forkApiSections[1].id).toBe('audit-webhooks-metrics');
    expect(forkApiSections[1].endpoints).toHaveLength(10);
  });
});
