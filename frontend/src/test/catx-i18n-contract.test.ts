import { readFileSync, readdirSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';
import { forkApiSections, forkNavigationItems } from '@/forkext/registry';
import { directionForLanguage } from '@/i18n/direction';

const translationDir = resolve(process.cwd(), '../internal/web/translation');
const locales = [
  'ar-EG',
  'en-US',
  'es-ES',
  'fa-IR',
  'id-ID',
  'ja-JP',
  'pt-BR',
  'ru-RU',
  'tr-TR',
  'uk-UA',
  'vi-VN',
  'zh-CN',
  'zh-TW',
];

const englishIdenticalTechnicalKeys = new Set([
  'fork.portal.subjectTypePlaceholder',
  'fork.common.labels.sni',
  'fork.policy.clientPlaceholder',
  'fork.policy.groupPlaceholder',
  'fork.policy.domainPlaceholder',
  'fork.policy.ipPlaceholder',
  'fork.policy.categoryPlaceholder',
  'fork.policy.scopes.dns',
  'fork.policy.scopes.qos',
  'fork.policy.labels.ipCidr',
  'fork.activity.labels.dns',
]);

// These CatX v0.3 keys are intentionally carried as explicit English fallback
// copy in locales whose reviewed translations have not adopted the new surface
// yet. EN/RU remain fully localized; the fallback is tested as a contract.
const englishFallbackKeyPrefixes = [
  'fork.common.labels.trafficControl',
  'fork.common.labels.quotaWindow',
  'fork.common.labels.seconds',
  'fork.common.labels.bytes',
  'fork.common.labels.activeSpeed',
  'fork.common.labels.throttleSpeed',
  'fork.common.labels.uploadSpeed',
  'fork.common.labels.downloadSpeed',
  'fork.common.labels.lifecycle',
  'fork.common.labels.enforcement',
  'fork.common.labels.used',
  'fork.common.labels.remaining',
  'fork.common.labels.featureUnavailable',
  'fork.common.labels.windowResetReason',
  'fork.common.labels.quotaReachedReason',
  'fork.common.labels.upstreamDisabledReason',
  'fork.common.labels.attributionUnsupportedReason',
  'fork.common.labels.rateUnsupported',
  'fork.common.labels.actionFailed',
  'fork.common.labels.trafficControlUnconfigured',
  'fork.common.labels.configureTrafficControl',
  'fork.fleet.emptyTitle',
  'fork.fleet.emptyDescription',
  'fork.fleet.openNodes',
  'fork.fleet.nodeStates.',
  'fork.fleetUpdate.abortConfirm',
  'fork.fleetUpdate.states.',
  'fork.fleetUpdate.reasons.',
  'fork.fleetUpdate.dispatchStates.',
  'fork.fleetUpdate.resultStates.',
  'fork.audit.outcomes.',
  'fork.policy.labels.startsAt',
  'fork.policy.labels.expiresAt',
  'fork.policy.labels.startTime',
  'fork.policy.labels.endTime',
  'fork.policy.labels.action',
  'fork.policy.labels.allow',
  'fork.policy.labels.deny',
  'fork.policy.labels.services',
  'fork.policy.labels.destinations',
  'fork.policy.labels.resolver',
  'fork.policy.labels.scheduleRef',
  'fork.policy.labels.quarantine',
  'fork.policy.labels.quarantineAllowlist',
  'fork.policy.labels.advancedJson',
  'fork.policy.runtimeApplyFailed',
];

function isEnglishFallbackKey(key: string): boolean {
  return (
    key === 'fork.common.local' ||
    key.startsWith('fork.common.states.') ||
    key === 'fork.activity.labels.nodes' ||
    key === 'fork.settings.runtimeState' ||
    englishFallbackKeyPrefixes.some((prefix) => key === prefix || key.startsWith(prefix))
  );
}

type Json = string | number | boolean | null | Json[] | { [key: string]: Json };

function duplicateJsonKeys(text: string): string[] {
  let index = 0;
  const duplicates: string[] = [];
  const skipWhitespace = () => {
    while (/\s/.test(text[index] || '')) index += 1;
  };
  const parseString = (): string => {
    const start = index;
    index += 1;
    while (index < text.length) {
      if (text[index] === '\\') {
        index += 2;
        continue;
      }
      if (text[index] === '"') {
        index += 1;
        return JSON.parse(text.slice(start, index)) as string;
      }
      index += 1;
    }
    throw new Error('unterminated JSON string');
  };
  const parseValue = (path: string): void => {
    skipWhitespace();
    if (text[index] === '{') {
      index += 1;
      skipWhitespace();
      const seen = new Set<string>();
      while (text[index] !== '}') {
        const key = parseString();
        if (seen.has(key)) duplicates.push(path ? `${path}.${key}` : key);
        seen.add(key);
        skipWhitespace();
        if (text[index] !== ':') throw new Error('malformed JSON object');
        index += 1;
        parseValue(path ? `${path}.${key}` : key);
        skipWhitespace();
        if (text[index] === ',') {
          index += 1;
          skipWhitespace();
        } else if (text[index] !== '}') {
          throw new Error('malformed JSON object separator');
        }
      }
      index += 1;
      return;
    }
    if (text[index] === '[') {
      index += 1;
      skipWhitespace();
      while (text[index] !== ']') {
        parseValue(path);
        skipWhitespace();
        if (text[index] === ',') {
          index += 1;
          skipWhitespace();
        } else if (text[index] !== ']') {
          throw new Error('malformed JSON array separator');
        }
      }
      index += 1;
      return;
    }
    if (text[index] === '"') {
      parseString();
      return;
    }
    while (index < text.length && !/\s|[,\]}]/.test(text[index])) index += 1;
  };
  parseValue('');
  return duplicates;
}

