export type PolicyDefinition = Record<string, unknown>;

export interface StructuredPolicyValues {
  action: string;
  services: string[];
  categories: string[];
  destinations: string[];
  quarantine: boolean;
  quarantineAllowlist: string[];
  managedDns: boolean;
  safeSearch: boolean;
  resolver: string;
  scheduleRef: string;
}

const stringList = (value: unknown): string[] =>
  Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string') : [];

export function structuredFromPolicyDefinition(spec: PolicyDefinition): StructuredPolicyValues {
  const dns =
    spec.dns && typeof spec.dns === 'object' && !Array.isArray(spec.dns)
      ? (spec.dns as Record<string, unknown>)
      : {};
  return {
    action: typeof spec.action === 'string' ? spec.action : 'allow',
    services: stringList(spec.services),
    categories: stringList(spec.categories),
    destinations: stringList(spec.destinations),
    quarantine: spec.quarantine === true,
    quarantineAllowlist: stringList(spec.quarantineAllowlist),
    managedDns: dns.managed === true,
    safeSearch: dns.safeSearch === true,
    resolver: typeof dns.dnsOutboundTag === 'string' ? dns.dnsOutboundTag : '',
    scheduleRef: typeof spec.scheduleRef === 'string' ? spec.scheduleRef : '',
  };
}

export function mergeStructuredPolicyDefinition(
  original: PolicyDefinition,
  values: StructuredPolicyValues,
): PolicyDefinition {
  const next: PolicyDefinition = { ...original };
  const setList = (key: string, value: string[]) => {
    if (value.length) next[key] = [...value];
    else delete next[key];
  };

  if (values.action) next.action = values.action;
  else delete next.action;
  setList('services', values.services);
  setList('categories', values.categories);
  setList('destinations', values.destinations);

  if (values.quarantine) {
    next.quarantine = true;
    setList('quarantineAllowlist', values.quarantineAllowlist);
  } else {
    delete next.quarantine;
    delete next.quarantineAllowlist;
  }

  const originalDns =
    next.dns && typeof next.dns === 'object' && !Array.isArray(next.dns)
      ? { ...(next.dns as Record<string, unknown>) }
      : {};
  if (values.managedDns) originalDns.managed = true;
  else delete originalDns.managed;
  if (values.safeSearch) originalDns.safeSearch = true;
  else delete originalDns.safeSearch;
  if (values.resolver.trim()) originalDns.dnsOutboundTag = values.resolver.trim();
  else delete originalDns.dnsOutboundTag;
  if (Object.keys(originalDns).length) next.dns = originalDns;
  else delete next.dns;

  if (values.scheduleRef.trim()) next.scheduleRef = values.scheduleRef.trim();
  else delete next.scheduleRef;
  return next;
}

export function policyDefinitionSummary(spec: string): string {
  try {
    const parsed = JSON.parse(spec) as unknown;
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return '';
    const value = parsed as PolicyDefinition;
    const action = typeof value.action === 'string' ? value.action : '';
    const destinations = stringList(value.destinations).slice(0, 2).join(', ');
    const categories = stringList(value.categories).slice(0, 2).join(', ');
    return [action, destinations || categories].filter(Boolean).join(' · ');
  } catch {
    return '';
  }
}
