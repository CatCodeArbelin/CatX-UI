import type { RouteObject } from 'react-router';

export interface ForkNavigationItem {
  key: string;
  labelKey: string;
  path: string;
  group: 'clients' | 'nodes' | 'routing' | 'operations';
  icon:
    | 'activity'
    | 'portal'
    | 'fleet'
    | 'fleet-updates'
    | 'policy'
    | 'audit'
    | 'webhooks'
    | 'sponsors';
}

const managedSponsorSchema = {
  type: 'object',
  required: ['id', 'enabled', 'name', 'priority', 'slots', 'destinationUrl', 'title', 'text'],
  properties: {
    id: { type: 'string' },
    enabled: { type: 'boolean' },
    name: { type: 'string' },
    priority: { type: 'integer' },
    slots: { type: 'array', items: { type: 'string', enum: ['dashboard', 'sidebar', 'page'] } },
    startAt: { type: 'string', format: 'date-time' },
    endAt: { type: 'string', format: 'date-time' },
    destinationUrl: { type: 'string', format: 'uri' },
    logoUrl: { type: 'string', format: 'uri' },
    title: { type: 'object', additionalProperties: { type: 'string' } },
    text: { type: 'object', additionalProperties: { type: 'string' } },
    createdAt: { type: 'string', format: 'date-time' },
    updatedAt: { type: 'string', format: 'date-time' },
  },
} as const;

const managedSponsorInputSchema = {
  ...managedSponsorSchema,
  required: ['enabled', 'name', 'priority', 'slots', 'destinationUrl', 'title', 'text'],
} as const;

export const forkRoutes: readonly RouteObject[] = [
  {
    path: '/fleet-updates',
    lazy: async () => ({ Component: (await import('../pages/fleet/FleetUpdatePage')).default }),
  },
  {
    path: '/portal-access',
    lazy: async () => ({ Component: (await import('../pages/portal/PortalAdminPage')).default }),
  },
  {
    path: '/fleet',
    lazy: async () => ({ Component: (await import('../pages/fleet/FleetPage')).default }),
  },
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
  {
    path: '/catx/sponsors',
    lazy: async () => ({ Component: (await import('./sponsors/SponsorsPage')).default }),
  },
  {
    path: '/catx/sponsors/manage',
    lazy: async () => ({ Component: (await import('./sponsors/SponsorsManagementPage')).default }),
  },
];
export const forkNavigationItems: readonly ForkNavigationItem[] = [
  {
    key: 'client-activity',
    labelKey: 'fork.activity.title',
    path: '/activity',
    group: 'clients',
    icon: 'activity',
  },
  {
    key: 'portal-access',
    labelKey: 'fork.portal.adminTitle',
    path: '/portal-access',
    group: 'clients',
    icon: 'portal',
  },
  { key: 'fleet', labelKey: 'fork.fleet.title', path: '/fleet', group: 'nodes', icon: 'fleet' },
  {
    key: 'fleet-updates',
    labelKey: 'fork.fleetUpdate.title',
    path: '/fleet-updates',
    group: 'nodes',
    icon: 'fleet-updates',
  },
  {
    key: 'policy-engine',
    labelKey: 'fork.policy.title',
    path: '/policies',
    group: 'routing',
    icon: 'policy',
  },
  {
    key: 'audit',
    labelKey: 'fork.audit.title',
    path: '/audit',
    group: 'operations',
    icon: 'audit',
  },
  {
    key: 'webhooks',
    labelKey: 'fork.webhooks.title',
    path: '/webhooks',
    group: 'operations',
    icon: 'webhooks',
  },
  {
    key: 'sponsors',
    labelKey: 'fork.sponsors.title',
    path: '/catx/sponsors',
    group: 'operations',
    icon: 'sponsors',
  },
  {
    key: 'sponsors-management',
    labelKey: 'fork.sponsors.title',
    path: '/catx/sponsors/manage',
    group: 'operations',
    icon: 'sponsors',
  },
];

