# CURRENT TASK

## Work Package

`CatX v0.2.0 Release Qualification`

## Status

`CURRENT`

This package qualifies the already-integrated upstream `v3.9.0` CatX
develop state for one public `v0.2.0-rc.1` release, actual-public-artifact
observation, and stable `v0.2.0` publication only if the exact RC passes.
This is not an upstream sync and does not add product features.

## Release baseline

- release branch: `feature/v0.2.0-release-qualification`
- release baseline SHA: `31a442cef876669aeae97ba9cb1e3e8888f42613`
- starting `origin/develop`: `31a442cef876669aeae97ba9cb1e3e8888f42613`
- starting `origin/main`: `7cb97e13def59280e92e5667d3f43ac23db8d470`
- previous CatX stable: `v0.1.0`
- previous stable runtime SHA: `fd28ea7144147d9164b70810d4a24872a3d48b4f`
- new fork version: `0.2.0`
- first and only planned RC: `0.2.0-rc.1`
- frozen RC candidate source SHA: `6a7e7aa6d9dcec4214ad0a7ba5aefc1cc9e0e998`
- upstream base: `MHSanaei/3x-ui v3.9.0`
- upstream SHA: `3cd4bf504c3cd8ea9b1c1fdb032a9796c5c43ddb`
- upstream integration merge SHA: `c5f2e4e0165d96577af3f0e9d3b40810a98e7764`
- Xray: `26.9.30`
- existing stable tag: `v0.1.0`
- `v0.2.0-rc.1` tag: not created
- `v0.2.0` tag: not created

`origin/main` is an ancestor of the release baseline. No legitimate main-only
commits require reconciliation before qualification.

## Scope

- release identity transition to CatX `0.2.0` / `0.2.0-rc.1`;
- release candidate source and artifact qualification;
- public `v0.2.0-rc.1` creation and immutable asset observation;
- SQLite and PostgreSQL migration/upgrade coverage;
- backup/restore, updater transaction, and rollback qualification;
- real `restartPanel`, Xray/live VLESS, native TUIC, subscriptions,
  AmneziaWG, multi-node, Traffic Control, quota, policy, Activity/DNS,
  localization, and product-reality checks;
- stable qualification and `v0.2.0` publication only if the exact public RC
  passes without a product/runtime source change;
- final evidence in `docs/31_V0_2_0_RELEASE_QUALIFICATION.md`.

## Non-goals

- no new product features;
- no new upstream sync;
- no unrelated refactor or repository-wide formatting;
- no Traffic Control attribution expansion;
- no generic plugin system;
- no architectural redesign;
- no rewrite of historical `v0.1.0` notes;
- no preemptive `rc.2`;
- no stable-only product/runtime patch;
- no modification of `AGENTS.md`;
- no force/movement of existing tags.

## Release and RC policy

The only planned public candidate is `v0.2.0-rc.1`. Do not create `rc.2`
unless public RC observation reveals a genuine source, release identity,
artifact, migration/update/rollback, or runtime defect. Infrastructure-only
failures are retried and classified. Observation-harness-only defects are
fixed without mutating immutable RC assets.

The RC tag must point exactly to the frozen qualified candidate. Stable must
reuse the same proven product/runtime source; any RC-to-stable commit may only
change the required release identity and evidence metadata. Stable publication
requires a qualified `main` merge, annotated `v0.2.0`, canonical stable release
workflow, Docker verification, and actual-public stable smoke.

## Qualification policy

Hosted Linux CI is authoritative. Required gates include `make verify`,
`make verify-fork`, Go/race/SQLite/PostgreSQL/migration tests, generated and
frontend checks, docs checks, release identity, updater/rollback, Xray,
Docker, localization, non-publishing release matrix, and actual public RC
observation. A skipped check is recorded as skipped with its reason, never as
pass.

Before tagging, the branch must be clean and pushed, all exact-SHA required
checks must be green, and the candidate SHA must be recorded in this file,
`TASK_QUEUE.md`, and `docs/31_V0_2_0_RELEASE_QUALIFICATION.md`.

Every risky apply or release operation follows:

```text
snapshot → validate → apply → healthcheck → commit known-good state
FAIL → restore previous state → restart/reload → healthcheck → audit failure
```

The optional Claude review workflow is not a source qualification gate unless
repository governance explicitly makes it required. If it appears again with
missing credentials, record the infrastructure condition without fabricating a
review or blocking an otherwise qualified release.

## Required reading

- `AGENTS.md`
- `CURRENT_TASK.md`
- `TASK_QUEUE.md`
- `docs/00_MASTER_SPEC.md`
- `docs/01_ARCHITECTURE.md`
- `docs/02_FEATURE_CATALOG.md`
- `docs/04_UPSTREAM_SYNC.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/06_DATABASE_MIGRATIONS.md`
- `docs/08_FRONTEND_UX.md`
- `docs/09_ANALYTICS_DNS.md`
- `docs/10_POLICY_ENGINE.md`
- `docs/11_QOS_TRAFFIC_CONTROL.md`
- `docs/12_UPDATER_IDENTITY.md`
- `docs/14_NON_GOALS.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`

## Completion condition

Do not mark this package complete until the final public stable qualification
record contains the RC and stable candidate/tag/release/observer/Docker/main
SHAs and URLs, migration/upgrade/backup/restore/updater/rollback evidence,
runtime and product-reality evidence, and immutable-tag verification.

At completion set this file and `TASK_QUEUE.md` to `COMPLETE`. Finish with
exactly one required release status phrase from the package request.
