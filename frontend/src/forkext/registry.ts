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
  {
    path: '/policies',
    lazy: async () => ({ Component: (await import('../pages/policy/PolicyPage')).default }),
  },
];
export const forkNavigationItems: readonly ForkNavigationItem[] = [
  { key: 'client-activity', label: 'Client activity', path: '/activity' },
  { key: 'policy-engine', label: 'Policy engine', path: '/policies' },
];
export const forkApiSections = [];
