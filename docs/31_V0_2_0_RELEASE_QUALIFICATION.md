# CatX v0.2.0 Release Qualification Evidence

Status: `CURRENT`

This document records the release-specific evidence for CatX `v0.2.0`. General
upstream maintenance, rollback, and release governance remains in the
referenced repository documents; this file records exact candidate, workflow,
asset, observer, and publication results only.

## Baseline and branch

| Item | Value |
| --- | --- |
| Release branch | `feature/v0.2.0-release-qualification` |
| Release baseline SHA | `31a442cef876669aeae97ba9cb1e3e8888f42613` |
| Starting `origin/develop` | `31a442cef876669aeae97ba9cb1e3e8888f42613` |
| Starting `origin/main` | `7cb97e13def59280e92e5667d3f43ac23db8d470` |
| Main/develop reconciliation | `origin/main` is an ancestor; no merge required |
| Previous CatX stable | `v0.1.0` |
| Previous stable runtime SHA | `fd28ea7144147d9164b70810d4a24872a3d48b4f` |
| Upstream base | `MHSanaei/3x-ui v3.9.0` |
| Upstream SHA | `3cd4bf504c3cd8ea9b1c1fdb032a9796c5c43ddb` |
| Upstream integration merge | `c5f2e4e0165d96577af3f0e9d3b40810a98e7764` |
| Xray | `26.9.30` |
| Frozen RC candidate source | `ec3eb8cd535b32ca78f9c2edb73e7f50880e281f` |

The local `dev-latest` tag differs from the remote tag and was not force-updated
or deleted. No `v0.2.0*` tag existed at package start.

## Release identity

| Field | RC target | Stable target | Result |
| --- | --- | --- | --- |
| Fork version | `0.2.0` | `0.2.0` | pending |
| RC version | `0.2.0-rc.1` | n/a | pending |
| Channel | RC / prerelease | stable | pending |
| Upstream version | `3.9.0` | `3.9.0` | pending |
| Xray version | `26.9.30` | `26.9.30` | pending |
| Release repository | `CatCodeArbelin/CatX-UI` | `CatCodeArbelin/CatX-UI` | pending |
| Docker namespace | `ghcr.io/catcodearbelin/catx-ui` | `ghcr.io/catcodearbelin/catx-ui` | pending |

Authoritative version sources and every release/updater/install path must be
audited before the RC candidate is frozen. Official upstream release URLs,
assets, and container ownership are not acceptable in CatX runtime paths.

## Candidate and publication records

| Milestone | Source/tag/workflow | Result |
| --- | --- | --- |
| RC source SHA | `ec3eb8cd535b32ca78f9c2edb73e7f50880e281f` | frozen; qualification pending |
| RC pre-tag matrix | pending | not run |
| Annotated tag `v0.2.0-rc.1` | pending | not created |
| RC GitHub Release | pending | not published |
| RC release workflow | pending | not run |
| RC Docker result | pending | not published |
| Actual-public RC observer | pending | not run |
| Stable source SHA | pending | not frozen |
| Main merge SHA | pending | not merged |
| Final main SHA | pending | not established |
| Annotated tag `v0.2.0` | pending | not created |
| Stable GitHub Release | pending | not published |
| Stable release workflow | pending | not run |
| Stable Docker result | pending | not published |
| Actual-public stable observer | pending | not run |
| Final evidence commit | pending | not committed |

Tags must be annotated, immutable, and point exactly to their qualified source
SHA. `v0.2.0-rc.1` must be fully observed from actual published assets before
stable qualification begins.

## Qualification matrix

Results use `PASS`, `FAIL`, `SKIPPED — reason`, or `PENDING`; a skipped check is
never converted to pass.

| Area | Result | Evidence |
| --- | --- | --- |
| `make verify` | pending | hosted exact-SHA run |
| `make verify-fork` | pending | hosted exact-SHA run |
| Go tests | pending | hosted exact-SHA run |
| Go race tests | pending | hosted exact-SHA run |
| SQLite tests | pending | hosted exact-SHA run |
| PostgreSQL tests | pending | hosted exact-SHA run |
| Migration tests | pending | fresh and populated upgrade matrix |
| `govulncheck` | pending | hosted exact-SHA run |
| `golangci-lint` | pending | hosted exact-SHA run |
| Fuzz smoke | pending | canonical workflow where applicable |
| Frontend install/generated/lint/format/typecheck/tests/build | pending | hosted exact-SHA run |
| Storybook | pending | hosted exact-SHA run |
| Docs typecheck/lint/format/tests/build | pending | hosted exact-SHA run |
| Release identity | pending | release identity tests |
| Updater transaction | pending | v0.1.0 → RC candidate |
| Rollback | pending | RC candidate → v0.1.0 |
| Xray config/start/reload | pending | Xray `26.9.30` |
| Docker build | pending | non-publishing and public artifact checks |
| `git diff --check` | pending | local and hosted verification |

