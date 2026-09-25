import { describe, expect, it } from 'vitest';

import { forkApiSections, forkNavigationItems, forkRoutes } from '@/forkext/registry';

describe('fork registries', () => {
  it('registers the client activity page through the fork boundary', () => {
    expect(forkRoutes).toHaveLength(2);
    expect(forkNavigationItems).toEqual([
      { key: 'client-activity', label: 'Client activity', path: '/activity' },
      { key: 'policy-engine', label: 'Policy engine', path: '/policies' },
    ]);
    expect(forkApiSections).toEqual([]);
  });
});
