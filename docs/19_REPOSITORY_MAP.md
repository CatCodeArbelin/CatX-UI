# Repository Map

This map records the checked-out repository at commit
`b3d6401f60ed579050521f58057264694b0007ea` on branch
`feature/task-001-repository-map`. The embedded panel version is `3.8.5`.
It is an analysis baseline for later fork work, not an implementation plan that
authorizes production changes.

The principal architectural constraint is to preserve the upstream execution
paths and add a small number of fixed, first-party fork hooks. A general plugin
framework would create more divergence than the planned features require.

## Current upstream architecture

- The backend is a Go application built around Gin, GORM, robfig/cron, and an
  externally executed Xray process. SQLite and PostgreSQL are supported.
- The frontend is React 19, TypeScript, Ant Design 6, and Vite 8. It is built
  into the Go binary and served as a single-page application.
- `main.go` owns process-level startup, signal handling, and panel/subscription
  server restarts.
- `internal/web/web.go` is the application composition root. `web.Server` owns
  controllers, the HTTP server, `XrayService`, settings, bots, the WebSocket
  hub, event bus, cron scheduler, and a cancellable context.
- `internal/web/runtime` abstracts local and remote Xray mutations. Local
  changes are applied through the local Xray API and sidecars; remote changes
  are directed to nodes.
- `internal/database` owns database selection, model migration, seeders,
  backups, and cross-engine migration helpers.
- `internal/xray` owns typed configuration fragments, Xray process management,
  gRPC API calls, traffic structures, and hot-diff calculation.
- `internal/web/service` contains most application behavior. Controllers are
  thin adapters around these services.
- The current source already contains normalized clients, client groups,
  local/remote nodes, node traffic, node IP observations, global traffic,
  outbound subscriptions, live metrics, and an internal event bus. These must
  be extended rather than recreated.

The repository has no generic extension framework. The safest downstream
shape is one small neutral hook contract, implemented only by first-party fork
modules and installed explicitly from `main.go`. Hook interfaces must not
import `internal/web/service`, otherwise importing the hook package from those
services would create cycles.

## Application lifecycle

`main.go` follows this normal path:

1. `main()` loads environment configuration and handles CLI modes.
2. `runWebServer()` initializes logging and token encryption.
3. It calls `database.InitDB(config.GetDBPath())`.
4. It creates `web.NewServer()` and calls `Server.Start()`.
5. It starts the separate subscription server.
6. It waits for signals. `SIGHUP` restarts the panel/subscription servers,
   `SIGUSR1` restarts Xray, and termination signals stop the running services.

`internal/web/web.go` supplies the application-level lifecycle:

- `NewServer()` constructs services and shared facilities.
- `initRouter()` installs middleware, controllers, WebSocket routes, SPA
  delivery, and static resources.
- `Server.start()` starts the traffic writer, creates a cron scheduler with
  `SkipIfStillRunning` and panic recovery, starts runtime management, starts
  Gin, starts the event bus, registers subscribers and jobs, and starts bots.
- `startTask()` contains the current job registrations.
- `Server.stop()`/`Stop()` cancels the context; stops Xray and sidecars, cron,
  event delivery, metrics persistence, the traffic writer, bots, WebSockets,
  and HTTP.
- `RestartXray()` exposes the runtime restart path.

Fork services that own goroutines need both start and stop hooks bound to the
server context. Registration must occur before jobs/subscribers start, while
shutdown must finish before their dependencies are torn down. A hook must not
hold a global lock across filesystem, DNS, network, or database work.

## Database and migrations

`internal/database/db.go` is the current database integration point.
`InitDB()` selects PostgreSQL when configured by `XUI_DB_TYPE`/DSN; otherwise
it opens SQLite with WAL, busy-timeout, and connection settings. It then calls
model initialization and data seeders.

`allModels()` currently includes, among others:

- `model.User`, `model.Inbound`, `model.OutboundTraffics`, `model.Setting`, and
  `model.InboundClientIps`;
- `xray.ClientTraffic` and seeder history;
- `model.Node`, `model.NodeClientTraffic`, `model.NodeClientIp`, and
  `model.ClientGlobalTraffic`;
- `model.ClientRecord`, `model.ClientInbound`, `model.ClientHwid`,
  `model.ClientExternalLink`, and `model.ClientGroup`;
- `model.InboundFallback`, `model.Host`, `model.ApiToken`,
  `model.OutboundSubscription`, and `model.SubBalancer`.

`initModels()` combines per-model `AutoMigrate` with explicit compatibility
repairs and migrations. PostgreSQL has a `postgresModelSettled` optimization
and sequence resynchronization. `runSeeders()` records completed data migrations
in `HistoryOfSeeders`. Dialect-specific expressions and JSON helpers live in
`internal/database/dialect.go`.

