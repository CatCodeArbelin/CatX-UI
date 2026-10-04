import { describe, expect, it } from 'vitest';
import {
  mergeStructuredPolicyDefinition,
  policyDefinitionSummary,
  structuredFromPolicyDefinition,
} from '@/pages/policy/policyDefinition';

describe('structured policy definition editor', () => {
  it('round-trips supported fields while preserving unknown policy extensions', () => {
    const original = {
      action: 'allow',
      destinations: ['example.com'],
      metadata: { owner: 'operator' },
      dns: { managed: true, customResolverMode: 'safe' },
    };
    const values = structuredFromPolicyDefinition(original);
    values.action = 'deny';
    values.destinations = ['blocked.example'];
    values.resolver = 'secure-dns';
    const merged = mergeStructuredPolicyDefinition(original, values);
    expect(merged).toMatchObject({
      action: 'deny',
      destinations: ['blocked.example'],
      metadata: { owner: 'operator' },
      dns: { customResolverMode: 'safe', managed: true, dnsOutboundTag: 'secure-dns' },
    });
  });

  it('removes cleared known fields without deleting unknown extensions', () => {
    const merged = mergeStructuredPolicyDefinition(
      {
        action: 'allow',
        quarantine: true,
        quarantineAllowlist: ['example.com'],
        qos: { mode: 'fair' },
      },
      {
        action: 'allow',
        services: [],
        categories: [],
        destinations: [],
        quarantine: false,
        quarantineAllowlist: [],
        managedDns: false,
        safeSearch: false,
        resolver: '',
        scheduleRef: '',
      },
    );
    expect(merged).toEqual({ action: 'allow', qos: { mode: 'fair' } });
  });

  it('provides a human summary instead of exposing raw JSON in the table', () => {
    expect(policyDefinitionSummary('{"action":"deny","destinations":["example.com"]}')).toBe(
      'deny · example.com',
    );
    expect(policyDefinitionSummary('{"metadata":true}')).toBe('');
  });
});
