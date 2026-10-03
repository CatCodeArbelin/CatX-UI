# CURRENT TASK

## Work Package

`Upstream Sync — 3x-ui v3.9.0`

## Status

`COMPLETE`

This work package performs the actual merge-based integration of the exact
upstream `v3.9.0` release into the CatX integration line. The qualified sync
was merged into `develop`; no release or tag was published.

## Integration baseline

- CatX integration base SHA: `65957a8988b5b51c4ac01dbd1121f04b55af6c8a`
- Strategy PR: `https://github.com/CatCodeArbelin/CatX-UI/pull/3`
- Strategy merge SHA: `7cb97e13def59280e92e5667d3f43ac23db8d470`
- Develop reconciliation: `main → develop` merge SHA
  `65957a8988b5b51c4ac01dbd1121f04b55af6c8a`
- Previous upstream base: `MHSanaei/3x-ui v3.8.5`
- Previous upstream base SHA: `7ef22f94c950ff09f0870e2295fa65ad5968742c`
- Target upstream: `MHSanaei/3x-ui v3.9.0`
- Target upstream SHA: `3cd4bf504c3cd8ea9b1c1fdb032a9796c5c43ddb`
- Sync branch: `sync/upstream-v3.9.0`
- Preserved dry-run branch: `sync/upstream-v3.9.0-dry-run`

## Known dry-run evidence

- upstream commits: `99`
- changed upstream paths: `563`
- sensitive paths: `224`
- CatX touchpoint overlaps: `64`
- dry-run merge conflicts: `12`
- actual merge conflict count: `12`

## Scope

- integrate upstream `v3.9.0` while preserving upstream ancestry;
- preserve CatX architecture, product identity, and release identity;
- adopt upstream implementations where they replace equivalent inherited
  behavior;
- adapt CatX hooks to the new upstream lifecycle boundaries;
- audit and qualify database/migration, Xray, TUIC, AmneziaWG, client,
  subscription, frontend/generated, updater, and feature-off behavior;
- integrate the exact qualified candidate into `develop`.

## Non-goals

- no new CatX feature work;
- no unrelated refactor or generic cleanup;
- no `v0.2.0` release or tag;
- no direct merge to `main`;
- no rewrite of CatX architecture;
- no fork of `xray-core`;
- no new sensitive-data collection;
- no automatic punitive action based on one heuristic signal.

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

## Qualification policy

The exact sync merge and every focused correction must be recorded in
`docs/30_UPSTREAM_3_9_0_SYNC.md`. No source changes may be made after the
exact qualified candidate is frozen without rerunning the affected gates.

`make verify`, `make verify-fork`, hosted Linux qualification, SQLite and
PostgreSQL migration coverage, generated-artifact checks, frontend checks,
runtime smoke, and the non-publishing release matrix are required where the
environment supports them. A check that was not run is not recorded as pass.

The real `restartPanel` lifecycle must preserve persisted CatX flags, schema
preparation/runtime-configuration separation, reload error propagation, and
Traffic Control rollback ownership. With all CatX features disabled, behavior
must follow integrated upstream `v3.9.0` semantics as closely as possible.

Do not modify `AGENTS.md` unless a new permanent repository-wide invariant is
discovered; stop and report the proposed wording before doing so.

## Completion record

- upstream merge commit: `9335eb4af5ab4c73976fdcbeda875926400e53f8`;
- final qualified code candidate: `196ca986414bcc8e6ec7af4a9d807669546faaee`;
- sync PR: [#4](https://github.com/CatCodeArbelin/CatX-UI/pull/4);
- `develop` merge commit: `c5f2e4e0165d96577af3f0e9d3b40810a98e7764`;
- `main` remains `7cb97e13def59280e92e5667d3f43ac23db8d470`;
- Xray version after sync: `26.9.30`;
- release publication and tag movement: not performed.
