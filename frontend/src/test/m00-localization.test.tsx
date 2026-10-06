/** @vitest-environment jsdom */
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';

import enUS from '../../../internal/web/translation/en-US.json';
import ruRU from '../../../internal/web/translation/ru-RU.json';
import CatxFeaturesTab from '@/pages/settings/CatxFeaturesTab';
import { ForkModuleTitle } from '@/components/fork/ForkModuleMaturity';
import { i18n } from '@/i18n/react';
import { HttpUtil, Msg } from '@/utils';
import { renderWithProviders } from './test-utils';

const m00Keys = [
  'fork.settings.title',
  'fork.settings.intro',
  'fork.settings.save',
  'fork.settings.saved',
  'fork.settings.loadFailed',
  'fork.settings.saveFailed',
  'fork.settings.restartRequired',
  'fork.settings.restartRequiredDescription',
  'fork.settings.dependencyWarning',
  'fork.settings.runtimeState',
  'fork.settings.desiredState',
  'fork.settings.noFeatures',
  'fork.settings.unknownFeature',
  'fork.settings.unknownFeatureDetails',
  'fork.settings.dependencyRequired',
  'fork.settings.features.analytics.name',
  'fork.settings.features.analytics.details',
  'fork.settings.features.dnsIntelligence.name',
  'fork.settings.features.dnsIntelligence.details',
  'fork.settings.features.policies.name',
  'fork.settings.features.policies.details',
  'fork.settings.features.trafficControl.name',
  'fork.settings.features.trafficControl.details',
  'fork.settings.features.riskIntelligence.name',
  'fork.settings.features.riskIntelligence.details',
  'fork.settings.features.audit.name',
  'fork.settings.features.audit.details',
  'fork.settings.features.clientPortal.name',
  'fork.settings.features.clientPortal.details',
  'fork.settings.features.fleetUpdates.name',
  'fork.settings.features.fleetUpdates.details',
  'fork.settings.features.sponsors.name',
  'fork.settings.features.sponsors.details',
  'fork.settings.features.productionFleetUpdates.name',
  'fork.settings.features.productionFleetUpdates.details',
  'fork.settings.states.feature_off',
  'fork.settings.states.active',
  'fork.settings.states.initializing',
  'fork.settings.states.restart_required',
  'fork.settings.states.error',
  'fork.settings.states.unknown',
  'fork.maturity.modules.m00',
  'fork.maturity.explanation',
  'fork.maturity.tooltip',
  'fork.maturity.accessible',
] as const;

const productionStates = [
  'feature_off',
  'active',
  'initializing',
  'restart_required',
  'error',
] as const;

const localizationItems = [
  {
    key: 'analytics.enabled',
    enabled: true,
    active: true,
    state: 'active',
    restartRequired: false,
  },
  {
    key: 'policies.enabled',
    enabled: false,
    active: false,
    state: 'feature_off',
    restartRequired: false,
  },
  {
    key: 'traffic_control.enabled',
    enabled: true,
    active: false,
    state: 'initializing',
    restartRequired: false,
  },
  {
    key: 'audit.enabled',
    enabled: true,
    active: false,
    state: 'restart_required',
    restartRequired: true,
  },
  {
    key: 'self_service.enabled',
    enabled: true,
    active: false,
    state: 'error',
    restartRequired: false,
  },
];

function valueAt(root: unknown, key: string): unknown {
  return key.split('.').reduce<unknown>((value, segment) => {
    if (!value || typeof value !== 'object') return undefined;
    return (value as Record<string, unknown>)[segment];
  }, root);
}

function placeholders(value: unknown): string[] {
  return [...String(value).matchAll(/\{[^{}]+\}/g)].map(([match]) => match).sort();
}

