# CURRENT TASK

## Work Package

`WP-9D — Product UX Integration & Human QA`

## RC-4 closure

RC-4 release engineering was technically qualified at exact SHA
`e780d9b87e7bb2fcb3cdeab6e33f11b893d9e68b` with Fork Verification
`36514812058`, Release CatX-UI `36514812043`, and Deploy Smoke Tests
`36514812068` all green. It was merged into `develop` with merge SHA
`b66ac15f79d411b5a0b2386e8770b05e88254bc2`.

The public `v0.1.0-rc.2` cut was intentionally deferred after manual smoke
testing found product UX and localization blockers. No public tag or release
was created. The RC-4 feature branch was deleted locally and remotely.

`main` remains untouched. WP-9D starts from the exact post-RC-4 `develop`
state on `feature/wp-9d-product-ux-integration`.

## Scope

Integrate CatX-owned frontend surfaces into the upstream 3x-ui information
architecture without changing backend semantics, database schema, Xray,
updater, release behavior, or upstream Sponsors implementation. Preserve all
direct CatX routes and keep fork modules low-divergence and independently
reviewable.

Required gates before this package can be considered ready for human review:

- frontend formatting, lint, typecheck, tests, generated checks;
- i18n parity, Russian leakage guards, RTL/accessibility and navigation tests;
- `make verify` and `make verify-fork`;
- disposable local preview instructions for manual light/dark/ultra-dark,
  responsive, RTL, and long-string review.

## WP-9D implementation evidence

- `npm run format:check`, `npm run lint`, `npm run typecheck`, and
  `git diff --check` pass locally.
- Vitest and the frontend build are not runnable in the local Node `20.12.0`
  environment: the repository requires Node `>=26.0.0`, and both commands
  stop before project code executes. No repository files, dependencies,
  lockfiles, or test configuration were changed for this limitation.
- Go tests are likewise deferred to canonical Linux verification: the local
  Windows run lacks the SQLite backup binding and CGO race support.
- Feature-off rendering is driven by explicit `featureDisabled` state where
  available, with only page-scoped handling for known fork-owned endpoints;
  ordinary HTTP 404 responses remain ordinary errors.

This package must stop before merge and before any RC-2 tag or release. Human
visual approval is required after the preview.