## Database, backup, and restore

Required paths:

- empty SQLite → RC candidate;
- empty PostgreSQL → RC candidate;
- CatX `v0.1.0` SQLite → RC candidate;
- CatX `v0.1.0` PostgreSQL → RC candidate;
- backup integrity and disposable restore of the previous known-good state.

Preservation must cover users, inbounds, clients, traffic counters, renewal
fields, subscriptions, CatX flags, Analytics, DNS, Policy, Traffic Policy,
Group Quota, Audit, Fleet, and Portal state. New upstream migration fields
must be present without destructive reset or silent CatX-state loss.

| Evidence | Result | Run / artifact |
| --- | --- | --- |
| Empty SQLite migration | pending | pending |
| Empty PostgreSQL migration | pending | pending |
| `v0.1.0` SQLite upgrade | pending | pending |
| `v0.1.0` PostgreSQL upgrade | pending | pending |
| Backup contents/integrity | pending | pending |
| Disposable restore | pending | pending |

## Installer, updater, rollback, and restart lifecycle

The actual CatX transaction must prove:

```text
download → checksum/identity → stage → backup → install → migrate
→ start → panel healthcheck → Xray healthcheck → commit
```

Failure must restore the previous binary, service, config, and database state
as required, then healthcheck the restored installation. The candidate must
never fetch official upstream `3x-ui` binaries.

| Evidence | Result | Run / artifact |
| --- | --- | --- |
| Installer compatibility | pending | pending |
| v0.1.0 → public RC updater | pending | actual published assets |
| RC candidate → v0.1.0 rollback | pending | actual previous release |
| Real enabled `restartPanel` lifecycle | pending | persisted flags and active runtimes |
| Real disabled `restartPanel` lifecycle | pending | persisted flags and inactive runtimes |

## Runtime and protocol qualification

| Area | Result | Required evidence |
| --- | --- | --- |
| Xray config validation/start/reload | pending | generated config plus live Xray |
| Live synthetic VLESS | pending | actual proxy traffic and counters |
| Subscription generation | pending | standard/JSON/Clash/Happ/TUIC as applicable |
| Native TUIC | pending | config, auth, live data, traffic, restart, shutdown |
| AmneziaWG/WireGuard | pending | focused changed-area smoke |
| Multi-node | pending | fanout, reconnect, reset, capability state |
| Traffic Control | pending | capability, apply, reconcile, remove, no-clobber, rollback |
| Quota lifecycle | pending | group quota and fixed/rolling windows where implemented |
| Policy | pending | decoration and representative decision flow |
| Activity/DNS | pending | evidence and feature-off behavior |

## Product, localization, and public observation

The narrow product-reality audit must confirm that Activity, DNS Intelligence,
Policy/Simulator/Explain, Traffic History, Group Quota, Traffic Control
capability, Risk, Audit, Webhooks, Portal, Fleet, and Fleet Update do not claim
unsupported runtime behavior. Generic Xray-user attribution remains
unsupported. Generic per-user kernel shaping remains capability-dependent and
must not be advertised as universal.

| Area | Result | Evidence |
| --- | --- | --- |
| Russian LTR | pending | navigation, CatX pages, settings, forms |
| Persian RTL | pending | navigation, CatX pages, settings, forms |
| Actual-public RC clean install | pending | public `v0.2.0-rc.1` assets |
| Actual-public RC v0.1.0 upgrade | pending | public `v0.2.0-rc.1` assets |
| Actual-public RC updater identity | pending | CatX owner/version checks |
| Actual-public RC Docker | pending | namespace, manifest, OCI revision |
| Actual-public stable smoke | pending | public `v0.2.0` assets |
| Actual-public stable v0.1.0 upgrade | pending | public `v0.2.0` assets |

## Decisions and blockers

- `rc.2` is not planned and must not be created preemptively.
- A genuine public RC source/release/runtime/migration/updater/rollback defect
  blocks stable and requires a new qualified RC.
- Infrastructure-only failures must be classified and retried under existing
  workflow policy.
- Observation-harness-only defects may be fixed without changing immutable
  public assets; failed attempts remain recorded here.
- The optional Claude review workflow is not a release gate unless governance
  explicitly requires it; missing credentials must be recorded as tooling
  infrastructure rather than fabricated review evidence.

## Final record

- RC candidate SHA: `ec3eb8cd535b32ca78f9c2edb73e7f50880e281f`;
- RC tag target: pending;
- RC Release URL/workflow: pending;
- RC Docker result: pending;
- public RC observer: pending;
- stable candidate SHA: pending;
- main merge SHA/final main SHA: pending;
- stable tag target: pending;
- stable Release URL/workflow: pending;
- stable Docker result: pending;
- public stable observer: pending;
- actual public `v0.1.0 → v0.2.0` upgrade: pending;
- final evidence commit: pending.
