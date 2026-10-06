# M00 Qualification Review and Test Contract

## Scope and evidence boundary

This document is the code-review and test-design record for:

`M00 — Core Lifecycle / Feature Settings`

The qualification candidate is isolated on `qualify/v0.3-m00-core`, based on
`efe647004aba5e6624edf02819c493db8246b377`. M00 tests shared lifecycle
semantics only. Later module packages may appear as passive lifecycle fixtures;
their business behavior is not asserted here.

Product code is frozen for this iteration. A failing test is baseline evidence,
not permission to repair the implementation.

## Production call-flow map

### Main startup / full process restart

```text
main.runWebServer
  → database.InitDB
      → initModels / upstream migrations + fork model inventory
      → forkext.RegisterMigrations
          → set settings DB
          → load persisted desired flags and validate dependencies
          → prepareRuntimeSchema
          → validateRuntimeSchema
          → publishPreparedRuntime
      → forkext.ConfigureRuntimeFromSettings
          → require prepared desired configuration
          → applyRuntimeConfig
          → configure/disable fork consumers
          → publishRuntimeSnapshot
  → web.Server.Start
      → router and authenticated fork routes
      → forkext.Start (background lifecycle consumers)
      → scheduler hooks
```

`database.InitDB` intentionally logs CatX preparation/activation failure and
continues the upstream panel startup. This is the optional-feature recovery
boundary; it must not be mistaken for a successful CatX activation.

### Feature Settings update

```text
authenticated API router
  → forkext.RegisterRoutes
      → PUT /fork/settings/features
          → Settings.UpdateFeatures (validate complete resulting state,
             persist requested desired flags transactionally)
          → policy pending marker when policies changes
          → PrepareRuntimeFromSettings
              → prepare + validate schema only
              → publishPreparedRuntime
          → FeatureFlags response
```

The update response can therefore be successful with
`state=restart_required`: desired persistence and explicit preparation have
completed, while in-process runtime activation still belongs to restart/apply.
Preparation failure returns a conflict/error response while leaving the
desired value available for deterministic retry.

### Generic `restartPanel` / SIGHUP

```text
POST /setting/restartPanel
  → SettingController.restartPanel
      → PanelService.RestartPanel (delayed asynchronous trigger)
          → global restart hook on Windows, or process SIGHUP on POSIX
              → main SIGHUP handler
                  → StopPanelOnly / stop fork background consumers
                  → new web.Server.StartPanelOnly
                      → ReloadRuntimeFromSettings
                          → read desired flags
                          → require already-prepared matching snapshot
                          → applyRuntimeConfig
                          → publishRuntimeSnapshot
                      → start panel and fork background consumers
                  → restart subscription server
```

`restartPanel` returns a pending acknowledgement. It does not synchronously
claim the final CatX state. `ReloadRuntimeFromSettings` is deliberately not a
migration entry point; a configuration without a prepared schema remains
non-active/error and does not create tables.

### Database import / restart

```text
POST /server/importDB
  → ServerService.ImportDB
      → validate/stage/restore or migrate uploaded database
      → database.InitDB on the restored database
          → M00 startup preparation + activation reconstruction
      → Xray/import recovery checks
  → schedule PanelService.RestartPanel
      → normal SIGHUP/restartPanel flow above
```

SQLite and PostgreSQL use their existing snapshot/recovery paths. M00 owns the
fork runtime reconstruction at the database boundary, not the database import
transaction or Xray recovery implementation itself.

### SIGHUP and USR1

SIGHUP is the production-relevant generic panel lifecycle and follows the
restart flow above. `SIGUSR1` only restarts Xray through the upstream service;
it is not a CatX feature-setting activation boundary and is outside M00 except
for observing that M00 does not silently migrate during it.

## Required lifecycle paths