There is no standalone, ordered migration directory or independent schema
version framework. Consequently, any fork persistence layer must provide an
explicit, idempotent migration sequence for both engines and integrate through
one database hook. Runtime schema mutation elsewhere is not acceptable.

`internal/database/db.go` is already a permanent upstream-touch hotspot because
it owns model discovery and migration ordering. The fork should add one call
there, not append each fork model and migration directly to upstream lists.

Settings currently use the key/value `model.Setting` table.
`internal/web/service/setting.go` combines an embedded/default configuration,
`defaultValueMap`, generic `getSetting()`/`saveSetting()` helpers, and reflected
population of `entity.AllSetting`. Adding every fork flag independently to
those upstream maps and DTOs would create recurring conflicts. A later task
should select a namespaced fork settings facade or fork-owned persisted
representation, default every large subsystem to disabled, and expose only the
small adapter required by existing settings APIs.

## Clients / Groups / Inbounds / Nodes

The current data model is more advanced than a JSON-only inbound client model:

- `model.ClientRecord` is the normalized client identity in the `clients`
  table.
- `model.ClientInbound` links clients to inbounds.
- `xray.ClientTraffic` stores per-client/per-inbound traffic state.
- Inbound settings JSON remains synchronized for Xray compatibility.
- `model.ClientGroup` and group membership support client grouping.
- `model.ClientHwid` and `model.ClientExternalLink` attach additional client
  state.

Client behavior is deliberately split across `internal/web/service/client*.go`.
`ClientService` CRUD is in `client_crud.go`, with separate lookup, links,
traffic, HWID, portable-data, group, and paging files. Group operations are
methods on `ClientService` in `client_groups.go`; there is no generic `Group`
domain service. Routes are implemented by `internal/web/controller/client.go`
and `group.go`, with group API routes nested below the client API.

`internal/web/service/inbound.go` owns `InboundService` and its core CRUD.
Protocol, node, and traffic concerns are split into adjacent files. Inbound
configuration must remain compatible with `Inbound.GenXrayInboundConfig()` and
the normalized client hydration performed by `XrayService`.

`model.Node`, `internal/web/service/node.go`, and
`internal/web/runtime/{manager,local,remote}.go` already provide node CRUD,
probe/update behavior, dirty-state synchronization, and local/remote execution.
Node traffic and IP observations have dedicated models. A new parallel node
abstraction would duplicate upstream behavior and make synchronization harder.

Policy assignment should therefore reference stable normalized client IDs,
existing group IDs, inbound IDs/tags, and node IDs. It should not introduce a
second client, group, inbound, or node source of truth.

## Xray config generation

`internal/web/service/xray.go`, `XrayService.GetXrayConfig()`, is the canonical
configuration compiler. In order, it:

1. loads and unmarshals the stored Xray template;
2. normalizes log paths;
3. ensures required Handler, Stats, and Routing API services;
4. ensures the online-user stats policy;
5. strips disabled routing UI fields/rules;
6. performs legacy-lift and AmneziaWG transformations;
7. loads enabled local inbounds, hydrates their clients from normalized records
   and traffic state, normalizes fallback/stream settings, and calls
   `Inbound.GenXrayInboundConfig()`;
8. merges outbound subscriptions;
9. injects MTProto, AmneziaWG relay/IPv6, panel-egress, and node-egress
   configuration; and
10. returns the final `xray.Config`.

`internal/xray/config.go` defines the configuration container using raw JSON for
many sections and typed inbound structures. Rule and outbound array order can
be semantically significant; compatibility comparisons must not sort them.

The lowest-divergence policy integration is one decorator call at the end of
`GetXrayConfig()`, after all existing injections and immediately before return.
The decorator must be deterministic, preserve the API routing rule, preserve
existing order unless a policy explicitly requires an insertion, and be an
exact no-op when fork features are disabled. It must not create a second config
generator.

`internal/web/service/xray_setting.go` validates JSON syntax and individual
outbounds when settings are saved, and ensures stats/DNS routing support. It is
not full candidate-config validation in an isolated Xray process.

## Xray runtime/API

`internal/xray/process.go` owns binary/config paths and the Xray child process.
`Process.Start()` serializes the generated config, atomically writes
`bin/config.json` with restrictive permissions, launches Xray with `-c`, and
monitors exit/crash state. Atomic file replacement prevents a partial config
file but is not a runtime transaction.

`internal/web/service/xray.go`, `RestartXray()`, locks restart operations,
generates the config, compares it with the running config, refuses known bind
conflicts, attempts hot apply, and otherwise stops the old process before
starting a new one. If the new process does not start, the previous known-good
runtime is not automatically restored.

Hot application uses `internal/xray/hot_diff.go` and Xray API methods to remove
or add users, inbounds, and outbounds and to apply routing. A partial hot-apply
failure falls back to full restart. `internal/xray/api.go` also implements
traffic, online-user, route-test, routing-apply, and balancer API calls.

