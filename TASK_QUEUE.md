# TASK QUEUE

The coding agent must execute only the task named in `CURRENT_TASK.md`.

Never automatically advance to the next task.

This queue supersedes the original planning queue and reflects the verified repository map in `docs/19_REPOSITORY_MAP.md`.

---

# PHASE 0 — Fork Safety Foundation

The goal of Phase 0 is to make CatX-UI safe to extend before adding product features.

## TASK-001 — Repository Mapping & Integration Audit

Recommended model: `Sol High`  
Status: `DONE`  
Output: `docs/19_REPOSITORY_MAP.md`  
Production code changes: `NONE`

Result:

- verified actual 3x-ui architecture;
- identified permanent upstream-touch hotspots;
- confirmed existing upstream features that must be reused;
- identified updater overwrite and rollback gaps;
- established the corrected Phase 0 sequence.

---

## TASK-002 — Baseline Compatibility Suite

Recommended model: `Luna Medium`  
Status: `NEXT`

Goal:

Create a baseline test/fixture suite that proves current CatX-UI behavior matches upstream `v3.8.5` before fork hooks are added.

The suite must establish reference behavior for:

- Xray config generation;
- standard subscription output;
- JSON subscription output;
- Clash/Happ output where existing tests support it;
- client CRUD;
- client-group behavior;
- inbound/client attachment behavior;
- routing behavior;
- local startup smoke;
- SQLite baseline;
- PostgreSQL baseline where supported by existing test infrastructure.

Important:

- Do not add fork hooks yet.
- Do not add feature flags yet.
- Do not change product behavior.
- Reuse existing upstream test infrastructure and golden fixtures where possible.
- Avoid duplicating tests that already prove the same behavior.
- The output of this task becomes the regression baseline for every later Phase 0 task.

Expected deliverables:

- baseline compatibility tests/fixtures;
- a short `docs/20_BASELINE_COMPATIBILITY.md` describing what is covered;
- no intentional runtime behavior change.

---

## TASK-003 — Fixed Fork Hook Contracts & Bootstrap

Recommended model: `Luna Medium`  
Review model for architecture concerns: `Sol Medium` only if needed.

Goal:

Add the smallest fixed first-party extension boundary based on `docs/19_REPOSITORY_MAP.md`.

Initial hook capabilities should exist only where required by verified integration points.

Candidate hook areas:

```text
bootstrap/install
database models/migrations
protected routes
jobs
event subscribers
long-running services
Xray config decoration
frontend/OpenAPI registration
```

Rules:

- Do not build a generic plugin framework.
- Hook defaults must be exact no-ops.
- Feature behavior must remain unchanged.
- Baseline compatibility suite must remain green.

---

## TASK-004 — `make verify-fork` + CI Gate

Recommended model: `Luna Medium`

Goal:

Create a fork verification target without modifying upstream verification semantics.

`make verify-fork` should compose:

- `make verify`;
- fork compatibility tests;
- fork migration tests;
- race tests for fork concurrent packages;
- parser fuzz tests when those packages exist;
- fork-specific smoke checks.

Add CI execution for the fork target.

Rules:

- Do not replace upstream CI targets.
- Keep upstream verification reusable and intact.

---

## TASK-005 — CatX-UI Identity & Updater Isolation

Recommended model: `Sol High`

Goal:

Ensure CatX-UI can never be unintentionally replaced by an official `MHSanaei/3x-ui` release.

Verified upstream-sensitive paths include:

```text
internal/web/service/panel/panel.go
update.sh
x-ui.sh
install.sh
.github/workflows/release.yml
```

Deliver:

- CatX-UI release identity;
- separate fork version / upstream base / Xray version;
- fork-owned release repository;
- stable/dev channels;
- consistent asset naming;
- checksum/integrity verification;
- tests that reject official-upstream assets.

This task isolates release identity.

Do not yet implement the full transactional updater rollback flow.

---

## TASK-006 — Transactional Updater Staging & Automatic Rollback

Recommended model: `Sol High`

Goal:

Harden the updater so a failed update restores the previous known-good CatX-UI installation.

Required flow:

```text
download
→ verify
→ stage
→ backup
→ install candidate
→ migrate
→ start
→ healthcheck
→ commit
```

Failure flow:

```text
restore previous binary/config/state
→ start
→ healthcheck
→ audit/report
```

Required:

