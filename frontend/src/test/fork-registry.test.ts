import { describe, expect, it } from 'vitest';

import { forkApiSections, forkNavigationItems, forkRoutes } from '@/forkext/registry';

describe('fork registries', () => {
  it('registers the client activity page through the fork boundary', () => {
    expect(forkRoutes).toHaveLength(6);
    expect(forkNavigationItems).toEqual([
      { key: 'portal-access', label: 'Portal access', path: '/portal-access' },
      { key: 'fleet', label: 'Fleet', path: '/fleet' },
      { key: 'client-activity', label: 'Client activity', path: '/activity' },
      { key: 'policy-engine', label: 'Policy engine', path: '/policies' },
      { key: 'audit', label: 'Audit', path: '/audit' },
      { key: 'webhooks', label: 'Webhooks', path: '/webhooks' },
    ]);
    expect(forkApiSections[0].id).toBe('risk-intelligence');
    expect(forkApiSections[0].endpoints).toHaveLength(7);
    expect(
      forkApiSections[0].endpoints.map((endpoint) => `${endpoint.method} ${endpoint.path}`),
    ).toContain('GET /panel/api/risk/clients/:email');
    expect(forkApiSections[1].id).toBe('audit-webhooks-metrics');
    expect(forkApiSections[1].endpoints).toHaveLength(12);
    const portal = forkApiSections.find((section) => section.id === 'self-service-portal');
    expect(portal?.endpoints.map((endpoint) => `${endpoint.method} ${endpoint.path}`)).toEqual(
      expect.arrayContaining([
        'POST /portal/auth',
        'GET /portal/me',
        'GET /portal/devices',
        'PUT /portal/devices/:id',
        'DELETE /portal/devices/:id',
        'GET /portal/hosts',
        'GET /portal/traffic',
        'POST /portal/access/rotate',
        'POST /portal/access/revoke',
      ]),
    );
  });
});
