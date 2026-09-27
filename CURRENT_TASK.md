# CURRENT TASK

## Work Package

`WP-8A — Audit / Webhooks / Metrics`

## Status

WP-7A is DONE, merged into `develop`, and its feature branch was deleted after
the canonical CI runs reported green. Do not modify `main`.

WP-8A is authorized for implementation-readiness review only. Do not implement
production code in this task.

## Review objective

Define a low-divergence, fork-owned design for:

- durable audit logging;
- webhook delivery;
- Prometheus metrics;
- API/UI exposure and administration.

The review must resolve the audit event model, existing actions/events to audit,
actor/target/node/request correlation, secret and sensitive-payload redaction,
retention/pagination/filtering/deletion, webhook signing/retry/backoff/timeout/
dedupe/idempotency/dead-letter behavior, event-bus reuse without treating it as
durable storage, bounded Prometheus labels, local/remote node attribution, API,
UI, migrations, rollback, concurrency, security, and the complete test matrix.

## Hard constraints

- Reuse the existing event bus for notifications only; do not create a second
  event system and do not use the bus as the durable audit ledger.
- Do not log credentials, tokens, cookies, Authorization headers, passwords,
  decrypted content, or HTTP bodies.
- Do not expose unbounded Prometheus labels such as raw IP, domain, client
  email, or session ID.
- Preserve upstream architecture and behavior with minimal integration points;
  prefer fork-owned modules, registries, adapters, and lifecycle hooks.
- Every persisted change requires explicit SQLite/PostgreSQL migration, index,
  retention, deletion, upgrade, and rollback review.
- Feature-disabled behavior must remain as close as possible to upstream.
- Do not modify `main`, preview/demo data, or unrelated upstream code.
- Keep the tree clean. This task must not implement production code.

## Required report

Report only:

- proposed architecture;
- audit event model;
- webhook delivery model;
- metrics model;
- data sources/hooks reused;
- schema/API/UI changes;
- privacy/security;
- retention;
- test matrix;
- blockers;
- whether Luna Medium can safely implement WP-8A.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`

## Authorized skills

- `skills/security-privacy-review/SKILL.md`

## Stop condition

Stop after the readiness review. Do not implement, merge, or modify `main`.
