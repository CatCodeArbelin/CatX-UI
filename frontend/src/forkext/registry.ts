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
  {
    path: '/audit',
    lazy: async () => ({ Component: (await import('../pages/audit/AuditPage')).default }),
  },
  {
    path: '/webhooks',
    lazy: async () => ({ Component: (await import('../pages/audit/WebhooksPage')).default }),
  },
];
export const forkNavigationItems: readonly ForkNavigationItem[] = [
  { key: 'client-activity', label: 'Client activity', path: '/activity' },
  { key: 'policy-engine', label: 'Policy engine', path: '/policies' },
  { key: 'audit', label: 'Audit', path: '/audit' },
  { key: 'webhooks', label: 'Webhooks', path: '/webhooks' },
];
export const forkApiSections = [
  {
    id: 'risk-intelligence',
    title: 'Risk intelligence',
    description:
      'Metadata-only, explainable risk signals. Informational only; no automatic enforcement is performed.',
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
  {
    id: 'audit-webhooks-metrics',
    title: 'Audit / Webhooks / Metrics',
    description:
      'Durable metadata-only audit records, signed webhook delivery, and bounded Prometheus metrics.',
    endpoints: [
      {
        method: 'GET',
        path: '/panel/api/fork/audit/events',
        summary: 'List durable audit events.',
        responseSchema: 'EventPage',
      },
      {
        method: 'GET',
        path: '/panel/api/fork/audit/events/:id',
        summary: 'Read one sanitized audit event.',
        responseSchema: 'AuditEvent',
      },
      {
        method: 'DELETE',
        path: '/panel/api/fork/audit/events',
        summary: 'Delete audit events before a timestamp.',
        params: [{ name: 'before', in: 'query', type: 'string' }],
      },
      {
        method: 'GET',
        path: '/panel/api/fork/audit/webhooks',
        summary: 'List webhook destinations without secrets.',
        responseSchema: 'WebhookEndpoint',
        responseSchemaArray: true,
      },
      {
        method: 'POST',
        path: '/panel/api/fork/audit/webhooks',
        summary: 'Create a public HTTPS webhook destination.',
        requestSchema: { $ref: '#/components/schemas/WebhookEndpointRequest' },
        responseSchema: 'WebhookEndpoint',
      },
      {
        method: 'DELETE',
        path: '/panel/api/fork/audit/webhooks/:id',
        summary: 'Delete a webhook destination.',
      },
      {
        method: 'POST',
        path: '/panel/api/fork/audit/webhooks/:id/replay',
        summary: 'Queue the latest audit event for delivery to a destination.',
        responseSchema: 'WebhookDelivery',
      },
      {
        method: 'GET',
        path: '/panel/api/fork/audit/webhooks/deliveries',
        summary: 'List webhook delivery attempts.',
        responseSchema: 'DeliveryPage',
      },
      {
        method: 'POST',
        path: '/panel/api/fork/audit/webhooks/deliveries/:id/replay',
        summary: 'Replay a delivery with a new delivery ID.',
        responseSchema: 'WebhookDelivery',
      },
      {
        method: 'GET',
        path: '/panel/api/fork/audit/retention',
        summary: 'Read audit and webhook retention.',
        responseSchema: 'AuditRetentionSettings',
      },
      {
        method: 'POST',
        path: '/panel/api/fork/audit/retention',
        summary: 'Update audit and webhook retention.',
        requestSchema: { $ref: '#/components/schemas/AuditRetentionSettings' },
        responseSchema: 'AuditRetentionSettings',
      },
      {
        method: 'GET',
        path: '/panel/api/fork/metrics',
        summary: 'Prometheus metrics with fixed bounded labels.',
      },
    ],
  },
] satisfies import('../pages/api-docs/endpoints').Section[];
