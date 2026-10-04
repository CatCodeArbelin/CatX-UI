export type ForkRuntimeState =
  | 'feature_off'
  | 'active'
  | 'disabled'
  | 'initializing'
  | 'restart_required'
  | 'empty'
  | 'unconfigured'
  | 'unsupported'
  | 'degraded'
  | 'error'
  | 'loading';

export type FeatureResponse = {
  status?: number;
  msg?: string;
  featureDisabled?: boolean;
  state?: ForkRuntimeState;
  obj?: unknown;
};

export type ForkFeature =
  | 'analytics'
  | 'dns_intelligence'
  | 'audit'
  | 'webhooks'
  | 'self_service'
  | 'fleet_updates'
  | 'policies'
  | 'risk'
  | 'sponsors'
  | 'traffic_control';

/**
 * Only feature-owned entrypoints may use the legacy 404 fallback. A normal
 * API response never becomes a feature-off state through this helper. New
 * endpoints should return featureDisabled explicitly when possible.
 */
export function isKnownForkFeatureUnavailable(
  response: FeatureResponse | null | undefined,
  feature: ForkFeature,
): boolean {
  const object =
    response?.obj && typeof response.obj === 'object' && !Array.isArray(response.obj)
      ? (response.obj as { featureDisabled?: boolean; state?: ForkRuntimeState })
      : undefined;
  if (response?.featureDisabled === true || object?.featureDisabled === true) return true;
  if (response?.state === 'feature_off' || object?.state === 'feature_off') return true;
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
    'traffic_control',
  ]);
  return (
    featureOwnedEntrypoint.has(feature) &&
    response?.status === 404 &&
    response.msg === 'Request failed with status 404'
  );
}
