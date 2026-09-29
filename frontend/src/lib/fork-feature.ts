export type FeatureResponse = { status?: number; msg?: string; featureDisabled?: boolean };

export type ForkFeature =
  | 'audit'
  | 'webhooks'
  | 'self_service'
  | 'fleet_updates'
  | 'policies'
  | 'risk';

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
  const featureOwnedEntrypoint = new Set<ForkFeature>([
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
