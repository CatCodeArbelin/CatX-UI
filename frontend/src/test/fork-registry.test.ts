import { describe, expect, it } from 'vitest';

import { forkApiSections, forkNavigationItems, forkRoutes } from '@/forkext/registry';

describe('fork registries', () => {
  it('registers the client activity page through the fork boundary', () => {
    expect(forkRoutes).toHaveLength(7);
    expect(forkNavigationItems).toEqual([
      { key: 'fleet-updates', label: 'Fleet updates', path: '/fleet-updates' },
      { key: 'portal-access', label: 'Portal access', path: '/portal-access' },
      { key: 'fleet', label: 'Fleet', path: '/fleet' },
      { key: 'client-activity', label: 'Client activity', path: '/activity' },
      { key: 'policy-engine', label: 'Policy engine', path: '/policies' },
      { key: 'audit', label: 'Audit', path: '/audit' },
      { key: 'webhooks', label: 'Webhooks', path: '/webhooks' },
    ]);
    const risk = forkApiSections.find((section) => section.id === 'risk-intelligence');
    expect(risk?.endpoints).toHaveLength(7);
    expect(risk?.endpoints.map((endpoint) => `${endpoint.method} ${endpoint.path}`)).toContain(
      'GET /panel/api/risk/clients/:email',
    );
    const audit = forkApiSections.find((section) => section.id === 'audit-webhooks-metrics');
    expect(audit?.endpoints).toHaveLength(12);
    const fleet = forkApiSections.find((section) => section.id === 'fleet-updates');
    expect(fleet?.endpoints.map((endpoint) => `${endpoint.method} ${endpoint.path}`)).toEqual(
      expect.arrayContaining([
        'POST /panel/api/fleet-updates/campaigns',
        'GET /panel/api/fleet-updates/campaigns',
        'POST /panel/api/fleet-updates/campaigns/:id/reconcile',
      ]),
    );
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
