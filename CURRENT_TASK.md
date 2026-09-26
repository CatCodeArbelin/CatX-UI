# CURRENT TASK

## Work Package

`WP-6A — Shaping Core`

## Status

WP-5B is DONE and merged into `develop`. WP-6A is authorized for
implementation-readiness review on the current feature branch. Do not
implement production code, do not modify `main`, and do not touch preview/demo
data.

## Scope

The package is the shaping core: capability detection, a shaping adapter, a
Linux implementation, and reconciliation. This task is limited to an
implementation-readiness review; production code is intentionally deferred.

The review must define Linux traffic-shaping primitives and ownership
boundaries, per-client identity/marking, Xray and client/node integration,
restart/config reconciliation, idempotent apply/remove, unsupported-platform
no-op behavior, failure/rollback, local versus remote-node operation,
interaction with WP-5B quota state, privilege requirements, concurrency, and
the test strategy. Do not create a second traffic/accounting engine.

## Required workflow

- Read the authorized documents and skills, inspect the actual current
  integration points, and produce the readiness review only.
- Do not implement production code, change schemas/API, or modify preview/demo
  data.
- Keep the working tree clean at handoff.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/11_QOS_TRAFFIC_CONTROL.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`

## Authorized skills

- `skills/feature-implementer/SKILL.md`
- `skills/security-privacy-review/SKILL.md`

## Hard constraints

- reuse existing Xray, client, inbound, and node boundaries;
- no second traffic/accounting engine;
- no sensitive payloads, cookies, credentials, or decrypted message storage;
- no runtime schema mutation outside migrations;
- no changes to `main` and no preview/demo data changes;
- keep the working tree clean at handoff.

## Stop condition

The WP-6A readiness review is complete. Do not implement production code.
