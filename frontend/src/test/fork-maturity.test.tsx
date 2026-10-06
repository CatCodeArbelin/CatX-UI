/** @vitest-environment jsdom */
import { screen } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import enUS from '../../../internal/web/translation/en-US.json';
import ruRU from '../../../internal/web/translation/ru-RU.json';
import { ForkModuleTitle } from '@/components/fork/ForkModuleMaturity';
import {
  forkModuleIds,
  forkModuleRegistry,
  forkNavigationItems,
  forkRouteModules,
} from '@/forkext/registry';
import { i18n } from '@/i18n/react';
import { renderWithProviders } from './test-utils';

const allModuleIds = [
  'M00',
  'M01',
  'M02',
  'M03',
  'M04',
  'M05',
  'M06',
  'M07',
  'M08',
  'M09',
  'M10',
  'M11',
] as const;

beforeAll(() => {
  i18n.addResourceBundle('ru-RU', 'translation', ruRU, true, true);
});

afterAll(async () => {
  await i18n.changeLanguage('en-US');
});

describe('CatX module maturity', () => {
  it('registers exactly M00-M11 and keeps every current module at DEVS', () => {
    expect(forkModuleIds).toEqual(allModuleIds);
    expect(Object.keys(forkModuleRegistry)).toEqual(allModuleIds);
    expect(Object.values(forkModuleRegistry).map((module) => module.maturity)).toEqual(
      allModuleIds.map(() => 'devs'),
    );
    expect(Object.values(forkModuleRegistry).map((module) => module.maturity)).not.toContain(
      'ready',
    );
  });

  it('keeps grouped and legacy routes on one module maturity source', () => {
    expect(forkRouteModules['/audit']).toBe('M07');
    expect(forkRouteModules['/webhooks']).toBe('M07');
    expect(forkRouteModules['/portal']).toBe('M08');
    expect(forkRouteModules['/portal-access']).toBe('M08');
    expect(forkRouteModules['/catx/sponsors']).toBe('M11');
    expect(forkRouteModules['/sponsors']).toBe('M11');
    expect(
      forkNavigationItems.every((item) => forkModuleRegistry[item.moduleId].maturity === 'devs'),
    ).toBe(true);
  });

  it('renders a localized module name with an accessible, untranslated DEVS marker', async () => {
    await i18n.changeLanguage('en-US');
    const view = renderWithProviders(<ForkModuleTitle moduleId="M03" title="Policy Engine" />);
    expect(screen.getByText('Policy Engine')).toBeTruthy();
    expect(screen.getByText('DEVS')).toBeTruthy();
    expect(screen.getByRole('status').getAttribute('aria-label')).toContain('DEVS');
    expect(screen.getByRole('status').getAttribute('aria-label')).toContain(
      enUS.fork.maturity.explanation,
    );
    view.unmount();

    await i18n.changeLanguage('ru-RU');
    renderWithProviders(<ForkModuleTitle moduleId="M03" title="Движок политик" />);
    expect(screen.getByText('Движок политик')).toBeTruthy();
    expect(screen.getByText('DEVS')).toBeTruthy();
    expect(screen.getByRole('status').getAttribute('aria-label')).toContain(
      ruRU.fork.maturity.explanation,
    );
    expect(screen.queryByText('РАЗРАБОТКА')).toBeNull();
  });

  it('keeps the maturity independent from runtime state labels', () => {
    for (const runtimeState of ['active', 'feature_off', 'error', 'restart_required']) {
      expect(runtimeState).toBeTypeOf('string');
      expect(forkModuleRegistry.M03.maturity).toBe('devs');
    }
  });

  it('keeps the required English and Russian explanation strings in the reviewed catalogs', () => {
    expect(enUS.fork.maturity.explanation).toBe(
      'This module is implemented but has not yet completed release qualification.',
    );
    expect(ruRU.fork.maturity.explanation).toBe(
      'Модуль реализован, но ещё не прошёл квалификацию перед выпуском.',
    );
  });
});
