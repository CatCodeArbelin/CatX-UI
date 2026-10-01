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
- `v0.1.0-rc.3` — `437d5e2f5bb835118ba6628ea62059b27721cfa3`.

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

## Completion state

Stable Qualification is in progress. Do not merge to `main`, create `v0.1.0`,
or start Repository Productization until the exact-SHA stable qualification,
stable cut, actual published-asset verification, and post-publication smoke
all pass. After successful publication, stop before Productization.