Any future apply coordinator must wrap, not replace, this path:

```text
snapshot current known-good config/state
-> build candidate through GetXrayConfig
-> validate candidate with the installed Xray binary
-> apply through existing hot/restart behavior
-> healthcheck
-> commit known-good snapshot
```

On failure it must restore the snapshot, restart/reload as necessary, run a
healthcheck, and retain an audit record. This is a Phase 0 safety prerequisite
for fork features that alter routing or Xray configuration.

## Routing

The editable routing source is `xrayTemplateConfig.routing`. The frontend
schema and editor live under `frontend/src/schemas/routing.ts` and
`frontend/src/pages/xray/routing/`. Rules already support `user` (client email),
`inboundTag`, `outboundTag`/balancer, source/destination fields, protocol, and
the existing `vlessRoute` behavior. `RuleFormModal.tsx` can select client
emails.

`internal/xray/api.go` maps `RouteTestRequest.Email` to the Xray routing
context's user field. The current `RouteTester.tsx` does not expose/send that
email, so backend route testing is already more capable than the UI.

`internal/web/service/xray_setting_routing_sync.go` maintains routing references
when inbound tags change. `ensureStatsRouting()` preserves the required
API-to-API rule, and disabled rules are stripped during compilation. Policy
decoration must respect all three behaviors.

Subscription routing is a separate concern in `internal/sub`, including
standard and JSON routing settings and remote routing cache behavior. A policy
feature that affects generated client subscriptions must explicitly cover this
pipeline; changing only the runtime Xray template would otherwise create a
panel/runtime mismatch.

`model.Host.VlessRoute` encodes an existing subscription-specific VLESS route
choice into generated links. It is not a general client policy assignment
system and should not be repurposed as one.

## Traffic / statistics

The live traffic path is:

```text
Xray StatsService API
-> XrayTrafficJob (5 seconds)
-> InboundService.AddTraffic / outbound aggregation
-> traffic writer and database models
-> WebSocket/UI and optional external reporting
```

`internal/xray/api.go` defines `GetTraffic()` and traffic structures.
`internal/web/job/xray_traffic_job.go` samples traffic; the persistence path is
split through `internal/web/service/inbound_traffic.go` and
`traffic_writer.go`. Cumulative traffic already exists for clients, inbounds,
outbounds, nodes, and global shared-client views.

Remote-node synchronization is handled by `node_traffic_sync_job.go` and the
node/global traffic models. System, Xray, and observatory histories also exist
through metric-history services. A fork analytics feature should consume the
existing normalized traffic and node dimensions before adding a second counter
system. Fine-grained domain/destination analytics is a distinct data stream and
must have explicit retention, aggregation, and privacy limits.

## Online state / IP tracking

`XrayService.GetOnlineUsers()` uses the Xray StatsService online API and caches
whether that capability is available. `XrayTrafficJob` combines online API
results with traffic deltas and updates process/inbound activity state. Remote
node online trees are merged by the existing Xray process/runtime structures.

`internal/web/job/check_client_ip_job.go` runs approximately every ten seconds.
It deliberately uses the online API, persists observed identities/IPs in
`InboundClientIps` and `NodeClientIp`, applies the existing `LimitIP` logic and
fail2ban integration where configured, and removes stale state.

Therefore:

- `access.log` is not the current real-time online source;
- an IP observation is not itself proof of sharing;
- policy/security features should treat IP, ASN, GeoIP, timing, and node data as
  weighted evidence rather than an automatic-ban trigger; and
- optional API absence must degrade safely instead of silently switching to a
  weaker source of truth.

## access.log

`internal/xray/process.go`, `GetAccessLogPath()`, reads the generated log path.
`resolveXrayLogPaths()` confines enabled paths to the panel log directory.

`ServerService.GetXrayLogs()` in `internal/web/service/server.go` opens and scans
the whole access log on demand, parses/filter records, and retains the requested
tail. The parser is defensive against malformed records, but this path is a log
viewer, not an incremental collector.

`internal/web/job/clear_logs_job.go` prunes oversized logs and clears access and
error logs daily. There is currently no persisted cursor, inode/file-identity
tracking, rotation/truncation state machine, batching pipeline, or analytics
storage built around `access.log`.

A later analytics collector must be an independently disableable service with
graceful shutdown and explicit handling for create, append, rename/rotation,
copy-truncate, deletion, and malformed/partial records. It must coordinate with
the existing log cleanup job. It must not collect HTTP bodies, cookies,
Authorization values, credentials, decrypted messages, or passwords, and it
must not become the only online-state source.

## Jobs / schedulers

`internal/web/web.go`, `startTask()`, is the current scheduler composition
point. Jobs use robfig/cron with `SkipIfStillRunning` and panic recovery. Current
cadences include Xray liveness, pending restart, five-second traffic sampling,
ten-second sidecar/IP work, node heartbeat/traffic synchronization, outbound
subscription refresh, orphan cleanup, remote routing refresh, log cleanup,
WARP update, traffic resets, directory/notification integrations, and resource
monitoring.

