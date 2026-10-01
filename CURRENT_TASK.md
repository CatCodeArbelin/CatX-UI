# CURRENT TASK

## Work Package

`Stable Qualification — v0.1.0`

## Base and branch

RC-6 was merged into `develop` with merge SHA
`ecd8ee1afbee4996f9a099d67014c4ab2236ae44`. This package starts from that
exact `develop` HEAD on branch `feature/stable-v0.1.0-qualification`.

The immutable public tags remain:

- `v0.1.0-rc.1` — unchanged;
- `v0.1.0-rc.2` — `4d8feae2e62db914d3146504340d9b9f802088b2`;
- `v0.1.0-rc.3` — `437d5e2f5bb835118ba6628ea62059b27721cfa3`;
- `v0.1.0-rc.4` — `283abaa401d60b4374a4b618a5a5f8a72ec26cb4`.

`main` remains untouched. Stable `v0.1.0` has not been published.
Repository Productization has not started.

## Required reading

Read `AGENTS.md`, this file, and these canonical documents before changing
implementation or release state:

- `docs/03_ROADMAP.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/22_RC2_STAGING_RELEASE_QUALIFICATION.md`
- `docs/27_SECOND_RC_PLAN.md`
- `docs/28_RC6_REAL_RC_OBSERVATION.md`

## Scope

Qualify the already-observed RC3 code for stable release. This package may
carry only confirmed release-blocking fixes and qualification infrastructure.
It must not add features, redesign QoS, implement generic Xray-user
attribution, start repository productization, or change `main` before every
stable gate is green.

The first mandatory gate is the real user-visible panel restart lifecycle on
actual public `v0.1.0-rc.3` Linux assets:

1. install/start the public RC3 payload on disposable Linux;
2. authenticate and read CatX feature settings;
3. enable `analytics.enabled` and `dns_intelligence.enabled` through the
   normal CatX feature-settings API/UI contract and verify persistence;
4. invoke `POST /panel/api/setting/restartPanel`;
5. wait for panel recovery, authenticate again, and verify both flags;
6. verify Activity, DNS Intelligence, panel health, Xray health, database
   preservation, and absence of panic/fatal/startup-loop evidence;
7. disable both flags through the normal path and repeat the real restart;
8. verify explicit healthy feature-off behavior and persistence.

The regression harness must use the actual restart endpoint. A direct process
stop/start is not an accepted substitute.

If this gate finds a source-level defect, stop Stable Qualification, do not
publish stable, and require focused `v0.1.0-rc.4` qualification and real
observation before returning here. Infrastructure-only failures must be
recorded precisely and retried according to the existing policy.

## Stable identity

The eventual stable release must use:

- fork version `0.1.0`;
- release version `0.1.0`;
- channel `stable`;
- upstream base `MHSanaei/3x-ui v3.8.5`;
- bundled Xray `26.9.9`;
- annotated tag `v0.1.0` with `prerelease=false` and the normal stable/latest
  release semantics.

RC tags and assets must remain unchanged.

## Stable Qualification decision

The confirmed RC3 restart blocker is resolved by public corrective release
`v0.1.0-rc.4`. Stable Qualification is **READY TO RESUME**, but must not resume
automatically before review of this evidence.

Root cause and fix:

- RC3 coupled fork schema preparation and persisted runtime configuration in
  `forkext.RegisterMigrations()`; the panel-only SIGHUP path did not repeat the
  runtime configuration step;
- RC4 separates initial fork migration preparation from runtime configuration
  and reloads persisted managed-feature state after old panel-owned workers
  stop and before the new panel runtime starts;
- newly enabled fork schema is prepared before activation, failures disable the
  runtime and propagate, and upstream DB initialization/seeders are not rerun;
- enabled-to-enabled Traffic Control reload preserves CatX rollback ownership.

Qualification evidence:

- focused exact-SHA candidate restart regression: run `36806720158`, PASS;
- final exact SHA: `283abaa401d60b4374a4b618a5a5f8a72ec26cb4`;
- final Fork Verification: run `36807215175`, PASS;
- final restart plus Linux Traffic Control/quota/no-clobber qualification: run
  `36807215104`, PASS;
- non-publishing RC artifact, checksum, Linux, Windows, SQLite, PostgreSQL,
  clean-install, RC3-to-RC4 upgrade, DB-preservation, health, and rollback
  matrix: run `36807225858`, PASS;
- annotated public tag `v0.1.0-rc.4` targets the final exact SHA;
- canonical tagged Release CatX-UI workflow: run `36807931375`, PASS;
- public release is `draft=false`, `prerelease=true`, has 28 qualified assets,
  and did not create or move a stable latest release;
- actual-public-assets restart observation: run `36808874388`, PASS.

Public RC4 observation proved both directions through the real
`POST /panel/api/setting/restartPanel` endpoint: persisted enabled state became
enabled runtime with working Activity and DNS Intelligence, and persisted
disabled state became disabled runtime with explicit healthy feature-off
behavior. Panel, subscription server, SQLite state, and Xray remained healthy.
Representative policy, Traffic Control, audit, self-service, and fleet runtime
surfaces also matched persisted state after restart.

`v0.1.0-rc.1`, `v0.1.0-rc.2`, and `v0.1.0-rc.3` remain unchanged. `main`
remains `7ef22f94c950ff09f0870e2295fa65ad5968742c`; stable `v0.1.0` has not
been published, and Repository Productization has not started. Stop here until
Stable Qualification is explicitly resumed.
