import { describe, expect, it } from 'vitest';

import { forkApiSections, forkNavigationItems, forkRoutes } from '@/forkext/registry';

describe('fork registries', () => {
  it('are empty in the foundation package', () => {
    expect(forkRoutes).toEqual([]);
    expect(forkNavigationItems).toEqual([]);
    expect(forkApiSections).toEqual([]);
  });
});