The fork needs one `RegisterJobs` hook invoked after the scheduler exists and
before it starts accepting normal work. Fork jobs must accept server context,
avoid duplicate registration across restart paths, use overlap protection, and
finish cleanly during shutdown. High-volume ingestion should use a dedicated
bounded worker/batcher rather than a cron callback or the notification bus.

## Event system

`internal/eventbus` supplies an in-process, bounded, non-durable event bus.
Existing events include Xray crash, node up/down, outbound up/down, high CPU or
memory, and login attempts. `web.Server` creates and starts the bus, and
registers email, Discord, and optional Telegram subscribers. Publishers include
the Xray crash callback and relevant monitoring jobs.

The bus uses a bounded central queue and bounded subscriber queues, nonblocking
publish/drop behavior, worker isolation, and panic recovery. This is appropriate
for low-volume notification and audit-trigger signals where occasional loss is
acceptable. It is not suitable for exact-once audit history or high-volume
traffic/DNS/access analytics. Durable audit records must be written through a
database-backed service, with the bus optionally notifying after persistence.

One subscriber-registration hook in `web.Server.start()` is sufficient. Fork
code should use existing event types when their semantics match and add new
low-rate types in a fork-owned package rather than expanding upstream event
files for every feature.

## Backend routes / API

`internal/web/web.go`, `initRouter()`, installs Gin middleware, the configured
base path, controllers, WebSockets, static files, and SPA handling.

`internal/web/controller/api.go`, `APIController.initRouter()`, creates the
`/panel/api` group and applies the authentication, scope/config response, and
CSRF middleware before registering inbound, client/group, server, node, host,
settings, Xray, and balancer routes. `internal/web/controller/spa.go` serves the
panel SPA and provides an authenticated `NoRoute` fallback.

The safest backend route integration is one call at the end of
`APIController.initRouter()` that receives the already protected `*gin.RouterGroup`
and only registers a fixed `/fork/...` subtree. This preserves upstream
middleware and avoids editing each controller. Public endpoints, if ever
required, must use a distinct explicit hook and security review rather than
being accidentally mounted outside authentication.

Fork handlers should retain the existing response envelope, authorization
expectations, request validation, CSRF treatment, and base-path behavior.

## OpenAPI / generated code

`tools/openapigen/main.go` is an AST-based generator with explicit package/type
allowlists. It writes files below `frontend/src/generated/`; those files state
that they must not be edited manually.

`frontend/scripts/build-openapi.mjs` combines generated schemas/examples with a
manually maintained endpoint catalog in
`frontend/src/pages/api-docs/endpoints.ts`, then writes
`frontend/public/openapi.json`. The backend embeds and serves that artifact via
`internal/web/controller/dist.go` at `/panel/api/openapi.json`, rewriting base
path details as required.

Thus OpenAPI generation is not inferred entirely from backend route
annotations. A new fork API requires:

1. source request/response types in a generator-visible fork package;
2. a minimal extension to the generator input/allowlist;
3. endpoint catalog entries;
4. execution of the existing generation commands; and
5. `gen-check` verification.

Generated TypeScript and the public OpenAPI JSON must never be hand-edited.

## Frontend routes / navigation

`frontend/src/main.tsx` creates the router, while
`frontend/src/routes.tsx` contains lazy route objects rendered through
`PanelLayout`. Existing routes cover dashboard, inbounds, clients, groups,
nodes, hosts, settings, Xray configuration/routing, and API documentation.

`frontend/src/layouts/AppSidebar.tsx` contains hard-coded primary and submenu
navigation arrays. `frontend/src/components/CommandPalette.tsx` has another
hard-coded set of navigation/settings entries. Because the backend has an
authenticated SPA fallback, most new client-side paths do not need new explicit
Gin page routes.

The low-divergence frontend design is a fork-owned route descriptor file and a
fork-owned navigation descriptor file, each spread once into the upstream
arrays. Prefer sharing the navigation descriptor with the command palette so a
feature is not declared three times. Each descriptor should be omitted when its
feature flag is disabled. Existing client UI is under
`frontend/src/pages/clients/` (including `ClientsPage.tsx`) and uses hooks such
as `frontend/src/hooks/useClients.ts`; group UI is under
`frontend/src/pages/groups/` (including `GroupsPage.tsx`). These pages and their
existing service hooks should be extended rather than cloned.

## Updater

The current update chain is tied to the official upstream repository:

- `internal/web/service/panel/panel.go` defines `panelUpdaterURL` for the
  upstream `update.sh`; `fetchPanelRelease()` queries
  `api.github.com/repos/MHSanaei/3x-ui/releases`.
- `update.sh` queries/downloads official release artifacts, scripts, and
  service units.
