import type { RouteObject } from 'react-router';

import type { Section } from '../pages/api-docs/endpoints.ts';

export interface ForkNavigationItem {
  key: string;
  label: string;
  path: string;
}

// Empty until a fork feature owns a page. Keeping the descriptor boundary
// here prevents feature pages from being mixed into the upstream shell.
export const forkRoutes: readonly RouteObject[] = [
  {
    path: '/activity',
    lazy: async () => ({ Component: (await import('../pages/activity/ActivityPage')).default }),
  },
];
export const forkNavigationItems: readonly ForkNavigationItem[] = [
  { key: 'client-activity', label: 'Client activity', path: '/activity' },
];
export const forkApiSections: readonly Section[] = [];