- no destructive replacement before the candidate is staged;
- explicit DB migration recovery strategy;
- release/update smoke tests;
- rollback tests.

---

## TASK-007 — Fork Settings Facade & Feature Flags

Recommended model: `Luna Medium`

Goal:

Add a fork-owned settings layer without spreading every fork flag through upstream `AllSetting` and default maps.

Initial flags:

```text
analytics.enabled
dns_intelligence.enabled
policies.enabled
traffic_control.enabled
security_anomaly.enabled
```

Rules:

- all major fork features default OFF;
- all OFF must preserve baseline behavior;
- SQLite and PostgreSQL persistence must be tested;
- do not overload unrelated upstream settings fields.

---

## TASK-008 — Xray Candidate Validation & Known-Good Rollback

Recommended model: `Sol High`

Goal:

Wrap the existing Xray apply/restart path with safety while preserving upstream hot-diff/runtime behavior.

Required flow:

```text
snapshot known-good state
→ build candidate through existing GetXrayConfig()
→ validate candidate with installed Xray
→ apply through existing hot/restart path
→ healthcheck
→ commit known-good state
```

Failure:

```text
restore snapshot
→ restart/reload
→ healthcheck
→ audit/report
```

Rules:

- do not replace `GetXrayConfig()`;
- do not replace existing Xray process/runtime management;
- disabled fork features must remain exact no-ops.

---

## TASK-009 — DB Import / Xray Start Recovery Gap

Recommended model: `Sol Medium`

Goal:

Close the verified recovery gap where a database import can activate successfully but the resulting Xray configuration fails to start.

Required:

- SQLite recovery path;
- PostgreSQL recovery path where applicable;
- restore previous DB/config combination;
- healthcheck;
- failure-path tests.

---

## TASK-010 — Empty Fork API / OpenAPI / Frontend Registries

Recommended model: `Luna Medium`

Goal:

Add minimal empty/no-op registration points for fork UI/API surfaces.

Verified integration areas:

```text
internal/web/controller/api.go
tools/openapigen/main.go
frontend/src/pages/api-docs/endpoints.ts
frontend/src/routes.tsx
frontend/src/layouts/AppSidebar.tsx
frontend/src/components/CommandPalette.tsx
```

Deliver:

- protected fork API subtree registration;
- fork OpenAPI type/endpoint descriptors;
- fork route descriptor;
- shared fork navigation descriptor.

No product feature pages yet.

Baseline compatibility must remain green.

---

# PHASE 1 — Analytics Foundation

Phase 1 must reuse upstream client/group/node/traffic models instead of creating parallel sources of truth.

## TASK-101 — Analytics Persistence Schema

Recommended model: `Luna Medium`

Goal:

Add fork-owned analytics persistence only for data that upstream does not already store.

Examples:

- destination events;
- DNS observations;
- correlated sessions;
- service/category aggregates.

Rules:

- reference existing client/group/node IDs;
- do not duplicate upstream traffic counters;
- SQLite + PostgreSQL migration coverage;
- explicit retention design.

---

## TASK-102 — Incremental `access.log` Collector

Recommended model: `Sol High`

Must support:

- persistent offset/checkpoint;
- file identity/inode handling where applicable;
- append;
- rename/rotation;
- copy-truncate;
- deletion/recreation;
- partial lines;
- malformed input;
- batching;
- bounded backpressure;
- graceful shutdown;
- coexistence with upstream log cleanup;
- fuzz tests;
- race tests.

Must not become the real-time online-state source.

---

## TASK-103 — Destination Event Normalizer

Recommended model: `Luna Medium`

Goal:

Normalize allowed metadata from:

```text
access.log
Xray destination metadata
future DNS observer
future SNI/TLS metadata
```

Store source/provenance and confidence.

No decrypted HTTPS content.

---

## TASK-104 — Xray Online/Traffic Adapter

Recommended model: `Luna Medium`

Goal:

Reuse existing upstream online-user and traffic APIs as inputs to fork analytics.

Rules:

- do not build a second online-state system;
- do not build a second traffic-counter system;
- degrade safely when optional Xray APIs are unavailable.

---

## TASK-105 — Session Correlator

Recommended model: `Sol Medium`

Goal:

Correlate multiple observations into logical network sessions.

Inputs may include:

```text
client
domain/destination
port
network
time window
source evidence
```

Do not present estimated network activity as exact screen time.

---

## TASK-106 — Historical Analytics Aggregator