function flatten(value: Json, prefix = ''): Array<[string, Json]> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return [[prefix, value]];
  return Object.entries(value).flatMap(([key, child]) =>
    flatten(child, prefix ? `${prefix}.${key}` : key),
  );
}

function forkEntries(locale: string) {
  const value = JSON.parse(readFileSync(resolve(translationDir, `${locale}.json`), 'utf8')) as Json;
  return new Map(flatten(value).filter(([key]) => key === 'fork' || key.startsWith('fork.')));
}

function placeholders(value: Json): string[] {
  return [...String(value).matchAll(/\{[^{}]+\}/g)].map(([match]) => match).sort();
}

describe('CatX frontend i18n contract', () => {
  it('rejects duplicate JSON object keys in every locale catalog', () => {
    for (const locale of locales) {
      const raw = readFileSync(resolve(translationDir, `${locale}.json`), 'utf8');
      expect(duplicateJsonKeys(raw), locale).toEqual([]);
    }
  });

  it('keeps the CatX key tree, types, and placeholders identical across locales', () => {
    const english = forkEntries('en-US');
    for (const locale of locales) {
      const current = forkEntries(locale);
      if (locale === 'en-US' || locale === 'ru-RU') {
        expect([...current.keys()], locale).toEqual([...english.keys()]);
      } else {
        // English is the runtime fallback for newly introduced fork copy.
        // Other locales may adopt the same key in a later translation pass.
        expect(
          [...current.keys()].every((key) => english.has(key)),
          locale,
        ).toBe(true);
      }
      for (const [key, value] of english) {
        if (!current.has(key)) continue;
        expect(typeof current.get(key), `${locale}:${key}`).toBe(typeof value);
        expect(placeholders(current.get(key)!), `${locale}:${key}`).toEqual(placeholders(value));
        if (typeof current.get(key) === 'string')
          expect(current.get(key), `${locale}:${key}`).not.toBe('');
      }
    }
  });

  it('keeps fork navigation fully localized', () => {
    expect(forkNavigationItems.every((item) => item.labelKey.startsWith('fork.'))).toBe(true);
    expect(forkNavigationItems).toHaveLength(8);
    expect(new Set(forkNavigationItems.map((item) => item.icon)).size).toBe(8);
    expect(forkNavigationItems.filter((item) => item.group === 'operations')).toHaveLength(3);
    expect(forkApiSections.every((section) => section.translationKey?.startsWith('fork.'))).toBe(
      true,
    );
  });

  it('supports RTL only for the designated locales', () => {
    expect(directionForLanguage('ar-EG')).toBe('rtl');
    expect(directionForLanguage('fa-IR')).toBe('rtl');
    expect(directionForLanguage('en-US')).toBe('ltr');
    expect(directionForLanguage('ru-RU')).toBe('ltr');
  });

  it('keeps CatX visible strings out of page source fallback paths', () => {
    const pages = readdirSync(resolve(process.cwd(), 'src/pages'), { recursive: true })
      .filter((file): file is string => typeof file === 'string')
      .filter(
        (file) =>
          /\.(tsx|ts)$/.test(file) && /^(activity|policy|audit|fleet|portal)[/\\]/.test(file),
      );
    const source = pages
      .map((file) => readFileSync(resolve(process.cwd(), 'src/pages', file), 'utf8'))
      .join('\n');
    for (const literal of [
      'Policy engine',
      'No policies yet.',
      'Client activity & DNS intelligence',
      'Fleet updates',
      'Portal access',
      'Client portal',
    ]) {
      expect(source).not.toContain(literal);
    }
    expect(source).not.toMatch(/t\(\s*['"]fork\.[^)]*,\s*['"]/);
  });

  it('rejects untranslated English CatX values outside the reviewed allowlists', () => {
    const english = forkEntries('en-US');
    for (const locale of locales.filter((item) => item !== 'en-US')) {
      const current = forkEntries(locale);
      const identical = [...english].filter(
        ([key, value]) =>
          current.get(key) === value &&
          !englishIdenticalTechnicalKeys.has(key) &&
          !isEnglishFallbackKey(key),
      );
      expect(identical, `${locale} has untranslated CatX values`).toEqual([]);
      expect(
        [...english]
          .filter(
            ([key, value]) =>
              current.has(key) &&
              current.get(key) === value &&
              (englishIdenticalTechnicalKeys.has(key) || isEnglishFallbackKey(key)),
          )
          .map(([key]) => key)
          .sort(),
        `${locale} allowed English-identical count`,
      ).toEqual(
        [...english]
          .filter(
            ([key]) =>
              current.has(key) &&
              current.get(key) === english.get(key) &&
              (englishIdenticalTechnicalKeys.has(key) || isEnglishFallbackKey(key)),
          )
          .map(([key]) => key)
          .sort(),
      );
    }
  });

  it('guards the reviewed Russian CatX surface against known mixed-language regressions', () => {
    const ru = JSON.parse(readFileSync(resolve(translationDir, 'ru-RU.json'), 'utf8')) as {
      fork: Json;
    };
    const values = flatten(ru.fork)
      .filter(([, value]) => typeof value === 'string')
      .map(([, value]) => String(value));
    for (const broken of [
      'Парк Обновления',
      'Кампания Имя',
      'Канареечный count',
      'Пакет size',
      'Plan Пробный run',
      'Пробный run (Нет mutation)',
      'Confirm production Обновление',
      'Enable DNS Интеллект',
      'Risk Интеллект',
    ]) {
      expect(values).not.toContain(broken);
    }
  });

  it('keeps every CatX page on the shared page shell', () => {
    const expectedClasses = [
      'activity-page',
      'policy-page',
      'audit-page',
      'webhooks-page',
      'fleet-page',
      'fleet-update-page',
      'portal-admin-page',
      'portal-page',
    ];
    const pageFiles = [
      ['activity', 'ActivityPage.tsx'],
      ['policy', 'PolicyPage.tsx'],
      ['audit', 'AuditPage.tsx'],
      ['audit', 'WebhooksPage.tsx'],
      ['fleet', 'FleetPage.tsx'],
      ['fleet', 'FleetUpdatePage.tsx'],
      ['portal', 'PortalAdminPage.tsx'],
      ['portal', 'PortalPage.tsx'],
    ];
    const source = pageFiles
      .map(([directory, file]) =>
        readFileSync(resolve(process.cwd(), 'src/pages', directory, file), 'utf8'),
      )
      .join('\n');
    for (const name of expectedClasses) expect(source).toContain(name);
    for (const [directory, file] of pageFiles.slice(0, -1)) {
      expect(
        readFileSync(resolve(process.cwd(), 'src/pages', directory, file), 'utf8'),
        `${directory}/${file}`,
      ).toContain('ForkAdminPageShell');
    }
    const portalPage = readFileSync(
      resolve(process.cwd(), 'src/pages/portal/PortalPage.tsx'),
      'utf8',
    );
    expect(portalPage).not.toContain('ForkAdminPageShell');
    expect(portalPage).not.toContain('@/layouts/AppSidebar');
    const adminShell = readFileSync(
      resolve(process.cwd(), 'src/components/fork/ForkAdminPageShell.tsx'),
      'utf8',
    );
    expect(adminShell).toContain('ConfigProvider');
    expect(adminShell).toContain('AppSidebar');
    expect(adminShell).toContain('content-shell');
    expect(readFileSync(resolve(process.cwd(), 'src/styles/page-shell.css'), 'utf8')).toContain(
      'fork-admin-page',
    );
    expect(readFileSync(resolve(process.cwd(), 'src/styles/page-cards.css'), 'utf8')).toContain(
      'activity-page',
    );
  });

  it('keeps CatX narrow-layout and RTL safeguards in the shared styles', () => {
    const shell = readFileSync(resolve(process.cwd(), 'src/styles/page-shell.css'), 'utf8');
    expect(shell).toContain('.fork-admin-page .ant-table-wrapper');
    expect(shell).toContain('overflow-x: auto');
    expect(shell).toContain('.catx-technical-value');
    expect(shell).toContain('prefers-reduced-motion: reduce');
    expect(shell).toContain('.api-docs-page .swagger-ui .table-container');
  });

  it('keeps dangerous CatX actions confirmable and icon controls named', () => {
    const fleet = readFileSync(
      resolve(process.cwd(), 'src/pages/fleet/FleetUpdatePage.tsx'),
      'utf8',
    );
    const portal = readFileSync(resolve(process.cwd(), 'src/pages/portal/PortalPage.tsx'), 'utf8');
    const policy = readFileSync(resolve(process.cwd(), 'src/pages/policy/PolicyPage.tsx'), 'utf8');
    expect(fleet).toContain('<Popconfirm');
    expect(portal).toContain('Modal.confirm');
    expect(policy).toContain("aria-label={t('fork.policy.labels.deleteAssignment')}");
    expect(policy).toContain("aria-label={t('fork.policy.labels.deleteOverride')}");
    expect(policy).toContain("aria-label={t('fork.policy.labels.deleteSchedule')}");
  });
});
