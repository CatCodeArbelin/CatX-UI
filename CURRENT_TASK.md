# CURRENT TASK

## Work Package

`WP-9A — Frontend Design System & i18n Foundation`

## Authorization and guardrails

The first public CatX-UI RC is `0.1.0-rc.1`, tagged at
`09f5432aacfb2c45f3df79cdca2e15e77bc9a8a6`. WP-9A starts from develop
`0e6949831161db867630e3027db1a5a3d12fe95e`. `main` must remain untouched.

Create and use branch `feature/wp-9a-frontend-design-i18n-foundation`.
Do not merge this work package.

RC-2 must remain a low-divergence downstream qualification package. Preserve
upstream behavior, CatX-UI modular boundaries, updater ownership, transactional rollback,
feature-disabled compatibility, privacy limits, authentication contracts,
migration safety, Xray boundaries, and generated frontend/API contracts.

Production fleet mutation must remain disabled by default and may execute only
when both the explicit fleet-update mutation setting and explicit admin
confirmation are present. Dry-run never mutates. Only direct, eligible nodes
may execute. A timeout after POST is ambiguous: reconcile node status before
any retry and require positive evidence before redispatch. Never redispatch
blindly or accept status belonging to another `runId`.

## Required WP-9A behavior

- Preserve the upstream 3x-ui visual language and styling infrastructure.
- Audit and align CatX pages, navigation, command palette, and API docs labels.
- Normalize CatX translation keys under `fork.*` namespaces.
- Add locale parity, placeholder, type-shape, navigation, and hardcoded-string
  contract tests for all supported locales.
- Establish document and Ant Design RTL direction for `ar-EG` and `fa-IR`.
- Verify narrow responsive layouts without changing backend, database, Xray,
  updater, or release-channel behavior.

## Verification and delivery

No runtime state, database schema, Xray configuration, updater, release,
production, or `main` changes are in scope. Before handoff require frontend
format/lint/typecheck/tests, generated artifact checks, `make verify`,
`make verify-fork`, and canonical CI green on the exact final WP-9A commit.

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
