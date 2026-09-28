# CURRENT TASK

## Work Package

`RC-1 — Upstream Sync & Release Candidate Hardening`

## Authorization and guardrails

CatX-UI production baseline is `7ef22f94c950ff09f0870e2295fa65ad5968742c`.
The current base is `develop` after completed WP-8C. Upstream is
`MHSanaei/3x-ui`; merge exact commit
`17d7dd46b512d0a9c22921a6094f30c672e436c9` without rebasing. No release tag,
`main` update, or production update is authorized.

RC-1 must remain a low-divergence downstream sync. Preserve upstream behavior,
CatX-UI modular boundaries, updater ownership, transactional rollback,
feature-disabled compatibility, privacy limits, authentication contracts,
migration safety, Xray boundaries, and generated frontend/API contracts.

Production fleet mutation must remain disabled by default and may execute only
when both the explicit fleet-update mutation setting and explicit admin
confirmation are present. Dry-run never mutates. Only direct, eligible nodes
may execute. A timeout after POST is ambiguous: reconcile node status before
any retry and require positive evidence before redispatch. Never redispatch
blindly or accept status belonging to another `runId`.

## Required RC-1 behavior

- Fetch and merge the exact pinned upstream commit, resolving every conflict
  deliberately and documenting each nontrivial resolution.
- Audit runtime, updater, release identity, workflows, dependencies,
  database/migrations, auth/API, frontend, and Xray boundaries.
- Verify SQLite/PostgreSQL migration and recovery paths where touched.
- Verify checksum/signature, download, disk, service, healthcheck, rollback,
  interrupted-transaction, and stale-state update failures.
- Verify independent feature flags, feature-disabled compatibility, concurrency,
  graceful shutdown, race-sensitive code, and lock boundaries.
- Verify frontend route/API/OpenAPI/generated contracts and i18n as needed.
- Do not copy Remnawave code, add TLS MITM, or collect decrypted HTTP bodies,
  cookies, authorization headers, credentials, or passwords.

## Verification and delivery

Use snapshot → validate → apply → healthcheck → commit for risky state
changes. On failure, restore the previous known-good state, reload/restart as
required, healthcheck, and report/audit the failure. Before merge, require
`make verify`, `make verify-fork`, and canonical CI green on the exact final RC
commit. Keep `main`, tags, and production untouched throughout RC-1.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`