- `x-ui.sh` has official sources in `install()`, `update()`, `update_dev()`, and
  `installed_script_url()`.
- `install.sh` downloads official installers, release artifacts, scripts, and
  service units.

Consequently, a normal update can overwrite the fork with an upstream build.
This is a release blocker, not a cosmetic branding issue.

The updater must use a single fork release identity (owner, repository, API,
raw-content base, asset naming, checksum naming, and channel policy) consumed by
the Go updater and shell scripts. It should fail closed when release metadata or
asset identity does not match the fork. `.github/workflows/release.yml` must
publish the exact names expected by the clients. The internal name/version
files and user-facing repository link must be consistent with that identity.

The current update script performs integrity/status checks, but it can remove
the installed version before a replacement is proven healthy. Fork release
work therefore also needs snapshot, staged extraction/validation, atomic
activation where possible, service healthcheck, and automatic restoration of
the previous version. Updater isolation belongs before feature development in
Phase 0.

## Backup / restore

`ServerService.GetDb()` uses the SQLite online backup API or PostgreSQL
`pg_dump`. `GetMigration()` produces a portable SQLite-oriented migration
artifact. `ImportDB()` detects source type, stages and validates data, stops
Xray, optionally preserves host-bound settings, keeps a SQLite fallback,
activates the candidate database, reinitializes/migrates it, and restarts Xray.
PostgreSQL restore uses `pg_restore --single-transaction` and a readability
probe.

Database activation failures have restoration paths, but a successful database
activation followed by Xray restart failure does not automatically roll back
the imported database/config to the previous known-good combination. The
fallback remains available for manual recovery. This gap matters for future
fork tables and Xray-affecting settings.

Fork tables must participate in export/import, engine conversion, validation,
and retention review. A Phase 0 recovery test should cover the failure after DB
activation but before a healthy Xray result, for both database engines where
applicable.

## Test infrastructure

The `Makefile` supplies generation, lint, format, typecheck, Go/frontend test,
build, Storybook, and verification targets. `make verify` is comprehensive.
There is currently no `make verify-fork` target even though the fork governance
requires one; adding that target is Phase 0 work, not part of this mapping task.

CI includes:

- shuffled Go tests and selected PostgreSQL tests;
- code-generation drift checks;
- vulnerability and race jobs;
- fuzz smoke tests for selected parsers;
- Go/frontend linting and frontend type/test/build/Storybook checks;
- package audit, CodeQL, install smoke tests, and release smoke tests; and
- informational mutation testing.

Tests are generally colocated with source. JSON golden fixtures live below
`frontend/src/test/golden/fixtures` and are also checked against vendored Xray
builders by backend tests. Existing service tests cover substantial Xray
configuration and hot-diff behavior. Access-log parsing is tested, but no
incremental tail/rotation collector exists to test.

The required all-fork-features-disabled proof should include:

- exact ordered/semantic equality of generated Xray configuration before and
  after the no-op decorator;
- no fork routes, jobs, event subscribers, or background workers registered;
- unchanged client CRUD, inbound attachment, group behavior, routing template,
  standard/JSON/Clash/Happ subscription output, and local startup smoke tests;
- fresh and supported-upgrade database tests on SQLite and PostgreSQL;
- validation that fork migrations do not alter upstream tables/data; and
- release/update smoke tests that reject official-upstream assets.

## Proposed fork integration points

The proposed design is a fixed first-party hook set, not a runtime plugin
ecosystem. `main.go` explicitly installs fork implementations. Neutral hook
contracts may depend on Gin, GORM, cron, Xray config types, eventbus types, and
`context.Context`, but must avoid importing the web service package to prevent
cycles. Empty/default hooks are exact no-ops.

### Fork bootstrap

- File: `main.go` and new `internal/forkext` / fork module packages
- Function/type: `runWebServer()` and a new fixed `forkext.HookSet` installation
- Purpose: Construct first-party fork modules once, before database
  initialization and server startup.
- Fork hook: `forkext.Install(forkmodules.Build(...))`, with immutable/default
  no-op hooks.
- Why this is minimal: One composition call keeps module imports out of
  upstream services and makes every customization discoverable/removable.
- Expected upstream-sync conflict risk: LOW

### Database models and migrations

- File: `internal/database/db.go`
- Function/type: `allModels()`, `initModels()`, and `runSeeders()`
- Purpose: Register fork-owned models and ordered, idempotent
  SQLite/PostgreSQL migrations before application services use them.
- Fork hook: One `forkext.Database` hook returning models and executing
  explicit migrations/seeders through a narrow DB/dialect context.
- Why this is minimal: It replaces repeated edits to three upstream lists with
  a single stable call and keeps schema logic in fork files.
- Expected upstream-sync conflict risk: MEDIUM

### Protected backend routes

