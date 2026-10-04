# CURRENT TASK

## Work Package

`CatX Sponsors Management & Product UX`

## Status

`CURRENT`

This is the first post-`v0.2.0` CatX product feature. It must not create or
modify a `v0.2.1` release, move an immutable tag, or change the published
`v0.2.0` source, artifacts, or release evidence.

## Baseline

- branch: `feature/catx-sponsors-management`;
- starting branch: merged `origin/develop` at `d324840c034c8e9dd2a02f14e1ab4ca0cecbd42b`;
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

- no merge of the management PR; it must remain a draft PR for maintainer
  review;
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
rely on hosted required checks. Do not merge this management PR.

## Completion condition

Implementation, tests, documentation, migration evidence, feature-disabled
compatibility, and restart behavior are complete; the branch is pushed; hosted
required checks are green; and a draft, unmerged PR targeting `develop` is
open for review.

## Completion record

- pending hosted verification and draft PR creation;
- no `v0.2.1` release, tag movement, or stable artifact change.
