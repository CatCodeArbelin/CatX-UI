import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, test, vi } from 'vitest';

import TrafficPolicyPanel from '@/pages/clients/TrafficPolicyPanel';
import { HttpUtil, Msg } from '@/utils';
import { renderWithProviders } from './test-utils';

afterEach(() => vi.restoreAllMocks());

const configured = {
  enabled: true,
  windowSeconds: 3600,
  quotaBytes: 1048576,
  activeUploadBps: 0,
  activeDownloadBps: 0,
  throttleUploadBps: 0,
  throttleDownloadBps: 0,
  lifecycle: 'active',
  reason: '',
  usedBytes: 0,
  remainingBytes: 1048576,
  enforcement: 'unsupported',
  state: 'active',
  featureDisabled: false,
};

test('offers an explicit configure action and persists a real policy', async () => {
  const get = vi
    .spyOn(HttpUtil, 'get')
    .mockResolvedValueOnce(new Msg(true, '', { state: 'unconfigured', featureDisabled: false }));
  const put = vi.spyOn(HttpUtil, 'put').mockResolvedValue(new Msg(true, '', configured));
  renderWithProviders(<TrafficPolicyPanel email="review@example.invalid" />);

  expect((await screen.findByRole('alert')).textContent).toContain('Not configured');
  fireEvent.click(screen.getByRole('button', { name: 'Configure traffic control' }));
  expect(screen.getByText('Quota and window')).toBeTruthy();
  fireEvent.click(screen.getByRole('button', { name: 'Save' }));

  await waitFor(() => expect(put).toHaveBeenCalledOnce());
  expect(put.mock.calls[0][1]).toMatchObject({ enabled: true, windowSeconds: 3600 });
  expect(get).toHaveBeenCalledOnce();
});

test('keeps Client Information traffic controls read-only while exposing unsupported truthfully', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValueOnce(new Msg(true, '', configured));
  renderWithProviders(<TrafficPolicyPanel email="review@example.invalid" readOnly />);

  expect(await screen.findByText('Unsupported')).toBeTruthy();
  expect(screen.queryByRole('button', { name: 'Save' })).toBeNull();
  expect(
    screen.getAllByText('Rate/speed shaping is unsupported for generic Xray users.').length,
  ).toBe(2);
  expect(screen.queryByText(/Upload B\/s:/)).toBeNull();
});