| Path | Trigger | Desired persisted state | Preparation/schema state | Active runtime state | Reported API state | Failure and recovery |
|---|---|---|---|---|---|---|
| A | fresh startup, flags OFF | all false/default | startup boundary succeeds; no optional runtime activation | all consumers off | readable; every managed item `feature_off` | upstream panel remains usable |
| B | startup with persisted ON | selected true | startup prepares and validates selected schemas | active only after configure succeeds | `active` only after success | startup error leaves desired value and typed error/retry |
| C | OFF → desired ON | true before apply | explicit preparation required | remains off/pending until reload | `restart_required` or `error` | retry preparation, then reload |
| D | ON → desired OFF | false before apply | explicit preparation/validation of off config | old runtime may remain until apply | `restart_required` | reload disables the consumer |
| E | `restartPanel` | unchanged | no new generic migration | applies prepared desired config | request is `pending`; GET after apply is final | async failure is visible as non-active/error |
| F | full process restart | survives DB reopen | startup reconstruction | follows persisted desired state | truthful after startup | same startup retry/recovery boundary |
| G | SIGHUP | unchanged | prepared state only | reloads in-process consumers | no premature active claim | unprepared desired state remains non-active/error |
| H | database import/restart | imported values | InitDB/import recovery boundary | reconstructed, then normal restart | truthful post-restart | import rollback/recovery remains explicit |
| I | schema/preparation failure | desired retained | partial/no prepared state | prior known-good preserved if available | `error`, never healthy desired `active` | remove fault, rerun preparation |
| J | runtime activation failure | desired retained | prepared may exist | prior known-good preserved or inactive | `error` | deterministic retry/apply |
| K | malformed persisted flag | malformed row | cannot prepare that desired state | affected consumer inactive | item-level `error`; overall API readable | write a valid replacement |
| L | dependency violation | invalid combination rejected | no silent preparation | no invalid activation | request rejected or typed error | submit complete valid state |
| M | retry after activation failure | desired retained | preparation reruns idempotently | activates after fault removed | final `active` only after success | no manual recovery required |
| N | repeated apply | unchanged | no duplicate/corrupt DDL | stable active/off state | no semantic drift | idempotent no-op |

## Adversarial review findings

### Confirmed design facts

- Desired and active state are stored/reported separately.
- Feature Settings writes persist before runtime apply and use explicit
  preparation.
- Generic reload calls `ReloadRuntimeFromSettings`, which does not call a
  migrator.
- Startup/import use the approved preparation boundary before activation.
- Preparation errors are recorded without aborting unrelated panel startup.
- Dependency validation exists in the backend write path and in runtime config
  reconstruction.
- Malformed `strconv.ParseBool` values are handled item-by-item by
  `FeatureFlags`.
- The fork lifecycle uses existing upstream Xray/restart boundaries; it does
  not introduce a second Xray controller.

### Potential risks to preserve in tests

- A failed transition from an already-active known-good runtime intentionally
  preserves the prior consumer configuration while publishing `state=error`.
  Consumers that inspect only `active` could misunderstand this unless they
  honor the typed state as well.
- Preparation is ordered DDL without a universal transaction. A partial result
  is expected to remain and must be retryable; tests must record exact table
  state rather than assume rollback.
- The runtime snapshot and settings DB are process-global. Tests must remain
  serial and clean both state holders.
- `restartPanel` is asynchronous and the main handler owns the final result.
  A route-level HTTP 200 or pending response alone cannot prove activation.
- The panel's initial model inventory includes fork migration models before
  the feature-specific preparation hook. This is an intentional startup/import
  schema compatibility behavior in the current architecture; M00 tests the
  user-visible runtime-off semantics and the generic-reload migration boundary,
  not later module schema ownership.

### Expected design / outside M00

- Portal schema preparation is retained for historical compatibility even when
  the portal runtime is off.
- Webhooks and metrics are declared flags but are not managed runtime entries.
- Policy, analytics, traffic, risk, sponsor, and other module internals are
  passive fixtures only. Their business promises belong to M01–M11.
- Xray candidate validation/rollback remains an upstream/runtime recovery
  responsibility; M00 only observes the lifecycle handoff.

## Existing-test audit

