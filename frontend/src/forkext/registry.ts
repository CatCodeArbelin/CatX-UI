import type { RouteObject } from 'react-router';

export interface ForkNavigationItem {
  key: string;
  label: string;
  path: string;
}

export const forkRoutes: readonly RouteObject[] = [
  {
    path: '/activity',
    lazy: async () => ({ Component: (await import('../pages/activity/ActivityPage')).default }),
  },
];
export const forkNavigationItems: readonly ForkNavigationItem[] = [
  { key: 'client-activity', label: 'Client activity', path: '/activity' },
];
export const forkApiSections = [];
