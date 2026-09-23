# TASK QUEUE

The coding agent must only execute the task named in `CURRENT_TASK.md`.

Never automatically advance to the next task.

---

# PHASE 0 — Fork Foundation

## TASK-001 — Repository Mapping & Integration Audit

Recommended model: `Sol High`  
Status: `CURRENT`  
Output: `docs/19_REPOSITORY_MAP.md`  
Code changes: `NONE`

## TASK-002 — Fork Identity

Recommended model: `Luna Medium`

Add:

- fork version;
- upstream base version;
- build metadata;
- minimal UI/API exposure;
- tests.

Do not modify updater yet.

## TASK-003 — Fork Extension Registry

Recommended model: `Sol Medium/High` for design, `Luna Medium` for implementation.

Create the smallest extension boundary based on `docs/19_REPOSITORY_MAP.md`.

Only add capabilities actually needed.

## TASK-004 — Feature Flags

Recommended model: `Luna Medium`

Initial flags:

```text
analytics.enabled
dns_intelligence.enabled
policies.enabled
traffic_control.enabled
security_anomaly.enabled
```

All OFF must preserve upstream behavior.

## TASK-005 — Own Updater Source

Recommended model: `Sol High`

Prevent the fork from being overwritten by official upstream binaries.

Deliver:

- fork release source;
- separate fork/upstream/Xray versions;
- stable/dev channels;
- integrity checks;
- tests.

## TASK-006 — Snapshot / Validate / Apply / Rollback Foundation

Recommended model: `Sol High`

Reusable safe-change pipeline:

```text
snapshot
→ validate
→ apply
→ healthcheck
→ commit

failure
→ restore
→ healthcheck
→ audit
```

## TASK-007 — make verify-fork

Recommended model: `Luna Medium`

Create fork verification target.

---

# PHASE 1 — Analytics Foundation

## TASK-101 — Analytics DB Schema

Model: `Luna Medium`

## TASK-102 — Incremental access.log Tailer

Model: `Sol High`

Must support:

- persistent offset;
- rotation/truncate;
- batching;
- bounded backpressure;
- graceful shutdown;
- fuzz/race tests.

## TASK-103 — Event Normalizer

Model: `Luna Medium`

## TASK-104 — Xray API Observer

Model: `Luna Medium`

## TASK-105 — Session Correlator

Model: `Sol Medium/High`

## TASK-106 — Basic Activity API/UI

Model: `Luna Medium`

---

# PHASE 2 — DNS / Destination Intelligence

## TASK-201 — DNS Observer

Model: `Sol Medium`

## TASK-202 — Destination Evidence Fusion

Model: `Sol High`

Fuse:

```text
DNS
Xray destination
SNI
IP
access.log
```

## TASK-203 — ASN / GeoIP

Model: `Luna Medium`

## TASK-204 — Service / Category Classifier

Model: `Luna Medium`

## TASK-205 — DNS Dashboard

Model: `Luna Medium`

## TASK-206 — Service Graph / New-domain Intelligence

Model: `Luna Medium`

---

# PHASE 3 — Policy Engine v1

## TASK-301 — Policy Schema / Repository

Model: `Luna Medium`

## TASK-302 — Policy CRUD API

Model: `Luna Medium`

## TASK-303 — Group Assignment / Client Override

Model: `Luna Medium`

## TASK-304 — Policy Precedence Engine

Model: `Sol High`

## TASK-305 — Xray Policy Compiler

Model: `Sol High`

## TASK-306 — Policy Simulator

Model: `Luna Medium`

Must reuse real compiler/decision logic.

## TASK-307 — Explain Route / Decision

Model: `Luna Medium`

## TASK-308 — Policy UI

Model: `Luna Medium`

---

# PHASE 4 — Policy Engine v2

## TASK-401 — Categories

Model: `Luna Medium`

## TASK-402 — Schedules

Model: `Sol Medium`

Timezone/DST tests mandatory.

## TASK-403 — Temporary Overrides

Model: `Luna Medium`

## TASK-404 — Quarantine

Model: `Sol Medium`

## TASK-405 — Managed DNS Policy

Model: `Sol High`

## TASK-406 — SafeSearch

Model: `Luna Medium`

---

# PHASE 5 — Traffic Accounting

## TASK-501 — Historical Traffic

Model: `Luna Medium`

## TASK-502 — Per-Inbound Breakdown

Model: `Luna Medium`

## TASK-503 — Per-Node Breakdown

Model: `Luna Medium`

## TASK-504 — Shared Group Quota

Model: `Sol Medium`

Atomic accounting required.

---

# PHASE 6 — QoS / Traffic Control

## TASK-601 — Shaping Architecture / Capability Detection

Model: `Sol High`

## TASK-602 — Linux Shaping Adapter

Model: `Sol High`

## TASK-603 — Per-client Speed Limit

Model: `Sol High`

## TASK-604 — Rolling Quota

Model: `Sol Medium`

## TASK-605 — Post-quota Soft Throttle

Model: `Sol Medium`

---

# PHASE 7 — Security / Anomaly

## TASK-701 — IP / Session History

Model: `Luna Medium`

## TASK-702 — Sharing Risk Engine

Model: `Sol Medium`

No automatic ban by default.

## TASK-703 — ASN / Country Alerts

Model: `Luna Medium`

## TASK-704 — DNS Anomaly Heuristics

Model: `Sol Medium`

---

# PHASE 8 — Operations

## TASK-801 — Audit Log UI

Model: `Luna Medium`

## TASK-802 — Webhooks

Model: `Luna Medium`

## TASK-803 — Prometheus Metrics

Model: `Luna Medium`

## TASK-804 — Self-service HWID Portal

Model: `Sol Medium`

Security review required.

## TASK-805 — Host Visibility per Group / Client

Model: `Luna Medium`

Only if upstream has not implemented it.

## TASK-806 — Fleet Dashboard Extensions

Model: `Luna Medium`

## TASK-807 — Backup Validation

Model: `Sol Medium`

## TASK-808 — Automatic Update / Rollback Orchestration

Model: `Sol High`

---

# Before every task

1. Update `CURRENT_TASK.md`.
2. Create a feature branch.
3. Read only mandatory docs for that task.
4. Implement only the current scope.
5. Run required tests.
6. Stop after task completion.
7. Review before moving to the next queue item.

# High-risk tasks that deserve Sol

Prioritize Sol budget for:

```text
TASK-001
TASK-005
TASK-006
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
```
