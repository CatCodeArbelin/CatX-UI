import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { afterEach, expect, test, vi } from 'vitest';

import PolicyPage from '@/pages/policy/PolicyPage';
import { HttpUtil } from '@/utils';

afterEach(() => vi.restoreAllMocks());

const emptyCollections = { success: true, msg: '', obj: { enabled: true, items: [] } };

test('renders policy empty state and simulator controls', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue(emptyCollections as never);
  render(<PolicyPage />, { wrapper: ({ children }) => <MemoryRouter>{children}</MemoryRouter> });
  await waitFor(() => expect(screen.getByText('No policies yet.')).toBeTruthy());
  fireEvent.click(screen.getByText('Simulator'));
  expect(screen.getByPlaceholderText('alice@example.test')).toBeTruthy();
});

test('renders disabled policy state without enabling management actions', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue({
    success: true,
    msg: '',
    obj: { enabled: false, items: [] },
  } as never);
  render(<PolicyPage />, { wrapper: ({ children }) => <MemoryRouter>{children}</MemoryRouter> });
  await waitFor(() => expect(screen.getByText(/Policies are disabled/)).toBeTruthy());
  expect(screen.getByRole('button', { name: /New policy/ })).toHaveProperty('disabled', true);
});
