# CURRENT TASK

## Work Package

`WP-8A — Audit / Webhooks / Metrics`

## Status

WP-7A is DONE, merged into `develop`, and its feature branch was deleted after
the canonical CI runs reported green. Do not modify `main` or preview/demo data.

The WP-8A readiness review is approved. WP-8A is authorized for full
implementation on this feature branch. Do not merge WP-8A or start WP-8B.

## Fixed decisions

- Durable audit storage is authoritative. The existing event bus is
  notification/wake-up only and is never durable audit storage.
- Do not create a second event system.
- Use authoritative controller/service transaction boundaries. If an action
  already has a DB transaction, write the audit row and webhook outbox in it;
  otherwise write a success audit only after the authoritative action succeeds.
- Generate an internal UUID request ID server-side. Never trust an incoming
  request ID as identity.
- Webhook v1 destinations are public HTTPS only. Reject loopback, private,
  link-local, multicast, and other non-public destinations at configuration and
  after DNS resolution on every delivery. Do not follow redirects.
- Webhooks are at-least-once with stable event/delivery IDs, HMAC-SHA256,
  bounded timeout, exponential backoff, leases, crash recovery, and dead
  letters.
- Encrypt webhook secrets at rest with AES-GCM. Reuse an existing cryptographic
  installation key only when one actually exists; otherwise use a fork-owned
  random master-key file outside the DB with 0600 permissions. Never expose a
  secret after creation.
- Add official `prometheus/client_golang` support. Metrics use fixed bounded
  labels only; never email, IP, domain, request/session ID, URL, or arbitrary
  node/client IDs.
- Default retention: audit events 180 days, successful webhook deliveries 30
  days, dead letters 90 days.
- Database-level backup/restore includes WP-8A tables. Normal config/export
  APIs exclude audit history and webhook secret material.
- Feature-disabled behavior is an exact or near-exact no-op.

## Authorized implementation scope

Implement end-to-end:

1. Fork-owned audit persistence, sanitizer/allowlist, and request correlation.
2. Audit authoritative privileged mutations and security/destructive actions.
3. Durable webhook endpoints, transactional outbox, worker, retries, dead
   letters, and manual replay.
4. SSRF, DNS-rebinding, redirect, size, and timeout protections.
5. Prometheus endpoint and bounded metrics.
6. Protected audit/webhook APIs.
7. Audit and Webhooks UI with masked secrets, filters, pagination, detail, and
   dead-letter/replay views.
8. Retention and deletion.
9. OpenAPI/generated types/i18n.
10. SQLite/PostgreSQL migrations.
11. API-token/node-sync/monitor route authorization updates.
12. Transaction, replay, concurrency, race, redaction, SSRF, signing,
    crash-recovery, metrics-cardinality, API/UI, and feature-disabled tests.

Do not persist passwords, tokens, cookies, Authorization headers, bodies,
decrypted content, arbitrary request payloads, or webhook response bodies.

## Required workflow

- Use focused commits and keep the working tree clean.
- Run relevant tests after each implementation stage.
- Run `make verify` and `make verify-fork`.
- Push the feature branch and run canonical CI.
- Inspect exact failures and fix genuine defects until Fork verification and
  Release CatX-UI are green.
- Perform a final WP-8A scope and security audit.
- Do not merge WP-8A, start WP-8B, modify `main`, or touch preview/demo data.

## Required final report

Report final SHA, commits, schema, audited action coverage, webhook
security/delivery model, secret storage, metrics, retention/backup behavior,
API/UI, tests, both CI run IDs, remaining limitations, and clean-tree state.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`

## Authorized skills

- `skills/feature-implementer/SKILL.md`
- `skills/security-privacy-review/SKILL.md`

## Stop condition

Stop on the fully green, pushed WP-8A feature branch after the final scope and
security audit. Do not merge WP-8A or begin WP-8B.