| Test file / test name(s) | What it proves | Production boundary | Important mocks | Does not prove | Decision |
|---|---|---|---|---|---|
| `internal/forkext/settings_test.go` — `TestSettingsDefaultOffAndPersisted` | default-off key/value persistence and reopen | GORM settings table | none | runtime activation, API state, schema, restart | KEEP as unit/database support |
| same — `TestSettingsUnknownFlagRejected` | unknown flag rejection | Settings facade | none | lifecycle or API | KEEP |
| same — `TestUpdateFeaturesValidatesDependenciesAndPreservesState` | dependency validation and atomic desired-state preservation | Settings facade + DB | none | runtime state and restart | KEEP; map to T14 |
| same — `TestUpdateFeaturesRollsBackOnPersistenceFailure` | transaction rollback on write failure | real SQLite transaction | deterministic SQLite trigger | activation/recovery | KEEP |
| `internal/forkext/runtime_state_test.go` — `TestFeatureRuntimeStatusDistinguishesSavedAndActiveState` | pure state mapping for off/pending/active/error | pure helper | direct snapshot publication | persistence, schema, HTTP, startup | KEEP; strengthen with M00 IDs |
| `internal/forkext/feature_settings_test.go` — four route tests | JSON shape, desired persistence, malformed value repair, dependency rejection | real Gin route + settings DB + explicit preparation | SQLite only | restartPanel, full startup, PostgreSQL, runtime consumer behavior | KEEP and strengthen; map to T02/T13/T14/T17 |
| `internal/forkext/hooks_test.go` — `TestRuntimeActivationRequiresExplicitPreparation` | explicit prep before activation, schema checks, enable/disable/idempotent transitions | M00 lifecycle hooks + real module migrations/configure calls | none | actual process/startPanel handler and import | KEEP; map to T03/T04/T07/T08/T18/T19 |
| same — `TestRuntimeReloadDoesNotMigrateSponsorsOrLoseData` | generic reload does not repeat sponsor migration and preserves row | runtime reload hook + DB | migration counter seam | all feature schema boundaries, HTTP final state | KEEP; map to T07/T18 |
| same — `TestRuntimePreparationFailurePreservesKnownGoodRuntimeAndIsRetryable` | injected preparation failure preserves active runtime and retry succeeds | preparation/reload hooks | injected policy migrator | inactive-to-active failure, partial DB state, API | KEEP; map to T09/T11/T12 |
| same — `TestRuntimePreparationFailureAtEachMigrationBoundaryPreservesRuntime` | deterministic failure injection at ordered migration seams | preparation hooks + DB | injected migrators | exact partial schema state and API response | STRENGTHEN; map to T10 |
| same — `TestDisabledHooksAreExactNoOps` | nil/no-feature hook safety and Xray pointer identity | fork hook boundary | nil dependencies | persisted flags and real startup | KEEP; map to T01/T20 |
| `internal/forkext/hooks_postgres_test.go` — `TestRuntimeActivationPostgres` | PostgreSQL preparation, activation, retry behavior when DSN exists | real GORM PostgreSQL + M00 hooks | injected migrators | API/restart/import and exact cross-dialect comparison | STRENGTHEN; map to T16 |
| `internal/web/controller/import_db_restart_test.go` — `TestImportDBSchedulesPanelRestart` | successful import schedules a restart | real import handler + `PanelService` restart hook | Xray start/recovery hooks; global restart callback | CatX desired/active reconstruction and final runtime state | KEEP as import boundary; map to T05/T06/H |
| `internal/web/routes_contract_test.go` — `TestRouteRegistryContract` | registered API routes match frontend documentation | real router construction | audit enabled directly | lifecycle state or runtime activation | OUTSIDE M00; retain as route regression |
| `internal/policy/api_test.go` — policy runtime-state tests | policy module's own apply state | policy package | none | shared CatX lifecycle | OUTSIDE M00; passive fixture only |
| analytics/traffic/sponsors/portal/fleet tests | later-module business behavior | module-specific boundaries | module-specific | M00 contract | OUTSIDE M00; do not expand scope |

The audit conclusion is that existing tests cover many seams but do not form a
stable, explicitly named M00 acceptance suite. In particular, they lack a
single contract matrix, a full restart reconstruction proof, an item-level
malformed API proof after runtime error, a route-level pending/final restart
proof, and a clear SQLite/PostgreSQL equivalence mapping.

## Deterministic M00 test design matrix

