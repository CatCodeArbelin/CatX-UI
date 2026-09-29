import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { afterEach, expect, test, vi } from 'vitest';

import PolicyPage from '@/pages/policy/PolicyPage';
import { i18n } from '@/i18n/react';
import { HttpUtil } from '@/utils';

afterEach(() => vi.restoreAllMocks());

const emptyCollections = { success: true, msg: '', obj: { enabled: true, items: [] } };

test('renders policy empty state and simulator controls', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue(emptyCollections as never);
  render(<PolicyPage />, { wrapper: ({ children }) => <MemoryRouter>{children}</MemoryRouter> });
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
  render(<PolicyPage />, { wrapper: ({ children }) => <MemoryRouter>{children}</MemoryRouter> });
  await waitFor(() => expect(screen.getByText(i18n.t('fork.policy.labels.disabled'))).toBeTruthy());
  expect(screen.getByRole('button', { name: i18n.t('fork.policy.newPolicy') })).toHaveProperty(
    'disabled',
    true,
  );
});
