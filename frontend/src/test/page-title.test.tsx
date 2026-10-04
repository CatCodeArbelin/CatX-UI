import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, useNavigate } from 'react-router';
import { expect, test } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { usePageTitle } from '@/hooks/usePageTitle';

function PageTitleHarness() {
  usePageTitle();
  const navigate = useNavigate();
  return <button onClick={() => navigate('/catx/sponsors/manage')}>Manage sponsors</button>;
}

test('sets the management title on direct load', async () => {
  render(
    <MemoryRouter initialEntries={['/catx/sponsors/manage']}>
      <PageTitleHarness />
    </MemoryRouter>,
  );
  await waitFor(() => expect(document.title).toBe('localhost - Sponsors management'));
});

test('sets the management title after navigation from the public CatX page', async () => {
  render(
    <MemoryRouter initialEntries={['/catx/sponsors']}>
      <PageTitleHarness />
    </MemoryRouter>,
  );
  fireEvent.click(screen.getByRole('button', { name: 'Manage sponsors' }));
  await waitFor(() => expect(document.title).toBe('localhost - Sponsors management'));
});

test('keeps the management title localized in the acceptance locales', () => {
  for (const locale of ['en-US', 'ru-RU', 'fa-IR']) {
    const messages = JSON.parse(
      readFileSync(resolve(process.cwd(), '../internal/web/translation', `${locale}.json`), 'utf8'),
    ) as { pages: { sponsors: { managementTitle?: string } } };
    expect(messages.pages.sponsors.managementTitle, locale).toBeTruthy();
  }
});
