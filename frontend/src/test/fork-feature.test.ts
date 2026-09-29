import { describe, expect, it } from 'vitest';
import { isKnownForkFeatureUnavailable } from '@/lib/fork-feature';

describe('fork feature-off classification', () => {
  it('requires a known feature-owned entrypoint for legacy 404 fallback', () => {
    expect(
      isKnownForkFeatureUnavailable(
        { status: 404, msg: 'Request failed with status 404' },
        'audit',
      ),
    ).toBe(true);
    expect(
      isKnownForkFeatureUnavailable({ status: 404, msg: 'Request failed with status 404' }, 'risk'),
    ).toBe(true);
    expect(
      isKnownForkFeatureUnavailable(
        { status: 404, msg: 'Request failed with status 404' },
        'policies',
      ),
    ).toBe(true);
  });

  it('does not reinterpret an ordinary 404 or another HTTP error', () => {
    expect(isKnownForkFeatureUnavailable({ status: 404, msg: 'client not found' }, 'audit')).toBe(
      false,
    );
    expect(isKnownForkFeatureUnavailable({ status: 500, msg: 'server error' }, 'audit')).toBe(
      false,
    );
  });

  it('accepts an explicit feature-disabled envelope', () => {
    expect(isKnownForkFeatureUnavailable({ featureDisabled: true }, 'webhooks')).toBe(true);
  });

  it('accepts the explicit self-service disabled response without masking other errors', () => {
    expect(isKnownForkFeatureUnavailable({ msg: 'self-service is disabled' }, 'self_service')).toBe(
      true,
    );
    expect(isKnownForkFeatureUnavailable({ msg: 'invalid token' }, 'self_service')).toBe(false);
  });
});
