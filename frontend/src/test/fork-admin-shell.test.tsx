import { act, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { test, expect, vi } from 'vitest';

import ForkAdminPageShell from '@/components/fork/ForkAdminPageShell';
import { renderWithProviders } from './test-utils';

vi.mock('@/api/queries/useAllSettings', () => ({
  useAllSettings: () => ({ allSetting: {} }),
}));

test('renders fork admin content inside the standard themed shell', async () => {
  const view = renderWithProviders(
    <MemoryRouter>
      <ForkAdminPageShell pageClass="activity-page">
        <span>Admin content</span>
      </ForkAdminPageShell>
    </MemoryRouter>,
  );
  await act(async () => {});

  const shell = screen.getByTestId('fork-admin-shell');
  expect(shell.classList.contains('fork-admin-page')).toBe(true);
  expect(shell.classList.contains('activity-page')).toBe(true);
  expect(view.container.querySelector('.ant-sidebar')).not.toBeNull();
  expect(view.container.querySelector('.content-shell .content-area')).not.toBeNull();
  expect(screen.getByText('Admin content')).toBeTruthy();
});
