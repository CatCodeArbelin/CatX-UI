# CURRENT TASK

## Work Package

`Upstream Maintenance Strategy — COMPLETE`

This package defines and verifies the long-term, low-divergence maintenance
process for CatX-UI. It adds no product features, performs no production sync,
and must stop before any real upstream upgrade or CatX release.

## Baseline

- Starting `main`: `8f63afc6f1fdbac0f50d3bfd4f6f0bb8da4255fb`.
- Stable product/runtime: `v0.1.0` at `fd28ea7144147d9164b70810d4a24872a3d48b4f`.
- Recorded upstream base: `MHSanaei/3x-ui v3.8.5`.
- Bundled Xray: `26.9.9`.
- Actual upstream stable discovered during this package: `v3.9.0` at
  `3cd4bf504c3cd8ea9b1c1fdb032a9796c5c43ddb`.
- Strategy branch: `feature/upstream-maintenance-strategy`.

## Required outcome

Use `docs/04_UPSTREAM_SYNC.md` as the canonical operator procedure. Maintain
the explicit upstream/downstream remote and branch model, release discovery,
merge-based sync flow, conflict classes, sensitive-path review, CatX
touchpoints, migration/recovery policy, feature-off compatibility, and
PATCH/MINOR/MAJOR plus RC rules. `scripts/upstream-maintenance-report.sh` is
advisory only and must never merge, publish, or deploy.

## Dry-run boundary

Assess `v3.9.0` only on disposable `sync/upstream-v3.9.0`. Do not merge that
branch into `develop` or `main`, do not resolve a substantial upgrade here,
and make the actual upstream upgrade a separate work package.

## Completion record

Final strategy SHA: `11a5892492bbe93e0b7ed38d41a2deff6a3965dc`.
Strategy implementation commit: `b68daf5fc7187d86a63aa2ec0288d343e19ebfd7`.

The non-production `sync/upstream-v3.9.0` branch was created from `main` at
`8f63afc6f1fdbac0f50d3bfd4f6f0bb8da4255fb`. Merging exact upstream tag
`v3.9.0` produced 12 conflicts and was aborted without a merge commit. The
conflicts covered release/updater workflow and scripts, frontend dependency
manifests, Go dependencies, panel/update code, client/inbound traffic code,
and web lifecycle code. The upstream delta contained 563 changed paths, 224
classified sensitive paths, and 64 known CatX touchpoint overlaps.

Verification completed: upstream metadata/tag fetch, advisory report via Git
Bash, Bash syntax check for the new report, and `git diff --check` passed.
`make`, `make verify`, and `make verify-fork` were not run because `make` is
unavailable on this Windows host; no full verification result is claimed.

No merge SHA exists: the dry-run merge was intentionally aborted. `main`
remains `8f63afc6f1fdbac0f50d3bfd4f6f0bb8da4255fb`; `develop` remains
`ecd8ee1afbee4996f9a099d67014c4ab2236ae44`. The stable and RC tag targets
remain unchanged. No CatX release was created, no upstream upgrade entered
`main`, and no product feature was added.
