# CURRENT TASK

## Work Package

`WP-6B — Speed / Rolling Quota / Soft Throttle`

## Status

WP-6A is DONE and merged into `develop` at the green CI head. WP-6B is
authorized for an implementation-readiness review only. Do not implement
production code, modify `main`, or touch preview/demo data.

## Review scope

Resolve the minimum modular design for:

- per-client upload/download speed limits over the WP-6A shaping substrate;
- generic Xray-user-to-kernel-mark attribution without bypassing existing
  routing, policies, balancers, WARP, direct, or blocked paths;
- rolling/fixed-window quota accounting without duplicate counters;
- ACTIVE → THROTTLED → ACTIVE lifecycle and interaction with WP-5B hard
  quota/disable semantics;
- restart/reconciliation and local/remote-node enforcement;
- optional category caps only when attribution is reliable;
- schema/API/UI, rollback/failure, unsupported behavior, and concurrency tests.

Do not introduce a second routing compiler, traffic counter, or node subsystem.
The review must identify the smallest upstream integration boundary, exact
unsupported/degraded behavior, and genuine blockers before implementation.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/11_QOS_TRAFFIC_CONTROL.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`

## Authorized skills

- `skills/security-privacy-review/SKILL.md`

## Hard constraints

- reuse existing Xray, client, inbound, traffic, and node boundaries;
- use the WP-6A Shaper and existing traffic accounting only;
- no sensitive payloads, cookies, credentials, or decrypted message storage;
- no runtime schema mutation outside migrations;
- no generic mark path that changes existing route/policy decisions;
- no production implementation in this review;
- no changes to `main` or preview/demo data;
- keep the working tree clean at handoff.

## Stop condition

Stop after the WP-6B readiness report. Do not implement production code or
advance to another work package.