describe('M00 en-US and ru-RU localization', () => {
  beforeAll(() => {
    i18n.addResourceBundle('ru-RU', 'translation', ruRU, true, true);
  });

  afterEach(async () => {
    vi.restoreAllMocks();
    await i18n.changeLanguage('en-US');
  });

  it('contains every M00 key in both qualification locales with matching types and placeholders', () => {
    for (const key of m00Keys) {
      const english = valueAt(enUS, key);
      const russian = valueAt(ruRU, key);
      expect(english, `en-US:${key}`).toBeDefined();
      expect(russian, `ru-RU:${key}`).toBeDefined();
      expect(typeof russian, `ru-RU:${key}`).toBe(typeof english);
      expect(placeholders(russian), `ru-RU:${key}`).toEqual(placeholders(english));
      if (typeof russian === 'string') expect(russian.trim(), `ru-RU:${key}`).not.toBe('');
    }
  });

  it('localizes every production runtime state without changing its machine value', () => {
    for (const state of productionStates) {
      const english = valueAt(enUS, `fork.settings.states.${state}`);
      const russian = valueAt(ruRU, `fork.settings.states.${state}`);
      expect(english).not.toBe(state);
      expect(russian).not.toBe(state);
      expect(english).not.toBe(russian);
    }
  });

  it('switches the M00 surface to Russian while preserving desired/active semantics', async () => {
    const get = vi
      .spyOn(HttpUtil, 'get')
      .mockResolvedValue(new Msg(true, '', { items: localizationItems, restartRequired: true }));
    renderWithProviders(
      <MemoryRouter>
        <CatxFeaturesTab />
      </MemoryRouter>,
    );

    await waitFor(() => expect(get).toHaveBeenCalledOnce());
    expect(screen.getAllByText('Desired state')).toHaveLength(localizationItems.length);
    expect(screen.getAllByText(/Active runtime state:/)).toHaveLength(localizationItems.length);
    expect(screen.getByText('Activation error')).toBeTruthy();
    for (const rawState of productionStates) {
      expect(screen.queryByText(rawState, { exact: true })).toBeNull();
    }

    const enabledBeforeLocaleChange = (
      screen.getAllByRole('switch')[0] as HTMLButtonElement
    ).getAttribute('aria-checked');
    await i18n.changeLanguage('ru-RU');

    expect(screen.getAllByText('Желаемое состояние')).toHaveLength(localizationItems.length);
    expect(screen.getAllByText(/Фактическое состояние:/)).toHaveLength(localizationItems.length);
    expect(screen.getByText('Ошибка активации')).toBeTruthy();
    expect(screen.getByText('Аналитика рисков')).toBeTruthy();
    expect(screen.getAllByRole('switch')[0].getAttribute('aria-checked')).toBe(
      enabledBeforeLocaleChange,
    );
    for (const rawState of productionStates) {
      expect(screen.queryByText(rawState, { exact: true })).toBeNull();
    }
  });

  it('uses localized safe fallbacks for unexpected API values and an empty list', async () => {
    const get = vi.spyOn(HttpUtil, 'get');
    get.mockResolvedValueOnce(
      new Msg(true, '', {
        items: [
          {
            key: 'future.enabled',
            enabled: true,
            active: false,
            state: 'future_state',
            restartRequired: false,
          },
        ],
        restartRequired: false,
      }),
    );
    const view = renderWithProviders(
      <MemoryRouter>
        <CatxFeaturesTab />
      </MemoryRouter>,
    );
    await waitFor(() => expect(get).toHaveBeenCalledOnce());
    expect(screen.getByText('Unknown feature')).toBeTruthy();
    expect(screen.getByText('This feature is not recognized by this panel version.')).toBeTruthy();
    expect(screen.getByText('Unknown state')).toBeTruthy();
    expect(screen.queryByText('future.enabled', { exact: true })).toBeNull();
    expect(screen.queryByText('future_state', { exact: true })).toBeNull();

    view.unmount();
    get.mockResolvedValueOnce(new Msg(true, '', { items: [], restartRequired: false }));
    renderWithProviders(
      <MemoryRouter>
        <CatxFeaturesTab />
      </MemoryRouter>,
    );
    await waitFor(() => expect(get).toHaveBeenCalledTimes(2));
    expect(screen.getByText('No CatX features are available.')).toBeTruthy();
  });

  it('keeps normal M00 copy behind i18n and preserves the exact DEVS marker', async () => {
    const source = readFileSync(
      resolve(process.cwd(), 'src/pages/settings/CatxFeaturesTab.tsx'),
      'utf8',
    );
    for (const literal of [
      'Active',
      'Disabled',
      'Error',
      'Restart required',
      'Save feature settings',
      'Initializing',
    ]) {
      expect(source).not.toMatch(new RegExp(`>\\s*${literal}\\s*<`));
    }
    expect(source).not.toContain('featureCopy[key]?.nameKey || key');
    expect(source).not.toContain('featureCopy[item.key]?.detailsKey || item.key');
    expect(source).toContain("'fork.settings.unknownFeature'");
    expect(source).toContain("'fork.settings.unknownFeatureDetails'");
    expect(source).toContain("'fork.settings.states.unknown'");

    await i18n.changeLanguage('en-US');
    const view = renderWithProviders(<ForkModuleTitle moduleId="M00" title="CatX-UI Features" />);
    expect(screen.getByText('DEVS')).toBeTruthy();
    expect(screen.getByRole('status').getAttribute('aria-label')).toContain(
      enUS.fork.maturity.explanation,
    );
    view.unmount();

    await i18n.changeLanguage('ru-RU');
    renderWithProviders(<ForkModuleTitle moduleId="M00" title="Функции CatX-UI" />);
    expect(screen.getByText('DEVS')).toBeTruthy();
    expect(screen.getByRole('status').getAttribute('aria-label')).toContain(
      ruRU.fork.maturity.explanation,
    );
    expect(screen.queryByText('РАЗРАБОТКА')).toBeNull();
  });
});
