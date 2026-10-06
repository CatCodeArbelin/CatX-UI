# CatX v0.3 Module Qualification Roadmap

## Purpose

CatX v0.3 is being requalified module-by-module.

Previous implementation work and previous green CI runs do not automatically
mean that a module is release-ready.

Each CatX-added product module remains `DEVS` until it passes the complete
Module Qualification Protocol and receives explicit maintainer Human Review
approval.

## Technical baseline

Starting technical baseline:

`5b20b4af7fec6ad5a382e1f6a30022b8ed498d0f`

Architecture-correction functional gates passed on this baseline.

This baseline is not itself evidence that every individual module is
release-qualified.

## Qualification integration branch

Planned integration branch:

`qualification/v0.3-module-review`

Qualified module branches merge into this branch only after explicit Human
Review PASS.

No module qualification branch merges directly into `develop` or `main`.

## Product maturity

Allowed product maturity values:

- `DEVS`
- `READY`

`DEVS` means implemented but not release-qualified by the current protocol.

`READY` means all applicable automated gates and explicit maintainer Human
Review have passed.

Product maturity is independent from runtime states such as `ACTIVE`,
`DISABLED`, `ERROR`, `DEGRADED`, `INITIALIZING`, or `UNSUPPORTED`.

## Qualification workflow states

- `PENDING`
- `CODE_REVIEW`
- `TEST_SPEC`
- `AUTOMATED_TESTING`
- `REWORK`
- `LOCALIZATION`
- `UI_REVIEW`
- `WAITING_HUMAN_REVIEW`
- `HUMAN_REVIEW_FAILED`
- `READY`

## Program sequence

### Step 0 — DEVS marking — COMPLETE

All CatX-added modules that have not passed this protocol must be visibly marked
`DEVS`.

No module may lose `DEVS` without explicit maintainer approval.

The centralized frontend maturity registry and visible markers are implemented
and verified on `qualification/v0.3-module-review`. M00–M11 remain `DEVS`.

### Module loop

For each module:

1. Code Review
2. Test Contract / Acceptance Specification
3. Automated Baseline
4. Fix / Retest Loop
5. en-US / ru-RU Localization Review
6. UI / CSS / Product Review
7. Full Module Regression
8. Human Review
9. PASS → READY / merge to qualification integration branch
10. FAIL → REWORK and repeat until PASS

## Module status

The table below is reconciled against the current implementation inventory.
The implementation remains authoritative for future additions or boundary
changes.

| ID | Module | Maturity | Workflow | Automated | Human Review | Merge |
|---|---|---|---|---|---|---|
| M00 | Core Lifecycle / Feature Settings | DEVS | AUTOMATED_TESTING | PASS (20/20) | PENDING | NO |
| M01 | Analytics / Activity | DEVS | PENDING | PENDING | PENDING | NO |
| M02 | DNS Intelligence | DEVS | PENDING | PENDING | PENDING | NO |
| M03 | Policy Engine | DEVS | PENDING | PENDING | PENDING | NO |
| M04 | Traffic History / Quotas | DEVS | PENDING | PENDING | PENDING | NO |
| M05 | Traffic Control / QoS | DEVS | PENDING | PENDING | PENDING | NO |
| M06 | Security / Risk | DEVS | PENDING | PENDING | PENDING | NO |
| M07 | Audit / Webhooks / Metrics | DEVS | PENDING | PENDING | PENDING | NO |
| M08 | Portal / Self-service | DEVS | PENDING | PENDING | PENDING | NO |
| M09 | Fleet | DEVS | PENDING | PENDING | PENDING | NO |
| M10 | Fleet Updates | DEVS | PENDING | PENDING | PENDING | NO |
| M11 | Sponsors | DEVS | PENDING | PENDING | PENDING | NO |

### Implementation inventory reconciliation

The expected M00–M11 sequence is retained because it matches the current
product boundaries and dependency order. The following implementation details
are intentional grouping notes, not additional maturity modules:

- M00 is the fork-owned lifecycle/settings boundary in `internal/forkext`,
  including the authoritative settings, runtime-state, migration, route, and
  hook integration code. It is not a standalone product page.
- M01 and M02 share `internal/analytics`, but remain separate qualification
  modules because activity and DNS evidence have distinct promises and flags.
  DNS evidence additionally depends on analytics activation.
- M03 includes `internal/policy`, `internal/policycompiler`, and
  `internal/policysim`; the simulator/compiler are part of the Policy Engine
  contract rather than separate products.
- M04 combines analytics traffic history with
  `internal/forkext/groupquota` and `internal/forkext/trafficpolicy` because
  history, quota accounting, reset/window behavior, and persistence form one
  observable accounting contract.
- M05 is the separate `internal/forkext/trafficcontrol` enforcement and
  capability boundary. Unsupported shaping must not be represented as active
  enforcement.
- M07 is implemented through `internal/forkext/audit`, whose API and delivery
  surfaces include audit, webhooks, and metrics. `webhooks.enabled` and
  `metrics.enabled` are declared in the authoritative flag inventory but are
  not currently independent managed runtime flags; they do not create extra
  qualification rows.
- M08 is the fork portal implementation over existing client/host identity;
  its schema is prepared for compatibility while active self-service behavior
  is independently controlled.
- M09 is the CatX Fleet page over the upstream Nodes model. There is no
  second fleet/node inventory or separate fork fleet control plane.
- M10 is the independent `internal/forkext/fleetupdate` campaign and
  reconciliation module, with mutation capability separately constrained by
  the authoritative settings implementation.
- M11 is the CatX-owned Sponsors API and management surface under
  `internal/forkext/sponsors` and `frontend/src/forkext/sponsors`.

## Current program position

Current stage:

`MODULE QUALIFICATION`

Current module:

`M00 — Core Lifecycle / Feature Settings`

Current workflow:

`AUTOMATED_TESTING`

Current maturity:

`DEVS`

Next action:

M00 localization review.

Step 0 is complete. M00 code review and test contract are frozen. The hosted
M00 baseline passed 20/20 tests with no product fixes; Human Review is still
pending and M00 remains `DEVS`.

## Human Review authority

Only the maintainer may approve a Human Review.

Playwright, screenshots, CI, AI inspection, or browser automation do not
constitute Human Review approval.

## Human Review failure

Every Human Review finding receives a stable ID:

`HR-Mxx-NNN`

A failed module remains `DEVS`.

The module returns to `REWORK`.

The code and tests must both be reviewed to determine why automated
qualification failed to detect the human-visible defect.

## Cross-module blockers

A module qualification must not silently expand into another module.

Cross-module defects are recorded as explicit blockers and corrected through a
separate shared/core task before the blocked module continues.

## Global release blockers

### Frontend dependency audit

Current inherited dependency findings:

- total: 6
- moderate: 4
- critical: 2
- introduced by current CatX architecture work: NO
- known non-breaking remediation: none currently proven
- full automated remediation currently requires a breaking upgrade

This is a GLOBAL RELEASE BLOCKER.

It does not automatically prevent an unrelated individual module from becoming
`READY`.

It must be resolved or explicitly dispositioned before final CatX v0.3 release
qualification.

## Final program gate

When every module is `READY`:

1. run complete whole-product integration regression;
2. run feature-off compatibility;
3. run SQLite/PostgreSQL upgrade/recovery;
4. run real Xray/runtime qualification;
5. run frontend integration;
6. resolve mandatory global release blockers;
7. only then consider merge to `develop`;
8. perform separate RC/release qualification.
