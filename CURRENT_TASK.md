# CURRENT TASK

## Work Package

`Upstream Maintenance Strategy — CURRENT`

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

Record the dry-run result, changed sensitive paths, CatX touchpoint overlaps,
verification results, final strategy SHA, and merge SHA here before stopping.
