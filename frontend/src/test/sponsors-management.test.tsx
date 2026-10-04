import { screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { beforeEach, expect, test, vi } from 'vitest';

import SponsorsManagementPage from '@/forkext/sponsors/SponsorsManagementPage';
import { HttpUtil, Msg } from '@/utils';
import { renderWithProviders } from './test-utils';

vi.mock('@/api/queries/useAllSettings', () => ({
  useAllSettings: () => ({ allSetting: {} }),
}));

const record = {
  id: 'local-one',
  enabled: true,
  name: 'Local One',
  priority: 1,
  slots: ['page'],
  destinationUrl: 'https://example.com/',
  title: { 'en-US': 'Local One' },
  text: { 'en-US': 'A local sponsor' },
  createdAt: '2026-10-04T00:00:00Z',
  updatedAt: '2026-10-04T00:00:00Z',
};

function mockSponsorReads(providerMode: 'local' | 'remote') {
  vi.mocked(HttpUtil.get).mockImplementation(async (url: string) => {
    if (url.endsWith('/settings')) {
      return new Msg(true, '', {
        enabled: true,
        configured: false,
        sourceUrl: '',
        contactUrl: '',
        providerMode,
      });
    }
    if (url.endsWith('/status')) {
      return new Msg(true, '', {
        enabled: true,
        providerMode,
        localSponsorCount: 1,
        activeSponsorCount: providerMode === 'local' ? 1 : 0,
        remoteProviderConfigured: providerMode === 'remote',
        cacheState: providerMode,
      });
    }
    return new Msg(true, '', [record]);
  });
}

beforeEach(() => {
  mockSponsorReads('local');
  vi.spyOn(HttpUtil, 'put').mockResolvedValue(new Msg(true, '', record));
});

test('renders local sponsor management and preview entry points', async () => {
  renderWithProviders(
    <MemoryRouter>
      <SponsorsManagementPage />
    </MemoryRouter>,
  );

  expect(await screen.findByText('Local One')).toBeTruthy();
  expect(screen.getByText('Provider status')).toBeTruthy();
  expect(screen.getByRole('button', { name: /Create/ }).getAttribute('disabled')).toBeNull();
  expect(HttpUtil.get).toHaveBeenCalled();
});

test('makes local CRUD controls read-only while remote provider is selected', async () => {
  mockSponsorReads('remote');
  renderWithProviders(
    <MemoryRouter>
      <SponsorsManagementPage />
    </MemoryRouter>,
  );

  expect(await screen.findByText('Local One')).toBeTruthy();
  expect(screen.getByRole('button', { name: /Create/ }).getAttribute('disabled')).not.toBeNull();
  expect(screen.getByText(/Remote mode is read-only/)).toBeTruthy();
});
