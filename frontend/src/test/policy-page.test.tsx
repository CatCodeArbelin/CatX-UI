import { fireEvent, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { afterEach, expect, test, vi } from 'vitest';

import PolicyPage from '@/pages/policy/PolicyPage';
import { i18n } from '@/i18n/react';
import { HttpUtil } from '@/utils';
import { renderWithProviders } from './test-utils';

afterEach(() => vi.restoreAllMocks());

const emptyCollections = { success: true, msg: '', obj: { enabled: true, items: [] } };

test('renders policy empty state and simulator controls', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue(emptyCollections as never);
  renderWithProviders(
    <MemoryRouter>
      <PolicyPage />
    </MemoryRouter>,
  );
  await waitFor(() => expect(screen.getByText(i18n.t('fork.policy.noPolicies'))).toBeTruthy());
  fireEvent.click(screen.getByRole('tab', { name: i18n.t('fork.policy.simulator') }));
  expect(screen.getByPlaceholderText(i18n.t('fork.policy.clientPlaceholder'))).toBeTruthy();
});

test('renders disabled policy state without enabling management actions', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue({
    success: true,
    msg: '',
    obj: { enabled: false, items: [] },
  } as never);
  renderWithProviders(
    <MemoryRouter>
      <PolicyPage />
    </MemoryRouter>,
  );
  await waitFor(() => expect(screen.getByText(i18n.t('fork.policy.labels.disabled'))).toBeTruthy());
  expect(
    screen.getByRole('button', { name: new RegExp(i18n.t('fork.policy.newPolicy')) }),
  ).toHaveProperty('disabled', true);
});

test('renders a policy error state instead of empty collections when loading fails', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue({
    success: false,
    msg: 'backend failure',
    obj: undefined,
  } as never);
  renderWithProviders(
    <MemoryRouter>
      <PolicyPage />
    </MemoryRouter>,
  );
  await waitFor(() => expect(screen.getByText('Policies could not be loaded.')).toBeTruthy());
  expect(screen.queryByText(/No policies yet/i)).toBeNull();
});
