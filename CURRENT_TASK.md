# CURRENT TASK

## Work Package

`CatX Sponsors Management & Product UX`

## Status

`COMPLETE — PR #10 MERGED; PRODUCT REALITY AUDIT READY FOR MAINTAINER REVIEW`

This is the first post-`v0.2.0` CatX product feature. It must not create or
modify a `v0.2.1` release, move an immutable tag, or change the published
`v0.2.0` source, artifacts, or release evidence.

## Baseline

- branch: `chore/catx-product-reality-audit`;
- starting branch: merged `origin/develop` at `6207a2d6fc94e91b96006d0ac68f350c355879a8`;
- starting `origin/main`: `070212b7c3ddf92cf303c4f352481a86f30d14b4`;
- starting `origin/develop`: `d324840c034c8e9dd2a02f14e1ab4ca0cecbd42b`;
- stable tag: `v0.2.0 → 2b1760e98e665bde388e94c420bd22aa182fe2b0`;
- upstream base: `MHSanaei/3x-ui v3.9.0`;
- Xray: `26.9.30`.

## Goal

Turn the CatX Sponsors provider into a practical operator-managed product. The
CatX database is authoritative in explicit `local` provider mode; the secure
PR #9 HTTPS feed remains available only in explicit `remote` mode. The admin
surface must support safe CRUD, scheduling, deterministic ordering,
localization, preview, provider status, and reversible feature-off behavior.

The upstream Sponsors implementation was inspected before coding. Its useful
shape and visual components may be reused through narrow adapters, but CatX
owns the route registration, configuration, fetch/cache policy, validation,
and feature behavior.

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
- `docs/33_CATX_SPONSORS.md`

## Scope

- add a minimal fork-owned `fork_sponsors` model with explicit SQLite and
  PostgreSQL migration coverage;
- make local DB sponsors the default source of truth, with explicit local or
  remote provider mode and no ambiguous merge behavior;
- add authenticated fork CRUD/status/provider APIs, server-side validation,
  deterministic ordering, active-window filtering, safe remote logo URL
  proxying, and metadata-only audit events;
- add a CatX-owned management page under `frontend/src/forkext/sponsors/` with
  list/create/edit/enable-disable/delete, scheduling, slots, localization,
  logo URL, destination URL, preview, and provider status;
- preserve only runtime-supported slots: dashboard, sidebar, and page;
- keep the existing production rendering path compatible with local and
  explicit remote providers; feature-off remains storage-preserving and
  network-free;
- document the source-of-truth decision, schema, migration, API, security,
  audit, restart behavior, tests, and deferred import follow-up.

## Non-goals

- the management PR was not merged during its implementation phase; it was
  merged only after the exact-SHA acceptance gate passed;
- no `v0.2.1` release, tag, Docker publication, or stable artifact change;
- no payments, billing, invoicing, ad bidding, tracking pixels, targeting,
  CRM, affiliate, or self-service sponsor purchasing;
- no silent remote/local merging and no automatic destructive remote import;
- no copy of `internal/web/service/panel/sponsor.go`;
- no decrypted traffic, cookies, credentials, or sensitive HTTP bodies;
- no unrelated upstream refactor, generated-file hand edit, or repository-wide
  formatting.

## Safety and recovery

Feature disabled returns no rendering, remote fetch, import, logo fetch, or
background sync while retaining local records. Local CRUD is transactional and
validated; remote mode retains the PR #9 bounded HTTPS cache and SSRF policy.
Logo URLs are validated at save time and fetched only through the safe client.
Recovery is configuration-level: disable `sponsors.enabled`, select `local`, or
clear provider URLs, then reload the panel. No Xray, routing, or upstream
database state is changed.

## Verification gate

Before requesting review, run SQLite and PostgreSQL migration/CRUD tests,
restart/reload coverage, audit coverage, frontend tests/typecheck, generated
OpenAPI/docs checks, `git diff --check`, `make verify`, and `make verify-fork`
where the required toolchain is available. Record unavailable local checks and
rely on hosted required checks. The management PR merge gate is historical;
this continuation is the post-merge product reality audit.

## Completion condition

Implementation, tests, documentation, migration evidence, feature-disabled
compatibility, and restart behavior are complete; PR #10 is accepted and
merged; and the post-Sponsors product reality audit records evidence,
limitations, and the v0.3.0 scope without adding major features.

## Completion record

- implementation commit: `fc010fb701089f3eed8bebb7d85743e19c1164f6`;
- hosted lint correction: `be6bce6677bc0d50a349db138a0b7229aa576a2e`;
- historical review PR: [#10](https://github.com/CatCodeArbelin/CatX-UI/pull/10),
  merged into `develop` at `6207a2d6fc94e91b96006d0ac68f350c355879a8` after the
  hosted acceptance gates passed;
- hosted required Go, race, frontend, PostgreSQL, fuzz, codegen, build,
  release-identity, release qualification, and `verify-fork` gates are green
  for the implementation branch; disposable install/PostgreSQL rehearsal and
  rolling-release publication are skipped by repository policy;
- local limitations recorded during verification: CGO/race and SQLite-backed
  tests require the unavailable local C toolchain; Node 26 was used for the
  Vitest/OpenAPI tools; the hosted matrix supplied the full required gate;
- no `v0.2.1` release, tag movement, or stable artifact change.

## Post-Sponsors Product Reality Audit — ACCEPTED INTO DEVELOP

- audit branch: `chore/catx-product-reality-audit`;
- audited develop SHA: `6207a2d6fc94e91b96006d0ac68f350c355879a8`;
- accepted Sponsors Management head: `3cc4e111fcf6d9161f02d540cbf07a37e271388f`;
- PR #10 merge SHA: `6207a2d6fc94e91b96006d0ac68f350c355879a8`;
- audit scope: actual integrated upstream 3x-ui v3.9.0 plus CatX Policy,
  Insight, Traffic Control, Operations, Risk, Audit, Portal, Fleet, and
  Sponsors surfaces;
- audit output: `docs/35_CATX_PRODUCT_REALITY_AUDIT.md`;
- no major feature work, release work, upstream sync, tag movement, or main
  change is authorized in this phase;
- audit PR: [#11](https://github.com/CatCodeArbelin/CatX-UI/pull/11), merged into
  `develop` with normal merge commit `49a29ad994fe9928c711d9306739ffe161165ce5`;
- hosted Docs CI passed for audit head
  `0ff542b686783762c0eedcc53060f88cf992226b`.

## Sponsors Real Panel Lifecycle Proof — CURRENT

- lifecycle branch: `chore/sponsors-real-restart-smoke`;
- starting develop SHA: `49a29ad994fe9928c711d9306739ffe161165ce5`;
- scope: disposable hosted Linux panel, authenticated Sponsors CRUD/provider
  flow, real `POST /panel/api/setting/restartPanel`, feature-off/re-enable,
  provider persistence, Xray/panel health, and migration boundary evidence;
- harness: `scripts/staging/observe-sponsors-real-linux.sh`;
- hosted workflow: `.github/workflows/sponsors-real-restart-smoke.yml`;
- no product feature or runtime source change is authorized by this package;
- no multi-node qualification, release work, upstream sync, or `main` change.
