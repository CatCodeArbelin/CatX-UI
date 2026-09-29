# CURRENT TASK

## Work Package

`WP-9B — Full CatX Localization`

## Authorization and guardrails

WP-9A was merged into `develop` at `cc6b13aa` after green Fork Verification
run `36504762871`. Work on branch
`feature/wp-9b-full-localization`. `main` must remain untouched.

Translate only the fork-owned `fork.*` tree in the 12 non-English supported
locales. English (`en-US`) remains the source of truth. Preserve the exact
CatX key tree, value types, interpolation placeholders, go-i18n compatibility,
navigation/API-doc references, RTL behavior, and upstream translations.

Do not change frontend behavior, backend behavior, database/schema, Xray,
updater, release semantics, dependencies, or feature logic. Do not merge
WP-9B.

## Required WP-9B behavior

- Fully translate all 238 stabilized `fork.*` keys in `ar-EG`, `es-ES`,
  `fa-IR`, `id-ID`, `ja-JP`, `pt-BR`, `ru-RU`, `tr-TR`, `uk-UA`, `vi-VN`,
  `zh-CN`, and `zh-TW`.
- Add and use `docs/26_LOCALIZATION_GLOSSARY.md` for consistent operator and
  network terminology.
- Reject unintended English-identical non-English values with a narrow,
  justified technical-term allowlist.
- Preserve placeholder parity, dead-key strictness, reserved-key safety,
  locale shape parity, navigation/API-doc references, and RTL foundations.
- Review operational language for update, rollback, abort, retry, delete,
  revoke, quarantine, reset, dry run, unsupported, degraded, failure, and
  unknown states.
- Check RTL and representative long-string responsive layouts without
  introducing duplicate or unrelated CSS.

## Verification and delivery

Run JSON parsing, CatX parity/type/placeholder checks, the untranslated-English
detector, reserved-key and dead-key checks, frontend format/lint/typecheck/
tests, generated checks, `make verify`, and `make verify-fork`. Use canonical
Linux CI for gates unavailable on Windows; do not modify dependencies to work
around local native bindings.

Push the branch and require green Fork Verification on the exact final SHA.
Do not modify release workflows merely to trigger unrelated qualification.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`
- `docs/21_RC1_UPSTREAM_SYNC.md`
- `docs/22_RC2_STAGING_RELEASE_QUALIFICATION.md`
- `docs/23_RC2_RELEASE_NOTES.md`
- `docs/24_FIRST_RC_PLAN.md`
- `docs/25_FRONTEND_I18N_CONTRACT.md`
- `docs/26_LOCALIZATION_GLOSSARY.md`
