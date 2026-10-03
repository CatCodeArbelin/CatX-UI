# CatX v0.2.0 Release Qualification Evidence

Status: `CURRENT`

This document records the release-specific evidence for CatX `v0.2.0`. General
upstream maintenance, rollback, and release governance remains in the
referenced repository documents; this file records exact candidate, workflow,
asset, observer, and publication results only.

## Baseline and branch

| Item                        | Value                                           |
| --------------------------- | ----------------------------------------------- |
| Release branch              | `feature/v0.2.0-release-qualification`          |
| Release baseline SHA        | `31a442cef876669aeae97ba9cb1e3e8888f42613`      |
| Starting `origin/develop`   | `31a442cef876669aeae97ba9cb1e3e8888f42613`      |
| Starting `origin/main`      | `7cb97e13def59280e92e5667d3f43ac23db8d470`      |
| Main/develop reconciliation | `origin/main` is an ancestor; no merge required |
| Previous CatX stable        | `v0.1.0`                                        |
| Previous stable runtime SHA | `fd28ea7144147d9164b70810d4a24872a3d48b4f`      |
| Upstream base               | `MHSanaei/3x-ui v3.9.0`                         |
| Upstream SHA                | `3cd4bf504c3cd8ea9b1c1fdb032a9796c5c43ddb`      |
| Upstream integration merge  | `c5f2e4e0165d96577af3f0e9d3b40810a98e7764`      |
| Xray                        | `26.9.30`                                       |
| Frozen RC candidate source  | `f9178025d79e8bc403cf7f3f77f83aab3028965a`      |

The local `dev-latest` tag differs from the remote tag and was not force-updated
or deleted. No `v0.2.0*` tag existed at package start.

## Release identity

| Field              | RC target                        | Stable target                    | Result                       |
| ------------------ | -------------------------------- | -------------------------------- | ---------------------------- |
| Fork version       | `0.2.0`                          | `0.2.0`                          | PASS — hosted identity gates |
| RC version         | `0.2.0-rc.1`                     | n/a                              | PASS — release qualification |
| Channel            | RC / prerelease                  | stable                           | RC PASS; stable pending      |
| Upstream version   | `3.9.0`                          | `3.9.0`                          | PASS — artifact metadata     |
| Xray version       | `26.9.30`                        | `26.9.30`                        | PASS — artifact metadata     |
| Release repository | `CatCodeArbelin/CatX-UI`         | `CatCodeArbelin/CatX-UI`         | PASS — CatX-owned            |
| Docker namespace   | `ghcr.io/catcodearbelin/catx-ui` | `ghcr.io/catcodearbelin/catx-ui` | RC PASS; stable pending      |

Authoritative version sources and every release/updater/install path must be
audited before the RC candidate is frozen. Official upstream release URLs,
assets, and container ownership are not acceptable in CatX runtime paths.

## Candidate and publication records

| Milestone                     | Source/tag/workflow                                                                                                                                                                                     | Result                                                                                             |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| RC source SHA                 | `f9178025d79e8bc403cf7f3f77f83aab3028965a`                                                                                                                                                              | PASS — frozen and publicly observed                                                                |
| RC pre-tag matrix             | [run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709)                                                                                                                   | PASS — non-publishing RC qualification                                                             |
| Annotated tag `v0.2.0-rc.1`   | `v0.2.0-rc.1 → f9178025d79e8bc403cf7f3f77f83aab3028965a`                                                                                                                                                | PASS — annotated, immutable, exact target                                                          |
| RC GitHub Release             | [v0.2.0-rc.1](https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.2.0-rc.1)                                                                                                                       | PASS — published prerelease                                                                        |
| RC release workflow           | [run 37149015516](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149015516)                                                                                                                   | PASS — full assets/staging                                                                         |
| RC Docker result              | [run 37149015510](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149015510)                                                                                                                   | PASS — attempt 4; multi-arch manifest; revision `f9178025d79e8bc403cf7f3f77f83aab3028965a`         |
| Actual-public RC observer     | [run 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602), [artifact 11284002718](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602/artifacts/11284002718) | PASS — published assets, clean Linux, public v0.1.0 upgrade/rollback, live VLESS, locale, TC/quota |
| Stable source SHA             | pending                                                                                                                                                                                                 | not frozen                                                                                         |
| Main merge SHA                | pending                                                                                                                                                                                                 | not merged                                                                                         |
| Final main SHA                | pending                                                                                                                                                                                                 | not established                                                                                    |
| Annotated tag `v0.2.0`        | pending                                                                                                                                                                                                 | not created                                                                                        |
| Stable GitHub Release         | pending                                                                                                                                                                                                 | not published                                                                                      |
| Stable release workflow       | pending                                                                                                                                                                                                 | not run                                                                                            |
| Stable Docker result          | pending                                                                                                                                                                                                 | not published                                                                                      |
| Actual-public stable observer | pending                                                                                                                                                                                                 | not run                                                                                            |
| Final evidence commit         | pending                                                                                                                                                                                                 | not committed                                                                                      |

