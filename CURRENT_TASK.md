# CURRENT TASK

## Work Package

`WP-9C — Accessibility, RTL, Responsive & Visual QA`

## Authorization and guardrails

WP-9B was merged into `develop` at `323e5028` after green Fork Verification
run `36508672070` for implementation SHA
`4a02a2edfefe48fa30a81acc9e47f521f3e2c8d1`. Work on branch
`feature/wp-9c-visual-accessibility-qa`. `main` must remain untouched.

Audit and correct only real frontend defects across Activity, Policy, Audit,
Webhooks, Fleet, Fleet Updates, Portal Admin, Client Portal,
navigation/sidebar, command palette, and API docs. Validate light, dark, and
ultra-dark themes; 375, 430, 768, and 1024+ pixel viewports; LTR; RTL
(`ar-EG`, `fa-IR`); and long-string locales (`ru-RU`, `uk-UA`, `es-ES`,
`pt-BR`, `tr-TR`). Preserve upstream architecture and behavior. Do not
modify backend, database, Xray, updater, release semantics, dependencies, or
`main`.

Use existing Ant Design/theme infrastructure, logical CSS properties, and
existing tokens. Do not create a second CatX theme or broadly rewrite
upstream CSS. Technical LTR values in RTL must remain readable. Dangerous
actions require clear confirmation. All visible CatX text remains localized
through `fork.*`.

## Required WP-9C behavior

- Check keyboard navigation, visible focus, labels, modal focus, ARIA names,
  semantic buttons/links, non-color status communication, reduced motion, and
  mobile table usability.
- Check overflow, clipped text/buttons, modal bounds, page shells/cards,
  spacing, RTL alignment, and mixed-direction IP, CIDR, domain, SNI, UUID,
  runId, version, and URL values.
- Perform translation-quality smoke review for update, rollback, abort, retry,
  delete, revoke, quarantine, dry run, unsupported, degraded, failed, and
  unknown. Change only semantic or obviously broken wording.
- Add narrowly scoped regression/accessibility tests for confirmed defects.

## Verification and delivery

Run frontend format, lint, typecheck, tests, i18n contracts, accessibility
tests, generated checks, `make verify`, and `make verify-fork`. Use canonical
Linux CI for unavailable Windows-native gates; do not modify dependencies to
work around local bindings.

Push the branch and require green Fork Verification on the exact final SHA.
Do not merge WP-9C. Do not publish a release.

## Authorized documents

- `docs/08_FRONTEND_UX.md`
- `docs/25_FRONTEND_I18N_CONTRACT.md`
- `docs/26_LOCALIZATION_GLOSSARY.md`
