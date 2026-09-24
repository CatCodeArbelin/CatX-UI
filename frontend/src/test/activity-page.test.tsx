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

test('renders DNS enrichment provenance confidence and ambiguity', async () => {
  vi.spyOn(HttpUtil, 'get').mockResolvedValue({
    success: true,
    msg: '',
    obj: {
      enabled: true,
      items: [
        {
          id: 1,
          observedAt: 2,
          domain: 'video.example',
          service: 'Example Video',
          category: 'video/streaming',
          classificationSource: 'domain_catalog',
          classificationProvenance: 'derived',
          classificationConfidence: 0.35,
          classificationLevel: 'low',
          classificationConflict: true,
          classificationReason: 'conflicting metadata',
          source: 'access_log',
          provenance: 'observed',
          confidence: 0.5,
        },
      ],
      page: 1,
      pageSize: 25,
      total: 1,
      from: 1,
      to: 3,
    },
  } as never);
  render(
    <MemoryRouter initialEntries={['/activity?email=alice@example.com']}>
      <ActivityPage />
    </MemoryRouter>,
  );
  await waitFor(() => expect(screen.getByText(/Example Video/)).toBeTruthy());
  expect(screen.getByText(/video\/streaming/)).toBeTruthy();
  expect(screen.getByText(/conflicting metadata/)).toBeTruthy();
  expect(screen.getByText('DNS observations')).toBeTruthy();
});
