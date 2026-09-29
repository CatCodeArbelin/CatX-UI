import { waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { afterEach, expect, test, vi } from 'vitest';

import PortalPage from '@/pages/portal/PortalPage';
import { HttpUtil, Msg } from '@/utils';
import { renderWithProviders } from './test-utils';

afterEach(() => vi.restoreAllMocks());

test('renders explicit self-service disabled state as informational UX', async () => {
  vi.spyOn(HttpUtil, 'get').mockImplementation(async (url: string) => {
    if (url.endsWith('/portal/me')) return new Msg(false, 'self-service is disabled');
    return new Msg(false, 'not requested');
  });

  const view = renderWithProviders(
    <MemoryRouter>
      <PortalPage />
    </MemoryRouter>,
  );

  await waitFor(() => expect(view.container.querySelector('.ant-alert-info')).not.toBeNull());
  expect(view.container.querySelector('.ant-alert-error')).toBeNull();
});
