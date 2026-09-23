# CURRENT TASK

## Task ID

`PHASE-0 / TASK-001`

## Title

Repository Mapping & Fork Integration Audit

## Recommended model

**Sol High** preferred.

Fallback: **Luna Medium**.

Reason: this task defines the permanent integration boundaries of the fork.

---

## Mandatory first reads

Read these files completely before doing anything else:

1. `AGENTS.md`
2. `FINAL_TZ.md`
3. `docs/00_MASTER_SPEC.md`
4. `docs/01_ARCHITECTURE.md`
5. `docs/03_ROADMAP.md`
6. `docs/04_UPSTREAM_SYNC.md`
7. `docs/05_TESTING_ROLLBACK_RELEASE.md`
8. `skills/repository-cartographer/SKILL.md`

Do **not** read every Markdown file in the repository unless the task requires it.

---

## Goal

Map the real current 3x-ui repository and identify the **minimum permanent integration points** required for this maintained fork.

This task is analysis-only.

Do not modify production code.

---

## Required repository mapping

Inspect and document actual code paths for:

- application startup/lifecycle;
- database initialization;
- migrations;
- Client model/service;
- Group model/service;
- Inbound model/service;
- Node model/service;
- Xray config generation;
- Xray runtime/process management;
- Xray API integration;
- routing;
- client-specific routing support;
- traffic/statistics;
- online-client/IP state;
- `access.log` handling;
- scheduled jobs;
- event bus/events;
- backend controllers/routes;
- OpenAPI generation;
- frontend route registration;
- frontend navigation/menu;
- client/group UI;
- updater;
- backup/restore;
- settings/feature storage;
- tests and CI.

---

## Critical questions

### 1. Fork registration boundary

Find the smallest practical place to introduce a fork extension entry point.

Conceptually:

```go
forkext.RegisterRoutes(...)
forkext.RegisterJobs(...)
forkext.RegisterMigrations(...)
forkext.RegisterEventSubscribers(...)
```

Do not force these exact names if the real repository suggests a better boundary.

### 2. Database integration

Determine:

- where upstream migrations live;
- how SQLite/PostgreSQL compatibility is handled;
- where fork migrations can hook in with minimal divergence;
- which upstream file(s) become permanent migration integration hotspots.

### 3. Xray integration

Determine:

- where upstream builds Xray config;
- where routing rules are constructed;
- where per-user/client routing is represented;
- where a fork policy compiler could safely decorate/augment config;
- how candidate config can be validated before reload/restart.

### 4. Analytics integration

Determine:

- how current online state is obtained;
- current `access.log` behavior;
- where an incremental collector can run independently;
- what existing stats/events can be reused.

### 5. Jobs and events

Determine:

- where jobs are scheduled;
- where event subscribers can be registered;
- whether the existing event bus is sufficient.

### 6. Frontend integration

Determine:

- safest place for fork routes/features;
- how API types/OpenAPI are generated;
- how to avoid hand-editing generated files;
- how fork pages can be added without replacing upstream navigation.

### 7. Updater

Determine:

- exact official upstream update source;
- exact files/functions that must change so the fork can never overwrite itself with official upstream binaries.

### 8. Compatibility

Identify how to prove:

```text
all fork feature flags OFF
=
upstream behavior
```

for:

- Xray config;
- subscriptions;
- client management;
- routing;
- startup.

---

## Required output

Create:

`docs/19_REPOSITORY_MAP.md`

Use this structure:

```text
# Repository Map

## Current upstream architecture
## Application lifecycle
## Database and migrations
## Clients / Groups / Inbounds / Nodes
## Xray config generation
## Xray runtime/API
## Routing
## Traffic / statistics
## Online state / IP tracking
## access.log
## Jobs / schedulers
## Event system
## Backend routes / API
## OpenAPI / generated code
## Frontend routes / navigation
## Updater
## Backup / restore
## Test infrastructure
## Proposed fork integration points
## Permanent upstream-touch hotspots
## Spec assumptions that were correct
## Spec assumptions that need correction
## Features already present upstream
## Duplicate functionality we should NOT implement
## Recommended Phase 0 implementation sequence
## Risks / blockers
```

For each proposed integration point include:

```text
File:
Function/type:
Purpose:
Fork hook:
Why this is minimal:
Expected upstream-sync conflict risk:
LOW / MEDIUM / HIGH
```

---

## Forbidden in this task

Do not:

- implement `forkext`;
- change database schema;
- change updater;
- create Policy Engine;
- create Analytics;
- create DNS Observer;
- create QoS;
- change frontend;
- refactor upstream code;
- rename/move upstream files.

---

## Completion criteria

Task is complete when:

- `docs/19_REPOSITORY_MAP.md` exists;
- all important paths are backed by actual repository inspection;
- permanent fork integration hotspots are identified;
- duplicate/upstream-existing features are identified;
- no production source files were modified.

After creating the document:

**STOP.**

Do not begin TASK-002 automatically.