- File: `internal/web/controller/api.go`
- Function/type: `APIController.initRouter()`
- Purpose: Mount fork APIs under the existing authentication, scope/config,
  and CSRF middleware.
- Fork hook: `forkext.RegisterProtectedRoutes(apiGroup, dependencies)` at the
  end of current route registration, using a fixed `/fork` subtree.
- Why this is minimal: One call avoids scattering fork handlers across
  upstream controllers and preserves upstream security behavior.
- Expected upstream-sync conflict risk: LOW

### Scheduled jobs

- File: `internal/web/web.go`
- Function/type: `Server.startTask()`
- Purpose: Register retention, aggregation, refresh, or health jobs owned by
  fork modules.
- Fork hook: `forkext.RegisterJobs(serverContext, cronScheduler, dependencies)`
  after upstream registrations and before scheduler execution.
- Why this is minimal: It reuses the existing scheduler, overlap policy,
  context, and shutdown rather than creating another scheduler.
- Expected upstream-sync conflict risk: MEDIUM

### Event subscribers

- File: `internal/web/web.go`
- Function/type: `Server.start()` around event bus subscriber registration
- Purpose: Attach low-volume fork notifications after the existing bus is
  created and before normal publishers run.
- Fork hook: `forkext.RegisterEventSubscribers(bus, dependencies)`.
- Why this is minimal: It uses the established bus lifecycle without modifying
  each publisher or upstream event subscriber.
- Expected upstream-sync conflict risk: LOW

### Long-running fork services

- File: `internal/web/web.go`
- Function/type: `Server.start()` and `Server.stop()`/`Stop()`
- Purpose: Start and gracefully stop bounded collectors or apply coordinators
  that are not cron jobs.
- Fork hook: `forkext.Start(ctx, dependencies)` returning closers, followed by
  `forkext.Stop(ctx)` before dependent upstream facilities stop.
- Why this is minimal: Two lifecycle calls prevent orphan goroutines and keep
  collector internals out of the upstream server type.
- Expected upstream-sync conflict risk: MEDIUM

### Final Xray configuration decoration

- File: `internal/web/service/xray.go`
- Function/type: `XrayService.GetXrayConfig()`
- Purpose: Apply enabled fork policies to the fully assembled candidate config.
- Fork hook: `forkext.DecorateXrayConfig(ctx, config, readOnlyContext)`
  immediately before the successful return.
- Why this is minimal: A single deterministic decorator preserves the
  canonical compiler and sees all upstream inbounds/outbounds/routing additions.
- Expected upstream-sync conflict risk: HIGH

### Xray apply safety wrapper

- File: `internal/web/service/xray.go` and a new fork-owned apply coordinator
- Function/type: `XrayService.RestartXray()`
- Purpose: Snapshot, validate, apply, healthcheck, and restore known-good state
  for fork-caused Xray changes.
- Fork hook: A narrowly scoped pre-apply/post-apply coordinator invoked around
  the existing hot-apply/full-restart path; disabled mode delegates unchanged.
- Why this is minimal: It retains upstream process and hot-diff behavior while
  adding the mandatory recovery boundary in one place.
- Expected upstream-sync conflict risk: HIGH

### OpenAPI type discovery

- File: `tools/openapigen/main.go`
- Function/type: generator package/type allowlists
- Purpose: Include fork request/response types in generated Zod/TypeScript
  output.
- Fork hook: One fork-package entry in the generator inputs, with fork types
  maintained outside upstream controller/service packages.
- Why this is minimal: It keeps the existing generator authoritative and
  avoids hand-edited generated code.
- Expected upstream-sync conflict risk: LOW

### OpenAPI endpoint catalog

- File: `frontend/src/pages/api-docs/endpoints.ts`
- Function/type: endpoint descriptor collection
- Purpose: Document fork endpoints in the generated OpenAPI artifact.
- Fork hook: Import and spread a fork-owned endpoint descriptor array once.
- Why this is minimal: It follows the current partially manual pipeline
  without mixing every fork endpoint into the upstream catalog.
- Expected upstream-sync conflict risk: LOW

### Frontend routes

- File: `frontend/src/routes.tsx`
- Function/type: root `RouteObject` children
- Purpose: Add lazy fork pages without duplicating the upstream router.
- Fork hook: Import and spread `forkRoutes`, whose factory filters disabled
  features.
- Why this is minimal: One insertion keeps all pages in fork-owned feature
  directories and preserves the SPA fallback.
- Expected upstream-sync conflict risk: LOW

### Frontend navigation and command palette

- File: `frontend/src/layouts/AppSidebar.tsx` and
  `frontend/src/components/CommandPalette.tsx`
- Function/type: navigation/submenu and command item arrays
- Purpose: Expose enabled fork pages consistently.
- Fork hook: Import one shared fork navigation descriptor and adapt/spread it
  into both consumers.
- Why this is minimal: Two small adapters avoid duplicating each feature's
  labels, paths, permissions, and enabled state in upstream files.
