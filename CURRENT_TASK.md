# CURRENT TASK

## Work Package

`WP-7A — Risk Intelligence`

## Status

WP-6B is DONE and merged into `develop`. Its generic Xray-user to
kernel-attribution limitation remains documented and deferred. Do not modify
`main`, reopen WP-6B, or touch preview/demo data.

WP-7A is active for implementation-readiness review only. Do not implement
production code in this task.

## Authorized review scope

Review the minimum modular architecture for:

- extended client IP/session history;
- ASN/country change detection;
- account-sharing heuristics;
- simultaneous-location and impossible-travel-style signals only where the
  evidence is sufficiently strong;
- DNS anomaly heuristics;
- a risk score and risk-event model with no automatic bans by default;
- reuse of existing analytics, session, DNS, IP, and remote-node data;
- false-positive control;
- retention/privacy, API/UI, alerting, fleet consistency, and concurrency;
- implementation, migration, rollback, and test readiness.

Do not create a second analytics/session/IP collection pipeline. Do not
auto-ban users by default. Do not modify `main`.

The generic Xray-user kernel-attribution limitation from WP-6B remains
unsupported/deferred and must not be reopened by this review.

## Required review output

Report only:

- proposed risk architecture;
- data sources reused;
- risk/event model;
- heuristics and confidence model;
- false-positive protections;
- schema/API/UI changes;
- privacy/retention;
- test matrix;
- blockers;
- whether Luna Medium can safely implement WP-7A.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`

## Authorized skills

- `skills/security-privacy-review/SKILL.md`

## Hard constraints

- reuse existing analytics/session/DNS/IP data and event/lifecycle hooks;
- no parallel collection pipeline or duplicate counters;
- no TLS MITM or decrypted payload/body/cookie/credential storage;
- no IP-only account-sharing proof;
- no automatic bans by default and no single-signal enforcement;
- every persisted change requires explicit SQLite/PostgreSQL migration and
  retention/index review;
- remote-node data must be source-labelled, deduplicated, and tolerant of
  clock skew, delay, restart, and partial loss;
- feature-disabled behavior remains as close as possible to upstream;
- keep the tree clean and do not modify `main`.

## Stop condition

Finish the readiness review and identify blockers. Do not implement production
code or begin WP-7B.
