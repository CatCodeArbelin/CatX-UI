# CURRENT TASK

## Work Package

`WP-7A — Risk Intelligence`

## Status

WP-6B is DONE and merged into `develop`. Its generic Xray-user to
kernel-attribution limitation remains documented and deferred. Do not modify
`main`, reopen WP-6B, or touch preview/demo data.

WP-7A readiness review is approved. WP-7A is authorized for full
implementation on this feature branch. Do not start WP-8A.

## Fixed decisions

- Extend the existing `CheckClientIpJob` / node IP attribution write path with
  durable fork-owned IP history. Do not create another collector.
- Reuse existing analytics sessions, DNS observations, destination evidence,
  node attribution, clients/groups, and `analytics.IPMetadataProvider`.
- ASN/country enrichment is optional. Missing or stale provider data must
  produce unknown/degraded confidence, never guessed values.
- Do not implement geographic impossible-travel speed without reliable
  coordinates. Use only defensible overlapping session/country/ASN divergence
  signals.
- Risk is informational only: no automatic ban, disable, throttle, or routing
  mutation.
- Scores are 0–100 with separately exposed confidence and evidence.
- A single IP change, VPN/mobile/CGNAT/datacenter/Tor classification, new ASN,
  or DNS anomaly alone must not create a high-risk conclusion.
- Durable IP history defaults to 30 days; risk events default to 90 days.
  Both are configurable and independently deletable.
- The event bus may notify but is not durable storage.

## Authorized implementation scope

Implement end-to-end:

- fork-owned IP history, risk event, risk score, and suppression/
  acknowledgement persistence;
- deterministic dedupe and replay-safe ingestion;
- simultaneous multi-IP/node/country session heuristics;
- repeated ASN/country divergence;
- source-IP churn correlated with active sessions;
- bounded DNS anomaly heuristics using existing DNS data;
- confidence, decay, stale/degraded, and insufficient-evidence states;
- client risk summary/timeline APIs;
- acknowledgement/suppression APIs with expiry;
- client Risk Intelligence UI with evidence and source-node attribution;
- retention/deletion;
- OpenAPI, generated types, and i18n;
- remote-node delay/restart/clock-skew handling;
- comprehensive SQLite/PostgreSQL, concurrency, replay, false-positive, API,
  UI, and feature-disabled tests.

Keep all heuristics deterministic and explainable. Store metadata only. Do not
create a second analytics/session/IP collection pipeline. Do not auto-ban
users. Do not modify `main` or preview/demo data.

## Required workflow

- Use focused commits and keep the working tree clean.
- Inspect exact test failures and fix genuine defects only.
- Run relevant tests after each implementation stage.
- Run `make verify` and `make verify-fork`.
- Push the feature branch and run canonical CI.
- Do not merge WP-7A or start WP-8A.
- Finish with a final WP-7A scope audit.

## Required final report

Report commits, schema, ingestion hook, heuristics, confidence/scoring model,
retention/privacy, API/UI, tests, CI run IDs, remaining limitations, and tree
state.

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

## Hard constraints

- reuse existing analytics/session/DNS/IP data and event/lifecycle hooks;
- the existing `CheckClientIpJob` / node attribution path is the sole IP
  ingestion boundary;
- no parallel collector, duplicate counters, or second node subsystem;
- no TLS MITM or decrypted payload/body/cookie/credential storage;
- no IP-only account-sharing proof;
- no automatic bans, disables, throttles, or routing changes;
- every persisted change requires explicit SQLite/PostgreSQL migration, index
  review, retention review, and deletion behavior;
- remote data must be source-labelled, deduplicated, replay-safe, and tolerant
  of clock skew, delay, restart, and partial loss;
- event bus notifications are best-effort only and never the event ledger;
- feature-disabled behavior remains as close as possible to upstream;
- generic Xray-user kernel attribution remains unsupported/deferred;
- keep the tree clean and do not modify `main`.

## Stop condition

Stop on the fully green, pushed feature branch after the final scope audit.
Do not merge WP-7A or begin WP-8A.