- Expected upstream-sync conflict risk: MEDIUM

### Updater release identity

- File: `internal/web/service/panel/panel.go`, `update.sh`, `x-ui.sh`, and
  `install.sh`
- Function/type: panel release lookup/download and shell install/update
  functions
- Purpose: Ensure every installation and update path resolves only fork-owned,
  verified release artifacts.
- Fork hook: A single fork release identity/constants source mirrored into
  shell defaults during packaging, plus owner/product validation.
- Why this is minimal: These are unavoidable distribution identity points; a
  shared source prevents unrelated feature logic from entering the updater.
- Expected upstream-sync conflict risk: HIGH

### Fork verification target

- File: `Makefile` and CI workflow invocation
- Function/type: new `verify-fork` target
- Purpose: Provide the governance-required compatibility, migration, rollback,
  privacy, and feature-disabled suite.
- Fork hook: One additive target composed from fork-owned test packages/scripts;
  CI invokes it after upstream verification.
- Why this is minimal: Upstream targets remain unchanged and reusable.
- Expected upstream-sync conflict risk: LOW

## Permanent upstream-touch hotspots

The following files/functions are expected to retain one small downstream hook
or fork identity edit across upstream synchronizations:

| Hotspot | Permanent reason | Discipline |
| --- | --- | --- |
| `main.go` / `runWebServer()` | Install fixed fork modules before DB/server startup | One call; no feature logic |
| `internal/database/db.go` | Model/migration discovery has no external registry | One database hook; fork migrations elsewhere |
| `internal/web/controller/api.go` / `initRouter()` | Protected API group is local to this function | One route-registration call |
| `internal/web/web.go` / startup, jobs, shutdown | Scheduler, event bus, and lifecycle are composed here | Fixed job/subscriber/start/stop calls only |
| `internal/web/service/xray.go` / `GetXrayConfig()` | Only safe point after complete upstream config assembly | One no-op-capable decorator |
| `internal/web/service/xray.go` / `RestartXray()` | Existing hot/full apply boundary needs transactional safety | Narrow wrapper; retain upstream implementation |
| `tools/openapigen/main.go` | Explicit source allowlist | One fork source registration |
| `frontend/src/routes.tsx` | Root frontend route list | One spread |
| `frontend/src/layouts/AppSidebar.tsx` | Root navigation list | One spread/adapter |
| `frontend/src/components/CommandPalette.tsx` | Separate discoverability list | Shared descriptor adapter |
| `frontend/src/pages/api-docs/endpoints.ts` | Manual endpoint catalog | One fork catalog spread |
| Panel and shell updater files | Distribution identity cannot remain upstream | Centralized constants only |
| `Makefile` and CI | Mandatory fork verification entry point | Additive target/job |

If feature logic begins accumulating in these files, the hook boundary has
failed and should be redesigned before continuing.

## Spec assumptions that were correct

- The repository is a Go/Gin/GORM backend with a React/TypeScript frontend and
  an external Xray runtime.
- SQLite and PostgreSQL are both supported and require dialect-aware migration
  testing.
- `XrayService.GetXrayConfig()` is the correct central config-generation
  boundary.
- Xray exposes traffic, online-user, routing, handler, and balancer operations
  through its API, and optional capability detection is necessary.
- The existing event bus can support low-volume notification integrations.
- `access.log` can supply allowed connection metadata for optional analytics,
  subject to privacy, retention, and rotation handling.
- Updater isolation, rollback, feature flags, and an all-disabled compatibility
  proof belong in Phase 0.
- The checked-out embedded version is `3.8.5`.

## Spec assumptions that need correction

- Clients are no longer only embedded JSON records. They have normalized
  `ClientRecord` and attachment tables while retaining JSON synchronization for
  Xray compatibility.
- Client groups already exist as `ClientGroup` and are managed through
  `ClientService`; a new generic group subsystem is unnecessary.
- Multi-node management, remote runtimes, node traffic, node IP observations,
  and global traffic already exist.
- Real-time online state already uses the Xray online API combined with traffic
  deltas. There is no current `access.log` fallback for the IP-limit job.
- Current `access.log` use is an on-demand full scan plus cleanup; there is no
  incremental analytics collector.
- The event bus is intentionally lossy/bounded and cannot be the durable audit
  or high-volume analytics transport.
- Database migrations are embedded in `db.go` helpers/seeders rather than a
  standalone ordered migration directory.
- Xray settings validation is not a complete isolated candidate validation, and
  failed restart does not automatically restore a known-good runtime.
- OpenAPI generation is partly AST-generated and partly a manual endpoint
  catalog; backend route declaration alone does not update it.
- The frontend stack in the checked-out repository is React 19, Ant Design 6,
  and Vite 8.
- The updater's official-upstream targeting is confirmed and can replace the
  fork; changing branding alone will not prevent it.

