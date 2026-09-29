import { act, fireEvent, screen } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router';
import { afterEach, expect, test, vi } from 'vitest';

import AppSidebar from '@/layouts/AppSidebar';
import { renderWithProviders } from './test-utils';

vi.mock('@/api/queries/useAllSettings', () => ({
  useAllSettings: () => ({ allSetting: {} }),
}));

afterEach(() => {
  localStorage.clear();
});

// rc-menu registers its items in a microtask after render; settle it inside act().
async function renderSidebar() {
  const view = renderWithProviders(
    <MemoryRouter>
      <AppSidebar />
    </MemoryRouter>,
  );
  await act(async () => {});
  return view;
}

function LocationProbe() {
  const { pathname } = useLocation();
  return <output data-testid="location-probe">{pathname}</output>;
}

test('keeps the sidebar expanded after pinning it from the header and restores the choice', async () => {
  const first = await renderSidebar();
  const sidebar = first.container.querySelector('.ant-layout-sider');
  const sidebarRoot = first.container.querySelector('.ant-sidebar');

  expect(sidebar?.classList.contains('ant-layout-sider-collapsed')).toBe(true);

  fireEvent.mouseEnter(sidebarRoot!);

  const pinButton = screen.getByRole('button', { name: 'Pin sidebar' });
  expect(pinButton.closest('.brand-actions')).not.toBeNull();

  fireEvent.click(pinButton);
  fireEvent.mouseLeave(sidebarRoot!);

  expect(sidebar?.classList.contains('ant-layout-sider-collapsed')).toBe(false);
  expect(sidebarRoot?.getAttribute('style')).toContain('--sider-rail: 220px');
  expect(localStorage.getItem('sidebar-pinned')).toBe('true');

  first.unmount();

  const second = await renderSidebar();
  const restoredSidebar = second.container.querySelector('.ant-layout-sider');
  const restoredSidebarRoot = second.container.querySelector('.ant-sidebar');

  expect(restoredSidebar?.classList.contains('ant-layout-sider-collapsed')).toBe(false);
  expect(restoredSidebarRoot?.getAttribute('style')).toContain('--sider-rail: 220px');
  expect(screen.getByRole('button', { name: 'Pin sidebar' })).not.toBeNull();
});

test('returns to the compact rail after unpinning', async () => {
  const view = await renderSidebar();
  const sidebar = view.container.querySelector('.ant-layout-sider');
  const sidebarRoot = view.container.querySelector('.ant-sidebar');

  fireEvent.mouseEnter(sidebarRoot!);
  fireEvent.click(screen.getByRole('button', { name: 'Pin sidebar' }));
  fireEvent.click(screen.getByRole('button', { name: 'Pin sidebar' }));
  fireEvent.mouseLeave(sidebarRoot!);

  expect(sidebar?.classList.contains('ant-layout-sider-collapsed')).toBe(true);
  expect(sidebarRoot?.getAttribute('style')).toContain('--sider-rail: 72px');
  expect(localStorage.getItem('sidebar-pinned')).toBe('false');
});

test('labels the palette shortcut with the modifier the platform actually uses', async () => {
  const view = await renderSidebar();
  const chip = view.container.querySelector('.sidebar-command-kbd');
  expect(chip?.textContent).toBe('CtrlK');
});

test('groups fork destinations into upstream navigation and removes Sponsors from primary menu', async () => {
  const view = await renderSidebar();
  const submenuLabels = Array.from(
    view.container.querySelectorAll('.ant-sidebar > .ant-layout-sider .ant-menu-submenu-title'),
  ).map((item) => item.textContent?.trim());
  expect(submenuLabels).toEqual(
    expect.arrayContaining(['Clients', 'Nodes', 'Routing', 'Operations']),
  );
  expect(screen.queryByText('Sponsors')).toBeNull();
});

test('keeps upstream root destinations as clickable submenu entries', async () => {
  const view = renderWithProviders(
    <MemoryRouter initialEntries={['/']}>
      <LocationProbe />
      <AppSidebar />
    </MemoryRouter>,
  );
  await act(async () => {});
  fireEvent.mouseEnter(view.container.querySelector('.ant-sidebar')!);

  for (const destination of ['Clients', 'Nodes', 'Routing']) {
    const groupTitle = Array.from(
      view.container.querySelectorAll('.ant-sidebar > .ant-layout-sider .ant-menu-submenu-title'),
    ).find((candidate) => candidate.textContent?.trim() === destination);
    expect(groupTitle, `${destination} group`).toBeTruthy();
    fireEvent.click(groupTitle!);
    const desktopItems = Array.from(
      view.container.querySelectorAll('.ant-sidebar > .ant-layout-sider .ant-menu-item'),
    );
    const item = desktopItems.find((candidate) => candidate.textContent?.trim() === destination);
    expect(item, destination).toBeTruthy();
    fireEvent.click(item!);
    expect(screen.getByTestId('location-probe').textContent).toBe(
      destination === 'Clients' ? '/clients' : destination === 'Nodes' ? '/nodes' : '/routing',
    );
  }
});
