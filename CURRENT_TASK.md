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

`main` and annotated stable tag `v0.1.0` now both resolve to the qualified
stable SHA `fd28ea7144147d9164b70810d4a24872a3d48b4f`. The public stable
release has been published without moving any RC tag. Repository
Productization has not started.

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

Stable Qualification is **COMPLETE**. The qualified and published stable
product/runtime SHA is `fd28ea7144147d9164b70810d4a24872a3d48b4f`.
The RC4-to-stable delta contains release/qualification infrastructure, tests,
and evidence only; it contains no non-test Go or `frontend/src` runtime-source
change.

Pre-publication qualification evidence:

- exact-SHA Fork Verification run `36811795614` passed the canonical
  `make verify` and `make verify-fork` coverage: generated-file drift, lint,
  formatting, type checking, all Go/frontend tests, i18n, accessibility, RTL,
  app/Storybook builds, race tests, and release tests;
- exact-SHA stable candidate lifecycle run `36811795653` passed stable binary
  identity, real enabled and disabled panel restarts, Traffic Control
  apply/remove/no-clobber, and quota lifecycle checks;
- non-publishing stable release matrix run `36811937162` passed all Linux and
  Windows builds, metadata/checksum/archive validation, clean SQLite install,
  PostgreSQL migration, actual public RC4-to-stable upgrade, database
  preservation, repeated restart health, and rollback to RC4;
- reconnect/state recovery remained covered by
  `TestSetRemoteTraffic_DirtyPreservesConfig`,
  `TestNodeService_UpdateMarksNodeDirty`, and
  `TestReconcileInbound_SkipsUnchanged`;
- node-restart recovery remained covered by
  `TestSetRemoteTraffic_EmptySnapshotKeepsCentralInbounds`,
  `TestReconcileInbound_SkipsUnchanged`, and
  `TestConfigureEnabledPreservesRollbackOwnershipAcrossPanelRestart`;
- multi-node/state-changing fanout remained covered by the create, update,
  delete, detach, reset-traffic, inbound-delete, and panel-update concurrency
  tests in the same Fork Verification run;
- remote Traffic Control readiness, authenticated mutation boundaries,
  unsupported/degraded refusal without mutation, and the intentionally
  unsupported generic-user mark path remained covered by focused hosted tests.

Stable publication evidence:

- `main` and annotated tag `v0.1.0` both resolve to
  `fd28ea7144147d9164b70810d4a24872a3d48b4f`;
- canonical tagged Release CatX-UI run `36813723610` passed;
- public release `v0.1.0` is `draft=false`, `prerelease=false`, is GitHub's
  latest release, and exposes all 28 expected assets;
- the canonical release and public observer validated asset names, checksums,
  archives, embedded stable identity, `channel=stable`, `releaseVersion=0.1.0`,
  and the exact build commit;
- corrected actual-public-assets observer run `36815136071` passed against the
  released Linux archive and metadata. It proved updater stable identity,
  clean feature-off SQLite/panel/Xray health, live VLESS traffic, Russian LTR,
  Persian RTL, both real `POST /panel/api/setting/restartPanel` directions,
  generic managed-feature reload, Traffic Control capability/apply/remove and
  idempotence, admin-owned qdisc no-clobber, authoritative group quota and
  fixed-window lifecycle, and cleanup on feature disable/restart;
- generic per-user kernel enforcement remains explicitly unsupported because
  generic Xray-user attribution remains false; the public observer confirmed
  no false active-enforcement claim.

The first public observer attempt `36814708161` ran immediately after release
publication and was initially classified as latest-release visibility lag.
The unchanged second attempt `36814901784` reproduced the updater assertion
failure and showed the real issue: the qualification assertion expected
`latestVersion=0.1.0`, while the stable API correctly returns the release tag
`latestVersion=v0.1.0`. Test-only commit
`35a03bb9d4407ebe343d9d8c0f8c62024dbdc4bb` corrected that observer assertion
and retained the response JSON before checking it. This commit changed only
`scripts/staging/observe-rc2-linux.sh`, was not merged to `main`, and did not
change the published binary or stable tag. The failed attempts remain recorded
and are qualification-observer failures, not published-binary defects.

Container publication evidence:

- RC4 Docker run `36807931390` completed successfully in the CatX-only
  `ghcr.io/catcodearbelin/catx-ui` namespace;
- stable Docker run `36813723615` completed successfully for the qualified
  stable SHA;
- stable aliases `v0.1.0`, `0.1.0`, and `latest` all resolve to OCI index
  digest `sha256:3e94b98560e8261c315b9f9ae4d97a0e8ea5977e00072f9b310088eb9fac1322`;
- the index contains Linux `amd64`, `arm64`, `arm/v7`, `arm/v6`, and `386`
  images plus one provenance/SBOM attestation manifest per platform;
- OCI labels identify `CatX-UI`, the CatX repository, version `0.1.0`, and
  revision `fd28ea7144147d9164b70810d4a24872a3d48b4f`. No upstream-owned image
  namespace was targeted.

Non-canonical post-merge workflow results are retained rather than silently
marked green:

- main CI run `36813724304` built, linted, type-checked, and tested
  successfully, then failed its dynamic npm audit after source freeze on new
  `brace-expansion` advisories (high) and a DOMPurify advisory (low). No
  confirmed stable-source runtime defect was established during qualification;
  dependency refresh remains follow-up work outside this frozen release;
- Docs CI run `36813724270` reported documentation format drift, and Pages run
  `36813724292` failed because GitHub Pages is not enabled. These are
  repository/tooling/deployment-configuration follow-ups, not stable binary
  runtime failures;
- secondary release-install smoke run `36814645136` passed clean install,
  release identity, credentials, and live panel HTTP on amd64 and arm64, then
  failed its same-container repeat-install phase because the no-systemd
  harness cannot perform the transactional updater's service control. The
  dedicated stable clean-install, actual RC4-to-stable upgrade, healthcheck,
  and rollback matrix passed in run `36811937162`.

Immutable tag targets remain: RC1
`09f5432aacfb2c45f3df79cdca2e15e77bc9a8a6`, RC2
`4d8feae2e62db914d3146504340d9b9f802088b2`, RC3
`437d5e2f5bb835118ba6628ea62059b27721cfa3`, and RC4
`283abaa401d60b4374a4b618a5a5f8a72ec26cb4`. Public stable binaries, stable
tag, and `main` remain at the qualified SHA. No production/runtime source
entered `main` after qualification. Repository Productization has not started.
Stop here.