## Features already present upstream

- Normalized clients and inbound attachment.
- Client groups, memberships, traffic reset, HWID state, and external links.
- Local and remote node management with runtime abstraction.
- Per-client, per-inbound, outbound, node, and global traffic accounting.
- Online-user API consumption and client-IP/limit tracking.
- Client-specific routing rule fields and backend route-test user context.
- Outbound subscriptions and remote subscription routing support.
- Xray hot-diff application for users, inbounds, outbounds, and routing.
- System/Xray/observatory metrics histories.
- Event-driven email, Telegram, and Discord notifications.
- SQLite online backup, PostgreSQL dump/restore, import staging, and cross-engine
  migration helpers.
- Generated schemas/types/examples and an embedded OpenAPI document.
- Log path confinement, log viewer parsing, and log cleanup.
- Release update channels, checksum verification, and install smoke tests.

## Duplicate functionality we should NOT implement

- A second client, group, inbound, or node database/domain model.
- A second traffic counter or remote-node synchronization system.
- An `access.log`-based replacement for Xray online state.
- A parallel Xray configuration compiler or a forked xray-core.
- A second cron scheduler, event bus, WebSocket hub, router, or frontend shell.
- A generic plugin marketplace/runtime for fixed first-party fork features.
- A hand-maintained copy of generated TypeScript/OpenAPI schemas.
- A second subscription routing engine disconnected from `internal/sub`.
- A second updater layered on top of the official-targeting updater; the
  existing paths must be safely retargeted and hardened.
- Automatic bans based on one IP, ASN, GeoIP, SNI, or behavioral heuristic.
- Storage of HTTPS bodies, cookies, Authorization headers, credentials,
  decrypted messages, or passwords.

## Recommended Phase 0 implementation sequence

1. Establish the all-features-disabled baseline fixtures and tests for config,
   subscriptions, core CRUD/routing, startup, and both database engines.
2. Add the fixed neutral fork hook contracts and explicit bootstrap, with every
   hook initially empty. Prove the baseline remains identical.
3. Add `make verify-fork` and CI execution, including race tests for fork
   concurrency and engine-specific migration tests.
4. Isolate fork release identity across Go, shell scripts, packaging, and
   release workflows. Add tests that reject upstream-owner assets.
5. Add transactional updater staging, healthcheck, and automatic rollback.
6. Introduce the fork feature-flag/settings facade with defaults disabled and
   explicit SQLite/PostgreSQL migration coverage.
7. Add Xray candidate validation and known-good snapshot/restore around the
   existing apply path, still with no policy decorator behavior enabled.
8. Close the database-import/Xray-start recovery gap and exercise restore
   failure paths.
9. Add fork route/OpenAPI/frontend descriptor registries as empty/no-op
   integrations and re-run the compatibility suite.
10. Only after this foundation passes `make verify` and `make verify-fork`,
    start the first separately scoped fork feature. Do not automatically move
    to that work as part of TASK-001.

## Risks / blockers

- **Updater overwrite — critical:** every active update/install path still
  points at `MHSanaei/3x-ui`. Fork deployments can be replaced by upstream.
- **No `verify-fork` target — critical process blocker:** the mandatory fork
  verification command does not yet exist.
- **No automatic Xray rollback — high:** full restart stops the previous
  process before the replacement is known healthy.
- **Import recovery gap — high:** an imported DB can remain active when its
  resulting Xray configuration fails to start.
- **Migration mechanism concentration — high:** schema compatibility logic is
  concentrated in `internal/database/db.go`; unmanaged fork additions would
  create frequent conflicts and engine drift.
- **Config decorator sensitivity — high:** routing/outbound order, API routing,
  normalized clients, subscriptions, and hot-diff behavior all meet at the
  final config. Even one hook requires strong golden and smoke coverage.
- **Feature-flag storage choice — medium:** adding every flag to upstream's
  reflected `AllSetting`/default maps would spread changes. A fork-owned facade
  or namespaced persisted representation should be chosen in its own task.
- **Access-log lifecycle — medium:** daily clearing and size pruning can race an
  incremental collector unless file identity and truncate semantics are
  explicit.
- **Event loss — medium:** the existing event bus drops under pressure and is
  not an audit ledger.
- **Frontend registry duplication — medium:** routes, sidebar, command palette,
  endpoint catalog, permissions, and flags can drift unless fork descriptors
  are shared.
- **PostgreSQL coverage depth — medium:** CI has targeted PostgreSQL cases, not
  exhaustive parity for future fork migrations.
- **Optional Xray capabilities — medium:** online/routing API differences must
  degrade safely and be observable without changing upstream behavior.
- **Privacy/retention — medium:** destination/DNS/SNI analytics requires strict
  allowlists, bounded retention, deletion behavior, and tests before collection
  is enabled.

No production source files were changed during this mapping pass.
