import { describe, expect, it } from 'vitest';

import { forkApiSections, forkNavigationItems, forkRoutes } from '@/forkext/registry';

describe('fork registries', () => {
  it('registers the client activity page through the fork boundary', () => {
    expect(forkRoutes).toHaveLength(8);
    expect(forkNavigationItems).toEqual([
      expect.objectContaining({
        key: 'client-activity',
        path: '/activity',
        group: 'clients',
        icon: 'activity',
      }),
      expect.objectContaining({
        key: 'portal-access',
        path: '/portal-access',
        group: 'clients',
        icon: 'portal',
      }),
      expect.objectContaining({ key: 'fleet', path: '/fleet', group: 'nodes', icon: 'fleet' }),
      expect.objectContaining({
        key: 'fleet-updates',
        path: '/fleet-updates',
        group: 'nodes',
        icon: 'fleet-updates',
      }),
      expect.objectContaining({
        key: 'policy-engine',
        path: '/policies',
        group: 'routing',
        icon: 'policy',
      }),
      expect.objectContaining({ key: 'audit', path: '/audit', group: 'operations', icon: 'audit' }),
      expect.objectContaining({
        key: 'webhooks',
        path: '/webhooks',
        group: 'operations',
        icon: 'webhooks',
      }),
      expect.objectContaining({
        key: 'sponsors',
        path: '/catx/sponsors',
        group: 'operations',
        icon: 'sponsors',
      }),
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
