THIS IS A 3x-ui FORK. ALL CHANGES MUST BE MODULAR AND MUST NOT BREAK UPSTREAM SOURCE STRUCTURE, ARCHITECTURE, OR BEHAVIOR.

# AGENTS.md

These rules are mandatory for every AI/coding agent working in this repository.

## 1. Primary goal

Maintain a **low-divergence downstream fork** of `MHSanaei/3x-ui`.

Do not turn a feature request into an upstream refactor.

Do not rewrite an existing upstream service because a different architecture looks cleaner.

Priority order:

1. upstream compatibility;
2. correctness;
3. testability;
4. rollback/recovery;
5. modularity;
6. performance;
7. developer convenience.

## 2. Before making any change

The agent MUST:

1. read `AGENTS.md`;
2. read `CURRENT_TASK.md`;
3. read only the documents listed in `CURRENT_TASK.md`;
4. inspect the actual current upstream code involved in the task;
5. identify the minimum integration points;
6. identify all upstream files expected to change;
7. stop and redesign if custom logic starts spreading through upstream core files.

## 3. Upstream-touch budget

By default, a fork feature should:

- add new fork-specific files/modules;
- modify as few upstream files as possible;
- use registration hooks, adapters, decorators, or lifecycle entry points;
- preserve existing upstream semantics.

If a feature requires edits across many unrelated upstream core files, that is an architectural warning.

Prefer boundaries conceptually similar to:

```go
forkext.RegisterRoutes(...)
forkext.RegisterJobs(...)
forkext.RegisterMigrations(...)
forkext.RegisterEventSubscribers(...)
forkext.DecorateXrayConfig(...)
```

Exact names may differ based on the real repository.

## 4. Forbidden without an accepted ADR

Do not:

- rewrite upstream controller/service/database layers;
- change existing API semantics casually;
- change DB structure without an explicit migration;
- spread fork logic across many upstream handlers;
- hand-edit generated files when a generator exists;
- copy Remnawave source code;
- introduce TLS MITM or HTTPS decryption;
- store decrypted HTTP bodies, cookies, Authorization headers, or passwords;
- use `access.log` as the only source of real-time online state;
- treat IP-only heuristics as proof of account sharing;
- auto-ban a user from one heuristic signal;
- auto-merge upstream directly into production;
- update production without backup, healthcheck, and rollback.

## 5. Tests and merge — mandatory rule

**BEFORE ANY MERGE INTO `develop`, AND ESPECIALLY INTO `main`, ALL RELEVANT TESTS MUST PASS.**

Minimum:

```bash
make verify
make verify-fork
```

If DB/migrations changed:

- SQLite migration tests;
- PostgreSQL migration tests;
- upgrade from supported previous schemas;
- forward-recovery/rollback strategy where applicable.

If collectors/parsers changed:

- unit tests;
- rotation/truncate tests;
- malformed-input tests;
- fuzz tests;
- race tests;
- benchmark/regression tests when relevant.

If Xray config/compiler behavior changed:

- generated config validation;
- Xray start/reload smoke test;
- policy/config fixtures;
- feature-disabled compatibility test.

If tests fail:

**MERGE IS FORBIDDEN.**

## 6. Rollback — mandatory rule

Every feature that can affect:

- Xray config;
- routing;
- updater;
- migrations;
- node control;
- traffic enforcement;
- DNS enforcement;

must have an explicit recovery path.

Risky apply flow:

```text
snapshot
→ validate
→ apply
→ healthcheck
→ commit state
```

Failure flow:

```text
FAIL
→ restore previous known-good state
→ restart/reload as required
→ healthcheck
→ report/audit failure
```

"Fix it manually through SSH" is not an acceptable default rollback plan.

## 7. Feature flags

Large fork subsystems should be independently disableable when practical:

```text
analytics.enabled
dns_intelligence.enabled
policies.enabled
traffic_control.enabled
security_anomaly.enabled
```

When disabled, behavior should remain as close as possible to upstream.

## 8. Database rules

Every persisted model change requires:

- explicit migration;
- SQLite compatibility;
- PostgreSQL compatibility;
- index review;
- retention/storage impact review;
- migration tests.

No runtime schema mutation outside the migration mechanism.

## 9. Concurrency rules

For custom goroutine/channel/shared-state code:

```bash
go test -race ./...
```

Do not hold global/service locks during:

- network calls;
- DNS calls;
- filesystem I/O;
- long DB operations;
- slow parsing/batching.

Concurrent services require graceful shutdown.

## 10. Xray boundary

Xray remains an external runtime/core.

Do not fork xray-core for panel features.

New functionality should:

- use existing Xray config/API where possible;
- detect capabilities;
- degrade safely when optional APIs are unavailable;
- preserve existing generated config when fork features are disabled;
- validate config before reload/restart.

## 11. Privacy boundary

Allowed metadata includes:

- client identity;
- DNS domains;
- destination domain/IP;
- SNI when visible;
- protocol metadata;
- port/network;
- inbound/outbound;
- ASN/GeoIP;
- traffic counters;
- timestamps;
- policy decisions.

Do not collect:

- HTTPS bodies;
- cookies;
- Authorization headers;
- credentials;
- decrypted messages;
- passwords.

## 12. Git discipline

One task = one feature branch / one focused PR.

Do not combine:

```text
feature
+ unrelated refactor
+ repository-wide formatting
```

Avoid mass upstream renames/moves.

## 13. Upstream synchronization

Production follows upstream stable releases.

Recommended flow:

```text
fetch upstream
→ create sync/upstream-vX.Y.Z
→ merge upstream release/tag
→ resolve semantic conflicts
→ make verify
→ make verify-fork
→ integration/migration/smoke
→ review
→ develop
→ release candidate if needed
→ main
```

Do not use rebasing of long-lived downstream history as the standard sync strategy.

## 14. Definition of Done

A feature is not done until it has, where applicable:

- working implementation;
- tests;
- migration;
- documentation;
- i18n;
- OpenAPI/generated-code updates;
- rollback/recovery review;
- feature-disabled compatibility coverage;
- no new sensitive-data collection;
- successful `make verify-fork`.

## 15. Decision rule

When uncertain, prefer the solution that:

- touches less upstream code;
- is easier to remove;
- is easier to disable;
- is easier to test;
- is more likely to survive the next upstream release.
