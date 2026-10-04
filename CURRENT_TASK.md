# CURRENT TASK

## Work Package

`CatX v0.3 Product Hardening`

## Status

`CURRENT — implementation and qualification in progress`

This package follows the completed Sponsors lifecycle proof. It is a product
coherence and qualification package, not a release. Human Review is the final
gate; passing automation does not authorize a merge into `main`, a tag, a
release candidate, or publication.

## Baseline

- base branch: `chore/sponsors-real-restart-smoke`;
- base SHA: `f2152a90992abbdfe402fd5694aa41cceeaddfe6`;
- preserved lifecycle evidence: `scripts/staging/observe-sponsors-real-linux.sh`
  and `.github/workflows/sponsors-real-restart-smoke.yml`;
- hardening branch: `feature/v0.3-product-hardening`;
- upstream base: `MHSanaei/3x-ui v3.9.0`;
- stable `v0.2.0` and `main` remain immutable and out of scope.

The remote fetch reported an existing local `dev-latest` tag collision and
refused to clobber that tag. Branch refs were resolved successfully; the
collision must remain visible in the final qualification report.

## Goal

Make the existing CatX product coherent, usable, localized in English and
Russian, behaviorally tested, and truthful about runtime capability. Reuse the
current upstream shell, CatX modules, database, Xray/runtime boundary, and
feature flags. Fix only gaps proven by the behavioral matrix or the maintainer
findings.

Required order:

1. UI/UX hardening;
2. complete CatX English and Russian localization;
3. serious unit, integration, API, restart, and real-panel browser coverage;
4. focused fixes for proven backend gaps;
5. full qualification;
6. exact candidate and maintainer review material;
7. stop pending Human Review.

## Required reading

- `AGENTS.md`
- `CURRENT_TASK.md`
- `TASK_QUEUE.md`
- `docs/00_MASTER_SPEC.md`
- `docs/01_ARCHITECTURE.md`
- `docs/04_UPSTREAM_SYNC.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/08_FRONTEND_UX.md`
- `docs/14_NON_GOALS.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`
- `docs/35_CATX_PRODUCT_REALITY_AUDIT.md`
- `docs/36_CATX_V0_3_PRODUCT_HARDENING.md`

## Scope and hard non-goals

In scope are the existing Analytics/Activity, DNS Intelligence, Policy,
Simulator/Explain, Schedules, Temporary Overrides, Traffic History, Traffic
Control/QoS, Group Quota, Risk, Audit, Webhooks, Metrics, Portal, Fleet,
Fleet Update, and Sponsors surfaces. The work must classify their actual
runtime state after save → real `POST /panel/api/setting/restartPanel` → panel
recovery and must distinguish feature-off, restart-required, initializing,
empty, unconfigured, unsupported, degraded, and real error.

The package must not add billing, payments, generic plugins, surveillance,
TLS MITM, HTTPS decryption, a new policy/analytics/traffic-control
architecture, a second source of truth, or new Sponsor scope. It must not
rewrite upstream 3x-ui or publish a release.

## Required product changes

- Move mutable Traffic/QoS controls into the existing client-edit workflow;
  keep client information primarily read-only and truthful about enforcement.
- Replace raw JSON as the normal Policy editor with structured controls for
  the supported `policy.Definition` schema. Preserve unknown valid fields in
  Advanced JSON when editing.
- Remove WP identifiers and implementation terminology from primary operator
  UI; retain technical detail only behind Advanced/diagnostic presentation.
- Replace raw Portal subject/client/host ID entry with existing searchable
  entity selectors where feasible and localize labels and validation.
- Make saved feature flags, pending restart, active runtime, unsupported
  capability, and runtime error visually distinct.
- Use one shared CatX state presentation pattern and keep empty/unconfigured
  states out of red error surfaces.
- Ensure all CatX-owned visible text is localized in English and Russian with
  no known mixed-language regressions. Other locales may use the repository’s
  English fallback strategy and are explicitly not expanded in this package.

## Verification gate

Before declaring the candidate ready for Human Review, run and record:

- `make verify` and `make verify-fork`;
- focused Go unit/API/restart tests, SQLite and PostgreSQL integration tests,
  race tests for changed concurrent code, and relevant fuzz/negative tests;
- frontend format, lint, typecheck, unit/component/browser tests, and build;
- real authenticated API and disposable real-panel browser coverage for login,
  settings/restart, Analytics, Policy CRUD/Simulator/Explain, client edit and
  read-only information, Portal, Sponsors, and RU/EN;
- generated OpenAPI/codegen drift, security/privacy, docs, and `git diff --check`;
- responsive visual review at 1920×1080 and 1366×768, including normal and
  collapsed sidebar, modal scrolling, tables, empty/error/degraded states,
  and Russian text expansion.

If a required environment is unavailable, record the exact check and reason;
do not claim it passed. Do not merge until the explicit Human Review gate is
recorded for release-facing frontend work.

## Recovery and release boundary

Config/runtime-affecting changes use snapshot → validate → apply → healthcheck
→ commit; failure restores the last known-good state, reloads/restarts as
needed, healthchecks, and records the failure. Feature-off behavior must be
storage-preserving and network-free where applicable.

At completion, leave the branch unmerged and provide an exact-SHA Draft PR or
compare URL, safe Windows PowerShell review setup and cleanup blocks, and a
maintainer checklist. Confirm `main`, stable tags, `v0.2.0`, and release
artifacts are unchanged. Stop with the exact required readiness or blocker
status from the work brief.

---

# Historical Sponsors Task Record

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