Recommended model: `Luna Medium`

Goal:

Build historical analytics by combining:

- existing upstream traffic dimensions;
- fork destination/session data;
- daily/hourly aggregates.

Do not duplicate existing per-client/per-inbound/per-node counters.

---

## TASK-107 — Basic Client Activity API/UI

Recommended model: `Luna Medium`

Show:

- recent service/destination activity;
- session count;
- traffic;
- first/last seen;
- allow/block action when policy exists later;
- confidence/source where relevant.

---

# PHASE 2 — DNS / Destination Intelligence

## TASK-201 — DNS Observer

Recommended model: `Sol Medium`

Goal:

Collect attributable DNS metadata without blocking core Xray operation.

---

## TASK-202 — Destination Evidence Fusion

Recommended model: `Sol High`

Fuse evidence from:

```text
DNS
Xray requested destination
SNI where visible
destination IP
access.log
TLS/QUIC metadata
```

Maintain source and confidence.

---

## TASK-203 — ASN / GeoIP Enrichment

Recommended model: `Luna Medium`

Enrichment failure must not block traffic.

---

## TASK-204 — Service / Category Classifier

Recommended model: `Luna Medium`

Architecture:

```text
evidence
→ provider
→ logical service
→ logical category
→ confidence
```

Do not permanently bind policy semantics to one provider/geosite format.

---

## TASK-205 — DNS Intelligence Dashboard

Recommended model: `Luna Medium`

Show:

- queries;
- blocked queries;
- unique domains;
- new domains;
- NXDOMAIN;
- clients;
- services/categories.

---

## TASK-206 — Service Relationship Graph & First-Seen Intelligence

Recommended model: `Luna Medium`

Group infrastructure domains under logical services and expose first/last-seen metadata.

---

## TASK-207 — Analytics Retention & Privacy Controls

Recommended model: `Luna Medium`

Deliver:

- configurable retention;
- delete-history action;
- privacy dashboard;
- explicit statement that HTTPS bodies/cookies/passwords/tokens are not collected.

---

# PHASE 3 — Policy Engine v1

Policy must reuse existing normalized clients, client groups, inbounds, nodes, and Xray routing.

## TASK-301 — Policy Schema / Repository

Recommended model: `Luna Medium`

Reference existing stable client/group/inbound/node IDs.

Do not create parallel client/group models.

---

## TASK-302 — Policy CRUD API

Recommended model: `Luna Medium`

Use the protected fork API registry.

---

## TASK-303 — Group Assignment / Client Override

Recommended model: `Luna Medium`

Support:

```text
inherited
override
return to policy
```

Reuse existing `ClientGroup`.

---

## TASK-304 — Policy Precedence Engine

Recommended model: `Sol High`

Define and test exact precedence for:

- system/emergency;
- temporary allow;
- client allow/block;
- group allow/block;
- category policy;
- upstream/global routing;
- default outbound.

---

## TASK-305 — Xray Policy Decorator / Compiler

Recommended model: `Sol High`

Goal:

Decorate the final upstream-generated Xray config.

Rules:

- one deterministic hook near the end of `GetXrayConfig()`;
- preserve API routing;
- preserve semantic rule/outbound order;
- no second config generator;
- exact no-op when policies are disabled;
- candidate validation and rollback must already exist.

---

## TASK-306 — Policy Simulator

Recommended model: `Luna Medium`

Must reuse real decision/compiler logic.

---

## TASK-307 — Explain Route / Decision

Recommended model: `Luna Medium`

Reuse existing backend route-test capabilities where possible, including user/client context.

Do not implement a second route-testing engine.

---

## TASK-308 — Policy UI

Recommended model: `Luna Medium`

Integrate into existing client/group UX rather than cloning those pages.

---

# PHASE 4 — Policy Engine v2

## TASK-401 — Category Policies

Recommended model: `Luna Medium`

## TASK-402 — Schedules

Recommended model: `Sol Medium`

Timezone and DST tests are mandatory.

## TASK-403 — Temporary Overrides

Recommended model: `Luna Medium`

Expiration must survive restart.

## TASK-404 — Quarantine Policy

Recommended model: `Sol Medium`

Preserve explicitly required management/subscription paths.

## TASK-405 — Managed DNS Policy

Recommended model: `Sol High`

Do not claim perfect DoH prevention.

## TASK-406 — SafeSearch

Recommended model: `Luna Medium`

