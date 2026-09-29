export const RTL_LANGUAGES = new Set(['ar-EG', 'fa-IR']);

export type PanelDirection = 'ltr' | 'rtl';

export function directionForLanguage(language: string): PanelDirection {
  return RTL_LANGUAGES.has(language) ? 'rtl' : 'ltr';
}

export function applyDocumentDirection(language: string): PanelDirection {
  const direction = directionForLanguage(language);
  if (typeof document !== 'undefined') {
    document.documentElement.dir = direction;
    document.body.dir = direction;
  }
  return direction;
}
