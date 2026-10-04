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

type Json = string | number | boolean | null | Json[] | { [key: string]: Json };

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
  it('keeps the CatX key tree, types, and placeholders identical across locales', () => {
    const english = forkEntries('en-US');
    for (const locale of locales) {
      const current = forkEntries(locale);
      expect([...current.keys()], locale).toEqual([...english.keys()]);
      for (const [key, value] of english) {
        expect(current.get(key), `${locale}:${key}`).toBeDefined();
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

  it('rejects untranslated English CatX values outside the technical allowlist', () => {
    const english = forkEntries('en-US');
    for (const locale of locales.filter((item) => item !== 'en-US')) {
      const current = forkEntries(locale);
      const identical = [...english].filter(
        ([key, value]) => current.get(key) === value && !englishIdenticalTechnicalKeys.has(key),
      );
      expect(identical, `${locale} has untranslated CatX values`).toEqual([]);
      expect(
        [...english]
          .filter(([key, value]) => current.get(key) === value)
          .map(([key]) => key)
          .sort(),
        `${locale} technical English-identical count`,
      ).toEqual([...englishIdenticalTechnicalKeys].sort());
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
