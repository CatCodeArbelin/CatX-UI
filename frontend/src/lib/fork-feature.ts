export type FeatureResponse = { status?: number; msg?: string; featureDisabled?: boolean };

export type ForkFeature =
  | 'analytics'
  | 'dns_intelligence'
  | 'audit'
  | 'webhooks'
  | 'self_service'
  | 'fleet_updates'
  | 'policies'
  | 'risk'
  | 'sponsors';

/**
 * Only feature-owned entrypoints may use the legacy 404 fallback. A normal
 * API response never becomes a feature-off state through this helper. New
 * endpoints should return featureDisabled explicitly when possible.
 */
export function isKnownForkFeatureUnavailable(
  response: FeatureResponse | null | undefined,
  feature: ForkFeature,
): boolean {
  if (response?.featureDisabled === true) return true;
  if (
    feature === 'self_service' &&
    response?.msg?.trim().toLowerCase() === 'self-service is disabled'
  ) {
    return true;
  }
  const featureOwnedEntrypoint = new Set<ForkFeature>([
    'analytics',
    'dns_intelligence',
    'audit',
    'webhooks',
    'self_service',
    'fleet_updates',
    'policies',
    'risk',
  ]);
  return (
    featureOwnedEntrypoint.has(feature) &&
    response?.status === 404 &&
    response.msg === 'Request failed with status 404'
  );
}