Only through provider-supported DNS/routing mechanisms.

---

# PHASE 5 — Historical Traffic & Quotas

This phase extends existing upstream accounting instead of rebuilding per-inbound/per-node counters.

## TASK-501 — Custom Date-Range Traffic History

Recommended model: `Luna Medium`

Provide Today / 7 days / 30 days / custom range.

## TASK-502 — Service / Category Traffic Breakdown

Recommended model: `Luna Medium`

Use correlated/classified analytics data.

## TASK-503 — Shared Group Quota

Recommended model: `Sol Medium`

Reuse existing `ClientGroup`.

Shared accounting updates must be atomic.

## TASK-504 — Optional Traffic Multiplier

Recommended model: `Luna Medium`

Use integer/fixed-point accounting.

---

# PHASE 6 — QoS / Traffic Control

## TASK-601 — Shaping Architecture & Capability Detection

Recommended model: `Sol High`

## TASK-602 — Linux Shaping Adapter

Recommended model: `Sol High`

## TASK-603 — Per-client Speed Limit

Recommended model: `Sol High`

## TASK-604 — Rolling/Window Quota

Recommended model: `Sol Medium`

## TASK-605 — Post-quota / Post-expiry Soft Throttle

Recommended model: `Sol Medium`

## TASK-606 — Optional Per-category Cap

Recommended model: `Sol Medium`

Only when category attribution confidence is sufficiently reliable.

---

# PHASE 7 — Security / Anomaly Intelligence

## TASK-701 — Extended IP / Session History

Recommended model: `Luna Medium`

Reuse upstream IP/node observations where available.

## TASK-702 — Account-sharing Risk Engine

Recommended model: `Sol Medium`

No automatic ban by default.

## TASK-703 — ASN / Country Alerts

Recommended model: `Luna Medium`

## TASK-704 — DNS Anomaly Heuristics

Recommended model: `Sol Medium`

Always label results as heuristic.

---

# PHASE 8 — Operations & Product Polish

## TASK-801 — Fork Audit Log UI

Recommended model: `Luna Medium`

Durable audit records must use DB persistence.

## TASK-802 — Webhooks

Recommended model: `Luna Medium`

Extend existing low-volume event patterns.

Do not create a second event bus.

## TASK-803 — Prometheus Export

Recommended model: `Luna Medium`

## TASK-804 — Self-service HWID / Device Portal

Recommended model: `Sol Medium`

Security review required.

Reuse existing `ClientHwid`.

## TASK-805 — Host Visibility per Group / Client

Recommended model: `Luna Medium`

Implement only if current upstream still lacks equivalent functionality.

## TASK-806 — Fleet Dashboard Extensions

Recommended model: `Luna Medium`

Extend existing node/runtime management.

Do not create a parallel node model.

## TASK-807 — Backup/Restore Validation UI & Tests

Recommended model: `Sol Medium`

Build on the recovery work from Phase 0.

## TASK-808 — Multi-node Update Orchestration

Recommended model: `Sol High`

Only after single-node updater/rollback is proven.

---

# PHASE 9 — Community Backlog

Before accepting any new community item:

1. verify current upstream state;
2. verify the request is still relevant;
3. check whether upstream already implemented it;
4. estimate sync/divergence cost;
5. require tests and rollback where relevant.

Do not automatically add new transport/protocol stacks that belong in Xray/upstream.

---

# Before Every Task

1. Merge the previous completed task into `develop`.
2. Ensure `develop` is clean and current.
3. Update `CURRENT_TASK.md`.
4. Create a new feature branch from `develop`.
5. Read only the files listed in `CURRENT_TASK.md`.
6. Implement only the current task.
7. Run all required tests.
8. Commit and push the task branch.
9. Review the result.
10. Merge into `develop` only when the task passes its merge gate.
11. Do not automatically begin the next queue item.

---

# High-risk Tasks That Deserve Sol

Prioritize Sol budget for:

```text
TASK-005
TASK-006
TASK-008
TASK-102
TASK-202
TASK-304
TASK-305
TASK-405
TASK-601
TASK-602
TASK-603
TASK-808
upstream release synchronization
security-sensitive regressions
complex migration/recovery failures
```

TASK-002, TASK-003, TASK-004, and most regular implementation work should start on `Luna Medium`.

Use local Qwen for repository reading, mechanical edits, boilerplate, documentation, and low-risk test expansion when useful.
