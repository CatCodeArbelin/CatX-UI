# CURRENT TASK

## Work Package

`WP-6B — Speed / Rolling Quota / Soft Throttle`

## Status

WP-6A is DONE and merged into `develop`. WP-6B Stage A and the supported,
capability-gated Stage B scope are complete on the current feature branch.
Generic Xray-user attribution/marking remains deferred. Do not modify `main`,
start WP-7A, or touch preview/demo data.

## Authorized Stage A scope

Implement only:

- fixed/tumbling-window quota accounting;
- soft-throttle lifecycle/state;
- fork-owned persistence and SQLite/PostgreSQL migrations;
- protected API, UI, OpenAPI, generated types, and i18n;
- restart/reconciliation state and capability/unsupported reporting;
- integration with WP-5B authoritative accounting and WP-6A shaping;
- comprehensive unit, integration, race, restart, remote-node, frontend, and
  feature-disabled compatibility tests;
- a non-production mark-feasibility spike and tests/harness only.

Rolling/sliding windows are not authorized. Generic Xray-user → kernel-mark
enforcement is not production-authorized unless a non-invasive experiment
proves semantic equivalence for existing routing, policy precedence, balancers,
WARP, direct, and blocked paths. If it remains unproven, report the exact
blocker and keep generic users explicitly unsupported.

Reuse authoritative `client_traffics` deltas, the existing WP-5B accounting
transaction, the WP-6A `Shaper`/`DesiredRule` contract, and authenticated
`runtime.Remote`. Do not add duplicate traffic counters, a second routing
compiler, or a second node subsystem.

## Authorized Stage B scope

Complete only the capability-gated end-to-end enforcement path:

- traffic policy/lifecycle state → attribution provider → `trafficcontrol.DesiredRule` → WP-6A `Shaper.Reconcile`;
- the smallest explicit attribution-provider boundary;
- production providers only when they prove stable kernel-visible identity;
- cleanup of stale shaping state on disable, expiry, detach, delete, unsupported,
  degraded, or lost attribution;
- deterministic ACTIVE/THROTTLED reconciliation after restart and window reset;
- existing authenticated remote runtime transport only;
- fixed/tumbling windows only for this release;
- API/UI capability gating and comprehensive end-to-end, rollback, and
  feature-disabled compatibility tests.

Generic Xray-user → kernel-mark enforcement remains explicitly not authorized.
If no production-safe provider exists, the production provider set must remain
empty and the system must prove explicit unsupported/no-op behavior. Do not
invent generic marking, clone Xray routing, add counters, or add a node
subsystem.

## Required workflow

- Use focused commits and keep the working tree clean.
- Run relevant tests after each stage; run `make verify` and `make verify-fork`.
- Inspect exact failures and fix genuine defects only.
- Do not merge WP-6B or advance to WP-7A.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/11_QOS_TRAFFIC_CONTROL.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`

## Authorized skills

- `skills/feature-implementer/SKILL.md`
- `skills/security-privacy-review/SKILL.md`

## Hard constraints

- hard quota, expiry, and manual disable always outrank soft throttle;
- ownership/reason must prevent reset/recovery from re-enabling unrelated
  disables;
- unsupported/degraded enforcement must never be presented as active;
- no sensitive payloads, cookies, credentials, or decrypted message storage;
- no runtime schema mutation outside migrations;
- no changes to `main` or preview/demo data.

## Stop condition

WP-6B is complete for the supported/capability-gated scope after verification
and final scope audit. Do not merge WP-6B or start WP-7A.