Tags must be annotated, immutable, and point exactly to their qualified source
SHA. `v0.2.0-rc.1` must be fully observed from actual published assets before
stable qualification begins.

## Qualification matrix

Results use `PASS`, `FAIL`, `SKIPPED — reason`, or `PENDING`; a skipped check is
never converted to pass.

| Area                                                         | Result | Evidence                                                                                                                                                                                                                   |
| ------------------------------------------------------------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `make verify`                                                | PASS   | [CI run 37146928789](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928789)                                                                                                                                   |
| `make verify-fork`                                           | PASS   | [Fork verification 37146928674](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928674)                                                                                                                        |
| Go tests                                                     | PASS   | [CI run 37146928789](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928789)                                                                                                                                   |
| Go race tests                                                | PASS   | [CI run 37146928789](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928789)                                                                                                                                   |
| SQLite tests                                                 | PASS   | [CI run 37146928789](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928789)                                                                                                                                   |
| PostgreSQL tests                                             | PASS   | [CI run 37146928789](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928789)                                                                                                                                   |
| Migration tests                                              | PASS   | [RC staging run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709)                                                                                                                           |
| `govulncheck`                                                | PASS   | [CI run 37146928789](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928789)                                                                                                                                   |
| `golangci-lint`                                              | PASS   | [CI run 37146928789](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928789)                                                                                                                                   |
| Fuzz smoke                                                   | PASS   | [CI run 37146928789](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928789)                                                                                                                                   |
| Frontend install/generated/lint/format/typecheck/tests/build | PASS   | [CI run 37146928789](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928789)                                                                                                                                   |
| Storybook                                                    | PASS   | [CI run 37146928789](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928789)                                                                                                                                   |
| Docs typecheck/lint/format/tests/build                       | PASS   | [Docs CI run 37146928683](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928683)                                                                                                                              |
| Release identity                                             | PASS   | [RC release run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709)                                                                                                                           |
| Updater transaction                                          | PASS   | [smoke run 37146928692](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146928692) and [RC staging](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709)                                           |
| Rollback                                                     | PASS   | [RC staging run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709)                                                                                                                           |
| Xray config/start/reload                                     | PASS   | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)                                                                                                                       |
| Docker build                                                 | PASS   | [public RC Docker run 37149015510](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149015510), job `111295304468`; GHCR manifest digest `sha256:cc7021156c971ac1f23d3b9728a5720afcd5a3d04e76ea62495efd4eaf762adf` |
| `git diff --check`                                           | PASS   | local verification                                                                                                                                                                                                         |

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

| Evidence                    | Result | Run / artifact                                                                                   |
| --------------------------- | ------ | ------------------------------------------------------------------------------------------------ |
| Empty SQLite migration      | PASS   | [RC staging run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709) |
| Empty PostgreSQL migration  | PASS   | [RC staging run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709) |
| `v0.1.0` SQLite upgrade     | PASS   | [RC staging run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709) |
| `v0.1.0` PostgreSQL upgrade | PASS   | [RC staging run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709) |
| Backup contents/integrity   | PASS   | [RC staging run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709) |
| Disposable restore          | PASS   | [RC staging run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709) |

## Installer, updater, rollback, and restart lifecycle

The actual CatX transaction must prove:

```text
download → checksum/identity → stage → backup → install → migrate
→ start → panel healthcheck → Xray healthcheck → commit
```

Failure must restore the previous binary, service, config, and database state
as required, then healthcheck the restored installation. The candidate must
never fetch official upstream `3x-ui` binaries.

| Evidence                               | Result | Run / artifact                                                                                                 |
| -------------------------------------- | ------ | -------------------------------------------------------------------------------------------------------------- |
| Installer compatibility                | PASS   | [RC staging run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709)               |
| v0.1.0 → public RC updater             | PASS   | [actual-public observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)       |
| RC candidate → v0.1.0 rollback         | PASS   | [RC staging run 37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709)               |
| Real enabled `restartPanel` lifecycle  | PASS   | [public RC restart regression 37150043169](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150043169) |
| Real disabled `restartPanel` lifecycle | PASS   | [public RC restart regression 37150043169](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150043169) |

## Runtime and protocol qualification

| Area                                | Result | Required evidence                                                                                       |
| ----------------------------------- | ------ | ------------------------------------------------------------------------------------------------------- |
| Xray config validation/start/reload | PASS   | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)    |
| Live synthetic VLESS                | PASS   | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)    |
| Subscription generation             | PASS   | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)    |
| Native TUIC                         | PASS   | [hosted CI / RC matrix 37149138545](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149138545) |
| AmneziaWG/WireGuard                 | PASS   | [hosted CI / RC matrix 37149138545](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149138545) |
| Multi-node                          | PASS   | [hosted CI / RC matrix 37149138545](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149138545) |
| Traffic Control                     | PASS   | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)    |
| Quota lifecycle                     | PASS   | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)    |
| Policy                              | PASS   | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)    |
| Activity/DNS                        | PASS   | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)    |

