import { fireEvent, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { afterEach, expect, test, vi } from 'vitest';

import CatxFeaturesTab from '@/pages/settings/CatxFeaturesTab';
import WebhooksPage from '@/pages/audit/WebhooksPage';
import { HttpUtil, Msg } from '@/utils';
import { renderWithProviders } from './test-utils';

vi.mock('@/api/queries/useAllSettings', () => ({
  useAllSettings: () => ({ allSetting: {} }),
}));

const items = [
  { key: 'analytics.enabled', enabled: false, restartRequired: true },
  {
    key: 'dns_intelligence.enabled',
    enabled: false,
    requires: ['analytics.enabled'],
    restartRequired: true,
  },
  { key: 'policies.enabled', enabled: false, restartRequired: true },
  { key: 'traffic_control.enabled', enabled: false, restartRequired: true },
  {
    key: 'security_anomaly.enabled',
    enabled: false,
    requires: ['analytics.enabled'],
    restartRequired: true,
  },
  { key: 'audit.enabled', enabled: false, restartRequired: true },
  { key: 'self_service.enabled', enabled: false, restartRequired: true },
  { key: 'fleet_updates.enabled', enabled: false, restartRequired: true },
  {
    key: 'fleet_updates.mutation.enabled',
    enabled: false,
    requires: ['fleet_updates.enabled'],
    restartRequired: true,
  },
];

afterEach(() => vi.restoreAllMocks());

test('reads feature flags, enforces dependencies, and saves with restart guidance', async () => {
  const get = vi
    .spyOn(HttpUtil, 'get')
    .mockResolvedValue(new Msg(true, '', { items, restartRequired: true }));
  const put = vi
    .spyOn(HttpUtil, 'put')
    .mockImplementation(async () => new Msg(true, '', { items, restartRequired: true }));
  renderWithProviders(
    <MemoryRouter>
      <CatxFeaturesTab />
    </MemoryRouter>,
  );

  await waitFor(() => expect(get).toHaveBeenCalledOnce());
  expect(screen.getByText('Panel restart required')).toBeTruthy();
  expect(screen.getByText('Analytics').tagName).not.toBe('CODE');
  expect(
    screen.getByText(
      'Collect and show metadata-only activity, DNS, session, and traffic insights.',
    ),
  ).toBeTruthy();
  expect(screen.getByText('analytics.enabled').tagName).toBe('CODE');
  const switches = screen.getAllByRole('switch');
  expect(switches[1].getAttribute('disabled')).not.toBeNull();
  fireEvent.click(switches[0]);
  fireEvent.click(switches[1]);
  fireEvent.click(screen.getByRole('button', { name: 'Save feature settings' }));

  await waitFor(() => expect(put).toHaveBeenCalledOnce());
  expect(put.mock.calls[0][1]).toMatchObject({
    flags: { 'analytics.enabled': true, 'dns_intelligence.enabled': true },
  });
});

test('uses localized feature names in dependency warnings', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue(
    new Msg(true, '', {
      items: items.map((item) =>
        item.key === 'dns_intelligence.enabled' ? { ...item, enabled: true } : item,
      ),
      restartRequired: true,
    }),
  );
  renderWithProviders(
    <MemoryRouter>
      <CatxFeaturesTab />
    </MemoryRouter>,
  );

  const message = 'Enable Analytics before enabling DNS intelligence.';
  await screen.findAllByText(message);
  const dependencyAlert = screen
    .getAllByRole('alert')
    .find((alert) => alert.textContent?.includes(message));
  expect(dependencyAlert).toBeTruthy();
  expect(dependencyAlert?.textContent).not.toContain('analytics.enabled');
});

test('keeps webhook mutation controls hidden when the feature entrypoint is disabled', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue(
    new Msg(false, 'Request failed with status 404', null, 404),
  );
  renderWithProviders(
    <MemoryRouter>
      <WebhooksPage />
    </MemoryRouter>,
  );
  await waitFor(() => expect(screen.getByText(/Webhooks is disabled/i)).toBeTruthy());
  expect(screen.queryByRole('button', { name: 'Create' })).toBeNull();
});
