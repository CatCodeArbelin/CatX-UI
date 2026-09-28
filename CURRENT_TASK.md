# CURRENT TASK

## Work Package

`RC-2 — Staging Qualification & Release Engineering`

## Authorization and guardrails

CatX-UI production baseline is `7ef22f94c950ff09f0870e2295fa65ad5968742c`.
RC-1 was completed and merged into `develop` with merge commit
`ec6fbc3b`. RC-2 starts from that exact post-RC1 develop head. No release
tag, `main` update, production update, or GitHub release is authorized.

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

## Required RC-2 behavior

- Qualify every release artifact emitted by the workflow, including exact
  asset/checksum pairing, archive contents, identity, permissions, and build
  metadata. Corruption and repository/URL mismatches must fail closed.
- Exercise fresh installation and supported upgrades in disposable Linux
  environments from the production baseline, pre-RC develop generation, and a
  populated deterministic database fixture.
- Exercise SQLite and PostgreSQL schema upgrade, backup/restore, repeated
  migration, retention, restart, and feature-disabled startup behavior.
- Exercise the real node-local updater transaction and the real Stage-B remote
  `StartUpdate`/`GetUpdateStatus` fleet boundary. Do not use FakeExecutor for
  final staging qualification.
- Inject download, verification, staging, install, migration, service,
  panel-health, Xray-health, interruption, and stale/malformed evidence
  failures. Verify actual recovery and distinguish rollback failure.
- Verify feature-off upstream compatibility, representative completed-feature
  integration, secret non-disclosure, bounded long-run resource behavior, and
  release documentation.
- Do not copy Remnawave code, add TLS MITM, or collect decrypted HTTP bodies,
  cookies, authorization headers, credentials, or passwords.

## Verification and delivery

Use snapshot → validate → apply → healthcheck → commit for risky state
changes. On failure, restore the previous known-good state, reload/restart as
required, healthcheck, and report/audit the failure. Before the RC-2 branch is
handed off, require `make verify`, `make verify-fork`, all applicable staging
checks, and canonical CI green on the exact final RC-2 commit. Keep `main`,
tags, releases, and production untouched throughout RC-2.

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
