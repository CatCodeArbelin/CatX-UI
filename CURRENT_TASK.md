# CURRENT TASK

## Work Package

`WP-5A — Historical Traffic`

## Status

Implementation authorized on a dedicated feature branch created from `develop`.
WP-4B is DONE and merged into `develop`. Do not modify `main`.

## Scope

Implement historical traffic reporting by reusing upstream traffic accounting
and the existing CatX analytics/history foundations. Do not create parallel
traffic counters or duplicate client, inbound, or node accounting.

Required scope:

- historical traffic queries over the existing accounting/history data;
- custom date ranges with bounded validation and timezone-safe semantics;
- service and category breakdowns using existing vocabulary and models;
- API, OpenAPI, and UI integration through existing analytics boundaries;
- retention-compatible behavior and documentation;
- SQLite and PostgreSQL coverage, including migration/upgrade coverage if a
  persisted model changes;
- idempotence, concurrency/race, authorization, pagination/aggregation, and
  feature-disabled compatibility tests;
- focused commits, pushed feature branch, and green `Fork verification` and
  `Release CatX-UI` workflows.

## Required workflow

- Inspect existing upstream traffic accounting and CatX analytics/history
  foundations before choosing integration points.
- Preserve upstream behavior and avoid a parallel accounting subsystem.
- Run relevant unit/integration/database/race/API/OpenAPI/UI tests plus
  `make verify` and `make verify-fork` as supported by the environment.
- Fix genuine CI failures and inspect exact failure output.
- After final scope review and both required workflows are green, merge WP-5A
  into `develop`, push it, delete the WP-5A feature branch, mark WP-5A DONE,
  activate WP-5B, create and push the WP-5B feature branch, then stop.
- Do not implement WP-5B. Do not touch preview/demo data.

## Authorized documents

- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/06_DATABASE_MIGRATIONS.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/10_POLICY_ENGINE.md`

## Authorized skills

- `skills/analytics-dns/SKILL.md`
- `skills/database-migration/SKILL.md`
- `skills/feature-implementer/SKILL.md`
- `skills/security-privacy-review/SKILL.md`

## Hard constraints

- reuse existing traffic accounting, analytics/history, client/inbound/node
  models, categories, retention, authorization, and lifecycle paths;
- no duplicate counters, parallel history store, or unrelated refactor;
- no sensitive payloads, cookies, credentials, or decrypted message storage;
- no runtime schema mutation outside migrations;
- no changes to `main` and no preview/demo data changes;
- keep the working tree clean at handoff.

## Next package

After WP-5A is green, merged, and its branch is deleted, activate
`WP-5B — Shared Group Quota / Accounting Extensions`, create and push its
feature branch, and stop without implementing it.
