import { screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { afterEach, expect, test, vi } from 'vitest';

import AuditPage from '@/pages/audit/AuditPage';
import { HttpUtil } from '@/utils';
import { renderWithProviders } from './test-utils';

afterEach(() => vi.restoreAllMocks());

test('renders an audit error state instead of an empty state when the request fails', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue({
    success: false,
    msg: 'backend failure',
    obj: undefined,
  } as never);
  renderWithProviders(
    <MemoryRouter>
      <AuditPage />
    </MemoryRouter>,
  );
  await waitFor(() => expect(screen.getByText('Audit events could not be loaded.')).toBeTruthy());
  expect(screen.queryByText('No audit events')).toBeNull();
});

test('renders the empty state for a successful empty audit response', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue({
    success: true,
    msg: '',
    obj: { items: [] },
  } as never);
  renderWithProviders(
    <MemoryRouter>
      <AuditPage />
    </MemoryRouter>,
  );
  await waitFor(() => expect(screen.getByText('No audit events')).toBeTruthy());
});