export const forkNavigationGroups = {
  clients: ['client-activity', 'portal-access'],
  nodes: ['fleet', 'fleet-updates'],
  routing: ['policy-engine'],
  operations: ['audit', 'webhooks', 'sponsors', 'sponsors-management'],
} as const;
// Keep the fork contract registry as the single source for generated API verification.
export const forkApiSections = [
  {
    id: 'fleet-updates',
    title: 'Fleet updates',
    translationKey: 'fork.apiDocs.fleetUpdates',
    description:
      'Protected Stage A update campaign planning and reconciliation; production dispatch is disabled.',
    endpoints: [
      {
        method: 'POST',
        path: '/panel/api/fleet-updates/campaigns',
        summary: 'Create a dry-run update campaign plan.',
        responseSchema: 'Plan',
      },
      {
        method: 'GET',
        path: '/panel/api/fleet-updates/campaigns',
        summary: 'List update campaigns.',
        responseSchema: 'Campaign',
        responseSchemaArray: true,
      },
      {
        method: 'GET',
        path: '/panel/api/fleet-updates/campaigns/:id',
        summary: 'Read a campaign and immutable target snapshot.',
        responseSchema: 'Plan',
      },
      {
        method: 'POST',
        path: '/panel/api/fleet-updates/campaigns/:id/reconcile',
        summary: 'Reconcile a campaign safely.',
        responseSchema: 'Campaign',
      },
      {
        method: 'POST',
        path: '/panel/api/fleet-updates/campaigns/:id/abort',
        summary: 'Abort a campaign.',
        responseSchema: 'Campaign',
      },
      {
        method: 'POST',
        path: '/panel/api/fleet-updates/campaigns/:id/retry',
        summary: 'Retry failed or unknown targets.',
        responseSchema: 'Campaign',
      },
    ],
  },
  {
    id: 'risk-intelligence',
    title: 'Risk intelligence',
    translationKey: 'fork.apiDocs.risk',
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
    translationKey: 'fork.apiDocs.audit',
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
  {
    id: 'self-service-portal',
    title: 'Self-service portal',
    translationKey: 'fork.apiDocs.portal',
    description: 'Dedicated client portal sessions with metadata-only self-service access.',
    endpoints: [
      {
        method: 'POST',
        path: '/portal/auth',
        summary: 'Establish a dedicated portal session from a one-time-issued portal token.',
      },
      {
        method: 'GET',
        path: '/portal/me',
        summary: 'Read the authenticated client profile and subscription status.',
      },
      {
        method: 'GET',
        path: '/portal/devices',
        summary: 'List devices owned by the authenticated client.',
      },
      { method: 'PUT', path: '/portal/devices/:id', summary: 'Rename an owned device.' },
      { method: 'DELETE', path: '/portal/devices/:id', summary: 'Revoke an owned device.' },
      { method: 'GET', path: '/portal/hosts', summary: 'List sanitized visible connection hosts.' },
      {
        method: 'GET',
        path: '/portal/traffic',
        summary: 'Read own traffic, quota, and expiry status.',
      },
      {
        method: 'POST',
        path: '/portal/access/rotate',
        summary: 'Rotate own portal access and invalidate existing sessions.',
      },
      {
        method: 'POST',
        path: '/portal/access/revoke',
        summary: 'Revoke own portal access and invalidate existing sessions.',
      },
      {
        method: 'GET',
        path: '/panel/api/portal/credentials',
        summary: 'Admin-only portal credential inventory.',
        responseSchema: 'CredentialView',
        responseSchemaArray: true,
      },
      {
        method: 'POST',
        path: '/panel/api/portal/credentials',
        summary: 'Admin-only one-time portal token issuance.',
        requestSchema: { type: 'object' },
        responseSchema: 'IssuedToken',
      },
      {
        method: 'POST',
        path: '/panel/api/portal/credentials/:clientId/rotate',
        summary: 'Admin-only portal token rotation.',
        responseSchema: 'IssuedToken',
      },
      {
        method: 'POST',
        path: '/panel/api/portal/credentials/:clientId/revoke',
        summary: 'Admin-only portal token revocation.',
      },
      {
        method: 'GET',
        path: '/panel/api/portal/settings',
        summary: 'Read the admin-only portal feature setting.',
      },
      {
        method: 'GET',
        path: '/panel/api/portal/options',
        summary: 'Read searchable client, group, and host labels for admin selectors.',
        responseObjectSchema: {
          type: 'object',
          properties: {
            clients: { type: 'array', items: { type: 'object' } },
            groups: { type: 'array', items: { type: 'object' } },
            hosts: { type: 'array', items: { type: 'object' } },
            state: { type: 'string', enum: ['active', 'feature_off'] },
            featureDisabled: { type: 'boolean' },
          },
        },
      },
      {
        method: 'POST',
        path: '/panel/api/portal/settings',
        summary: 'Update the admin-only portal feature setting.',
      },
      {
        method: 'GET',
        path: '/panel/api/portal/host-grants',
        summary: 'Admin-only host visibility grant inventory.',
        responseSchema: 'HostGrant',
        responseSchemaArray: true,
      },
      {
        method: 'POST',
        path: '/panel/api/portal/host-grants',
        summary: 'Admin-only positive host visibility grant.',
        requestSchema: { $ref: '#/components/schemas/HostGrant' },
        responseSchema: 'HostGrant',
      },
      {
        method: 'DELETE',
        path: '/panel/api/portal/host-grants/:id',
        summary: 'Admin-only host visibility grant deletion.',
      },
    ],
  },
  {
    id: 'feature-settings',
    title: 'CatX-UI Features',
    translationKey: 'fork.apiDocs.settings',
    description: 'Protected CatX feature flags and restart requirement.',
    endpoints: [
      {
        method: 'GET',
        path: '/panel/api/fork/settings/features',
        summary: 'Read CatX feature flags and restart requirement.',
      },
      {
        method: 'PUT',
        path: '/panel/api/fork/settings/features',
        summary: 'Update CatX feature flags.',
      },
    ],
  },
  {
    id: 'catx-sponsors',
    title: 'CatX Sponsors',
    translationKey: 'fork.apiDocs.sponsors',
    description: 'Authenticated, operator-configured sponsor metadata and safe logo proxying.',
    endpoints: [
      {
        method: 'GET',
        path: '/panel/api/fork/sponsors',
        summary: 'Read the active CatX sponsor metadata.',
        responseSchema: 'SponsorList',
      },
      {
        method: 'GET',
        path: '/panel/api/fork/sponsors/logo/:name',
        summary: 'Read a validated active sponsor logo.',
        responses: {
          '200': {
            description: 'Validated sponsor image bytes.',
            content: {
              'image/png': { schema: { type: 'string', format: 'binary' } },
              'image/webp': { schema: { type: 'string', format: 'binary' } },
              'image/jpeg': { schema: { type: 'string', format: 'binary' } },
            },
          },
          '404': { description: 'The logo is not part of an active sponsor set.' },
          '503': { description: 'The configured sponsor source or logo is unavailable.' },
        },
      },
      {
        method: 'GET',
        path: '/panel/api/fork/sponsors/settings',
        summary: 'Read sanitized sponsor source settings.',
        responseObjectSchema: {
          type: 'object',
          required: ['enabled', 'configured', 'sourceUrl', 'contactUrl', 'providerMode'],
          properties: {
            enabled: { type: 'boolean' },
            configured: { type: 'boolean' },
            sourceUrl: { type: 'string' },
            contactUrl: { type: 'string' },
            providerMode: { type: 'string', enum: ['local', 'remote'] },
          },
        },
      },
      {
        method: 'PUT',
        path: '/panel/api/fork/sponsors/settings',
        summary: 'Update the operator-configured provider mode, source, and contact URLs.',
        requestSchema: {
          type: 'object',
          properties: {
            sourceUrl: { type: 'string' },
            contactUrl: { type: 'string' },
            providerMode: { type: 'string', enum: ['local', 'remote'] },
          },
        },
        responseObjectSchema: {
          type: 'object',
          required: ['enabled', 'configured', 'sourceUrl', 'contactUrl', 'providerMode'],
          properties: {
            enabled: { type: 'boolean' },
            configured: { type: 'boolean' },
            sourceUrl: { type: 'string' },
            contactUrl: { type: 'string' },
            providerMode: { type: 'string', enum: ['local', 'remote'] },
          },
        },
      },
      {
        method: 'GET',
        path: '/panel/api/fork/sponsors/status',
        summary: 'Read local/remote provider status and bounded cache state.',
        responseObjectSchema: {
          type: 'object',
          required: [
            'enabled',
            'providerMode',
            'localSponsorCount',
            'activeSponsorCount',
            'remoteProviderConfigured',
            'cacheState',
          ],
          properties: {
            enabled: { type: 'boolean' },
            providerMode: { type: 'string', enum: ['local', 'remote'] },
            localSponsorCount: { type: 'integer' },
            activeSponsorCount: { type: 'integer' },
            remoteProviderConfigured: { type: 'boolean' },
            lastSuccessfulFetch: { type: 'string', format: 'date-time' },
            lastError: { type: 'string' },
            cacheState: { type: 'string' },
          },
        },
      },
      {
        method: 'GET',
        path: '/panel/api/fork/sponsors/manage',
        summary: 'List all locally managed sponsors in deterministic order.',
        responseObjectSchema: { type: 'array', items: managedSponsorSchema },
      },
      {
        method: 'POST',
        path: '/panel/api/fork/sponsors',
        summary: 'Create a locally managed sponsor.',
        requestSchema: managedSponsorInputSchema,
        responseObjectSchema: managedSponsorSchema,
      },
      {
        method: 'GET',
        path: '/panel/api/fork/sponsors/:id',
        summary: 'Read one locally managed sponsor.',
        responseObjectSchema: managedSponsorSchema,
      },
      {
        method: 'PUT',
        path: '/panel/api/fork/sponsors/:id',
        summary:
          'Replace a locally managed sponsor; the path ID is authoritative and omitted fields are cleared or defaulted deterministically.',
        requestSchema: managedSponsorInputSchema,
        responseObjectSchema: managedSponsorSchema,
      },
      {
        method: 'DELETE',
        path: '/panel/api/fork/sponsors/:id',
        summary: 'Delete a locally managed sponsor.',
      },
    ],
  },
] satisfies import('../pages/api-docs/endpoints').Section[];