## Product, localization, and public observation

The narrow product-reality audit must confirm that Activity, DNS Intelligence,
Policy/Simulator/Explain, Traffic History, Group Quota, Traffic Control
capability, Risk, Audit, Webhooks, Portal, Fleet, and Fleet Update do not claim
unsupported runtime behavior. Generic Xray-user attribution remains
unsupported. Generic per-user kernel shaping remains capability-dependent and
must not be advertised as universal.

| Area                                | Result  | Evidence                                                                                                                                                                                                       |
| ----------------------------------- | ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Russian LTR                         | PASS    | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)                                                                                                           |
| Persian RTL                         | PASS    | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)                                                                                                           |
| Actual-public RC clean install      | PASS    | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)                                                                                                           |
| Actual-public RC v0.1.0 upgrade     | PASS    | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)                                                                                                           |
| Actual-public RC updater identity   | PASS    | [public RC observer 37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)                                                                                                           |
| Actual-public RC Docker             | PASS    | [Docker run 37149015510](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149015510) — tags `0.2.0-rc.1` and `v0.2.0-rc.1`, linux/amd64, arm64/v8, arm/v7, arm/v6, 386; OCI revision matches RC source |
| Actual-public stable smoke          | pending | public `v0.2.0` assets                                                                                                                                                                                         |
| Actual-public stable v0.1.0 upgrade | pending | public `v0.2.0` assets                                                                                                                                                                                         |

## Decisions and blockers

- `rc.2` is not planned and must not be created preemptively.
- A genuine public RC source/release/runtime/migration/updater/rollback defect
  blocks stable and requires a new qualified RC.
- Infrastructure-only failures must be classified and retried under existing
  workflow policy.
- Observation-harness-only defects may be fixed without changing immutable
  public assets; failed attempts remain recorded here.
- The pre-public branch-push observer attempt [37146925435](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37146925435)
  created no jobs because it used the superseded push-triggered workflow. The
  observer was changed to manual dispatch before the candidate was frozen; no
  release asset was published by that attempt.
- Non-publishing RC qualification [37147211709](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37147211709)
  passed artifact inspection, Linux install/upgrade/rollback rehearsal, and
  PostgreSQL migration staging.
- The first public RC restart-regression dispatch [37149838636](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149838636)
  stopped at an obsolete historical-version expectation before lifecycle
  execution. The observer-only expectation fix was committed as `a0b17880`;
  the same immutable RC assets then passed [37150043169](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150043169).
- Public observer attempts [37150492777](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150492777) and
  [37150666706](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150666706) exposed two harness-only omissions
  in the new public upgrade stage (checksum filename, then updater-module
  checksum). The immutable RC assets were unchanged; corrected observer run
  [37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602)
  passed the complete public upgrade path.
- Hosted race attempt 1 had a transient TUIC test failure and attempt 2
  repeated the same flaky `close called for canceled stream` assertion without
  a data-race report; the failed race job was rerun again and must remain part
  of the qualification audit trail.
- The first RC Docker attempt failed while the upstream Docker helper resolved
  the latest `mtg-multi` release through GitHub API; the next two attempts were
  cancelled as infrastructure hangs. Retry attempt 4 completed successfully
  without changing the immutable RC source or tag. The published GHCR aliases
  resolve to one manifest digest with five required Linux platforms and OCI
  revision `f9178025d79e8bc403cf7f3f77f83aab3028965a`.
- The optional Claude review workflow is not a release gate unless governance
  explicitly requires it; missing credentials must be recorded as tooling
  infrastructure rather than fabricated review evidence.

## Final record

- RC candidate SHA: `f9178025d79e8bc403cf7f3f77f83aab3028965a`;
- RC tag target: `v0.2.0-rc.1 → f9178025d79e8bc403cf7f3f77f83aab3028965a`;
- RC Release URL/workflow: [release](https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.2.0-rc.1), [workflow 37149015516](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149015516);
- RC Docker result: [run 37149015510](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37149015510), job `111295304468`, PASS; aliases share manifest `sha256:cc7021156c971ac1f23d3b9728a5720afcd5a3d04e76ea62495efd4eaf762adf` and OCI revision `f9178025d79e8bc403cf7f3f77f83aab3028965a`;
- public RC observer: [37150948602](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150948602), restart regression [37150043169](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/37150043169);
- stable candidate SHA: pending;
- main merge SHA/final main SHA: pending;
- stable tag target: pending;
- stable Release URL/workflow: pending;
- stable Docker result: pending;
- public stable observer: pending;
- actual public `v0.1.0 → v0.2.0` upgrade: pending;
- final evidence commit: pending.