| ID | Layer | Production entry point | Fixture | Observable assertion | Required DB | Mocks / seams | False-positive risk |
|---|---|---|---|---|---|---|---|
| M00-T01 | synthetic integration | `RegisterMigrations` → `ConfigureRuntimeFromSettings` → Feature Settings GET | `catx-m00-test`, all managed flags OFF | every CatX consumer inactive; API readable; all state `feature_off` | SQLite | none | testing only helper state would miss a consumer left enabled |
| M00-T02 | synthetic integration | `Settings.Set` then `FeatureFlags` | analytics desired ON, no apply | desired true, active false, `restart_required` | SQLite | none | row-only test could miss premature active state |
| M00-T03 | synthetic integration | `PrepareRuntimeFromSettings` → `ReloadRuntimeFromSettings` | analytics OFF → ON | schema exists; ACTIVE only after reload | SQLite | none | calling configure without preparation would bypass boundary |
| M00-T04 | synthetic integration | prepared reload after desired OFF | analytics active → OFF | pending before apply; inactive/feature_off after apply | SQLite | none | checking only persisted false would miss stale consumer |
| M00-T05 | runtime-adjacent integration | real `restartPanel` controller/service boundary plus restart hook | analytics prepared ON | response is pending; final state becomes active only after callback apply | SQLite | global restart hook is OS/process isolation seam | response-only test could pass with no activation |
| M00-T06 | runtime-adjacent integration | close/reopen file through startup registration/configuration | persisted analytics ON, fixed DB file | startup reconstructs desired and active state | SQLite | none | persistence-only test could miss runtime reconstruction |
| M00-T07 | synthetic integration | `ReloadRuntimeFromSettings` | analytics table absent, desired ON | reload errors and does not create table | SQLite | none | a test calling preparation first would hide migration leakage |
| M00-T08 | synthetic integration | `PrepareRuntimeFromSettings` | analytics table absent, desired ON | explicit preparation creates/validates schema | SQLite | none | direct `AutoMigrate` would bypass production seam |
| M00-T09 | synthetic integration | explicit preparation | analytics desired ON, injected analytics DDL error | inactive/error, readable settings API, desired retained | SQLite | analytics migrator error injection | only checking returned error could miss false ACTIVE |
| M00-T10 | synthetic integration | ordered preparation | analytics + policies ON; policy step fails after analytics | analytics table exists, policy table absent, exact error/retry state | SQLite | policy migrator error injection | assuming transaction rollback could hide partial DDL |
| M00-T11 | synthetic integration | retry via explicit preparation then reload | same T09 fixture, fault removed | successful active state after retry | SQLite | one-shot migrator fault | direct service configuration could bypass retry path |
| M00-T12 | synthetic integration | active runtime then failed new preparation | analytics active, policies desired ON, policy prep fails | known-good analytics remains running; desired policy is error/non-active | SQLite | policy migrator error injection | asserting all active=false would reject promised preservation |
| M00-T13 | API integration | Feature Settings GET | `analytics.enabled=not-a-boolean` | item-level error/inactive; overall semantic response succeeds | SQLite | none | HTTP 200 alone is insufficient |
| M00-T14 | unit + API integration | `UpdateFeatures` complete-state validation | DNS ON, analytics OFF | invalid state rejected atomically; no silent activation | SQLite | none | UI-only validation would not prove backend boundary |
| M00-T15 | database | full SQLite lifecycle | fixed `catx-m00-test` file | preparation/failure/retry semantics match contract | SQLite | deterministic DDL seam | passing only in-memory DB could miss reopen semantics |
| M00-T16 | database | same M00 hooks | fixed disposable PostgreSQL DSN | same desired/active/error/retry semantics as SQLite | PostgreSQL | deterministic DDL seams | skipped DSN must be BLOCKED, not PASS |
| M00-T17 | API integration | Feature Settings GET/PUT | off → desired ON → active → desired OFF | enabled, active, state, restartRequired are distinct/truthful | SQLite | none | asserting only status code misses state contract |
| M00-T18 | synthetic integration | repeated prepare/reload | analytics ON, repeat three times | one setting row, one schema, no state drift/errors | SQLite | none | one successful run cannot prove idempotence |
| M00-T19 | synthetic integration | repeated disabled prepare/reload | analytics OFF, repeat three times | stable no-op, inactive/feature_off | SQLite | none | only first disable could still leave stale runtime |
| M00-T20 | compatibility | nil hooks + Xray decorator | all flags OFF | exact no-op pointer/consumer behavior and safe lifecycle calls | none / SQLite | nil boundary only | test could pass while a hidden module remains active |

## Baseline result record

This section is filled only after the new suite is implemented and run against
the unchanged product. `SKIPPED` and `UNAVAILABLE` are recorded as `BLOCKED`,
never as `PASS`.

| Test ID | Result | Evidence / failure summary |
|---|---|---|
| M00-T01 | PENDING | |
| M00-T02 | PENDING | |
| M00-T03 | PENDING | |
| M00-T04 | PENDING | |
| M00-T05 | PENDING | |
| M00-T06 | PENDING | |
| M00-T07 | PENDING | |
| M00-T08 | PENDING | |
| M00-T09 | PENDING | |
| M00-T10 | PENDING | |
| M00-T11 | PENDING | |
| M00-T12 | PENDING | |
| M00-T13 | PENDING | |
| M00-T14 | PENDING | |
| M00-T15 | PENDING | |
| M00-T16 | PENDING | |
| M00-T17 | PENDING | |
| M00-T18 | PENDING | |
| M00-T19 | PENDING | |
| M00-T20 | PENDING | |

Totals: `PASS: pending`, `FAIL: pending`, `BLOCKED: pending`.

Confirmed product defects: pending baseline.

Testability blockers: pending baseline.
