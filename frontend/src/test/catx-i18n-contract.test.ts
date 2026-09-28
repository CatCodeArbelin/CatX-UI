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
    expect(forkNavigationItems).toHaveLength(7);
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
    expect(readFileSync(resolve(process.cwd(), 'src/styles/page-shell.css'), 'utf8')).toContain(
      'activity-page',
    );
    expect(readFileSync(resolve(process.cwd(), 'src/styles/page-cards.css'), 'utf8')).toContain(
      'activity-page',
    );
  });
});
