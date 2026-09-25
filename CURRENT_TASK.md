# CURRENT TASK

## Work Package

`WP-5B — Shared Group Quota / Accounting Extensions`

## Status

WP-5A is DONE and merged into `develop`. WP-5B is activated for branch setup
only. Do not implement WP-5B in this task. Do not modify `main`.

## Scope

The next package is shared group quota and accounting extensions. This task
only activates the package and creates its feature branch; implementation is
intentionally deferred.

Required scope after the stopping point: shared group quota, atomic accounting,
and optional fixed-point traffic multiplier, reusing upstream counters.

## Required workflow

- Create and push the WP-5B feature branch, then stop immediately.
- Do not implement WP-5B. Do not touch preview/demo data.

## Authorized documents

- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/06_DATABASE_MIGRATIONS.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/10_POLICY_ENGINE.md`

## Authorized skills

- `skills/database-migration/SKILL.md`
- `skills/feature-implementer/SKILL.md`
- `skills/security-privacy-review/SKILL.md`

## Hard constraints

- reuse existing traffic accounting and client/group models;
- no duplicate counters, parallel accounting store, or unrelated refactor;
- no sensitive payloads, cookies, credentials, or decrypted message storage;
- no runtime schema mutation outside migrations;
- no changes to `main` and no preview/demo data changes;
- keep the working tree clean at handoff.

## Stop condition

The WP-5B feature branch is the final state for this task. Do not implement
any WP-5B production code.
