import type { RouteObject } from 'react-router';

export interface ForkNavigationItem {
  key: string;
  label: string;
  path: string;
}

export const forkRoutes: readonly RouteObject[] = [
  {
    path: '/activity',
    lazy: async () => ({ Component: (await import('../pages/activity/ActivityPage')).default }),
  },
  {
    path: '/policies',
    lazy: async () => ({ Component: (await import('../pages/policy/PolicyPage')).default }),
  },
];
export const forkNavigationItems: readonly ForkNavigationItem[] = [
  { key: 'client-activity', label: 'Client activity', path: '/activity' },
  { key: 'policy-engine', label: 'Policy engine', path: '/policies' },
];
export const forkApiSections = [
  {
    id: 'risk-intelligence',
    title: 'Risk intelligence',
    description: 'Metadata-only, explainable risk signals. No automatic enforcement is performed.',
    endpoints: [
      {
        method: 'GET',
        path: '/panel/api/risk/clients/:email',
        summary: 'Return the client risk score, evidence timeline, and source IP history.',
        responseSchema: 'Summary',
      },
      {
        method: 'POST',
        path: '/panel/api/risk/clients/:email/events/:id/ack',
        summary: 'Acknowledge a risk event.',
      },
      {
        method: 'POST',
        path: '/panel/api/risk/clients/:email/suppress',
        summary: 'Create an expiring informational signal suppression.',
        requestSchema: {
          type: 'object',
          properties: {
            kind: { type: 'string' },
            expiresAt: { type: 'integer' },
            reason: { type: 'string' },
          },
          required: ['kind'],
        },
      },
      {
        method: 'DELETE',
        path: '/panel/api/risk/clients/:email/ip-history',
        summary: 'Delete the client IP history without deleting upstream accounting.',
      },
      {
        method: 'DELETE',
        path: '/panel/api/risk/clients/:email/events',
        summary: 'Delete the client risk event history.',
      },
      {
        method: 'GET',
        path: '/panel/api/risk/settings',
        summary: 'Read independent risk retention settings.',
        responseSchema: 'RetentionSettings',
      },
      {
        method: 'POST',
        path: '/panel/api/risk/settings',
        summary: 'Update independent risk retention settings.',
        requestSchema: { $ref: '#/components/schemas/RetentionSettings' },
        responseSchema: 'RetentionSettings',
      },
    ],
  },
] satisfies import('../pages/api-docs/endpoints').Section[];
