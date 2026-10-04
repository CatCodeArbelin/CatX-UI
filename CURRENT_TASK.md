# CURRENT TASK

## Work Package

`CatX Sponsors`

## Status

`CURRENT`

This is the first post-`v0.2.0` CatX product feature. It must not create or
modify a `v0.2.1` release, move an immutable tag, or change the published
`v0.2.0` source, artifacts, or release evidence.

## Baseline

- branch: `feature/catx-sponsors`;
- starting branch: synchronized `origin/develop`;
- starting `origin/main`: `070212b7c3ddf92cf303c4f352481a86f30d14b4`;
- starting `origin/develop`: `b4f26cf2827c74592b6b733f0a2e9f24745622d0`;
- stable tag: `v0.2.0 → 2b1760e98e665bde388e94c420bd22aa182fe2b0`;
- upstream base: `MHSanaei/3x-ui v3.9.0`;
- Xray: `26.9.30`.

## Goal

Build a first CatX-owned Sponsors module with an authenticated panel API and a
localized frontend page/sidebar/dashboard presentation. The data source must
be explicitly configured by an operator, disabled by default, bounded, and
safe for remote metadata. The CatX module must not depend on or copy the
upstream Sanaei sponsor feed or source implementation.

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

- add the authoritative `sponsors.enabled` fork feature flag;
- add fork-owned backend configuration, bounded HTTPS metadata fetch/cache,
  active-window filtering, safe logo proxying, authenticated API routes, and
  tests under `internal/forkext/sponsors/`;
- persist only source/contact configuration through the existing settings
  mechanism; no new database table or runtime schema mutation;
- add a CatX-owned `/catx/sponsors` frontend route and registry entry;
- route existing authenticated sponsor slots through the CatX API;
- remove the unauthenticated login-page sponsor request;
- preserve RTL behavior and provide EN/RU/FA-compatible visible text;
- document the API/security contract and verification evidence.

## Non-goals

- no merge into `develop` from this branch;
- no `v0.2.1` release, tag, Docker publication, or stable artifact change;
- no copy of `internal/web/service/panel/sponsor.go`;
- no dependency on `sponsors.sanaei.dev`;
- no decrypted traffic, cookies, credentials, or sensitive HTTP bodies;
- no database migration unless implementation evidence proves the existing
  settings mechanism cannot represent the small configuration;
- no unrelated upstream refactor, generated-file hand edit, or repository-wide
  formatting.

## Safety and recovery

Feature disabled or unconfigured returns an empty sponsor set without a remote
request. Only bounded HTTPS metadata is accepted. Redirects, hosts, resolved
addresses, response sizes, logo names, content types, active dates, and link
schemes are validated. Cached known-good data remains available during a
temporary fetch failure; no network call is made while a service lock is held.
Configuration changes invalidate the cache. Recovery is configuration-level:
disable `sponsors.enabled` or clear the source URL, then reload fork settings;
the normal known-good empty/previous-cache behavior remains available without
altering Xray, routing, or database schema.

## Verification gate

Before requesting review, run relevant Go unit tests including race coverage,
frontend tests/typecheck, `git diff --check`, `make verify`, and
`make verify-fork` where the required toolchain is available. Record unavailable
local checks and rely on hosted required checks. Do not merge this Sponsors PR.

## Completion condition

Implementation, tests, documentation, and feature-disabled compatibility are
complete; the branch is pushed; hosted required checks are green; and an
unmerged PR targeting `develop` is open for review.
