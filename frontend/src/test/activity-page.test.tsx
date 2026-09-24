import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { afterEach, expect, test, vi } from 'vitest';

import ActivityPage from '@/pages/activity/ActivityPage';
import { HttpUtil } from '@/utils';

afterEach(() => vi.restoreAllMocks());

test('shows the analytics-disabled state without rendering activity data', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue({
    success: true,
    msg: '',
    obj: { enabled: false, items: [], page: 1, pageSize: 25, total: 0, from: 1, to: 2 },
  });
  render(
    <MemoryRouter initialEntries={['/activity?email=alice@example.com']}>
      <ActivityPage />
    </MemoryRouter>,
  );
  await waitFor(() => expect(screen.getByText(/Analytics is disabled/)).toBeTruthy());
  expect(screen.queryByText('example.com')).toBeNull();
});
