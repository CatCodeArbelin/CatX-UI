/** @vitest-environment jsdom */

import { expect, test } from 'vitest';

import {
  campaignIsAbortable,
  campaignIsReconcileable,
  campaignIsRetryable,
} from '@/pages/fleet/FleetUpdatePage';

test('fleet update actions follow persisted campaign state', () => {
  expect(campaignIsAbortable('ready')).toBe(true);
  expect(campaignIsReconcileable('waiting_restart')).toBe(true);
  expect(campaignIsAbortable('blocked')).toBe(false);
  expect(campaignIsAbortable('succeeded')).toBe(false);
  expect(campaignIsRetryable('blocked')).toBe(true);
  expect(campaignIsRetryable('failed')).toBe(true);
  expect(campaignIsRetryable('ready')).toBe(false);
});
