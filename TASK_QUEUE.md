# TASK QUEUE — WORK PACKAGE MODE

The agent executes one **work package** at a time, not one tiny task per branch.

A work package may contain several checkpoints/commits, but all work must stay inside the package scope.

The agent must stop before entering the next work package.

---

# PHASE 0 — Fork Safety Foundation

## WP-0A — Foundation Guardrails — DONE

Recommended model: `Luna Medium`

Branch:

```text
feature/wp-0a-foundation-guardrails
```

Includes the former tasks:

```text
TASK-002 Baseline Compatibility Suite
TASK-003 Fixed Fork Hook Contracts & Bootstrap
TASK-004 make verify-fork + CI Gate
TASK-007 Fork Settings Facade & Feature Flags
TASK-010 Empty API / OpenAPI / Frontend Registries
```

### Goal

Build the low-risk structural foundation required before product features.

### Checkpoints

#### CP-0A.1 — Baseline Compatibility

- map/reuse existing upstream tests;
- add only missing compatibility tests/fixtures;
- document coverage in `docs/20_BASELINE_COMPATIBILITY.md`;
- no product behavior changes.

#### CP-0A.2 — Fixed No-op Fork Hooks

- add the smallest verified first-party hook boundaries;
- no generic plugin framework;
- all hooks default to exact no-ops;
- baseline behavior must remain unchanged.

#### CP-0A.3 — verify-fork + CI

- add `make verify-fork`;
- keep upstream `make verify` unchanged;
- add fork-specific CI gate.

#### CP-0A.4 — Fork Settings / Feature Flags

Initial flags:

```text
analytics.enabled
dns_intelligence.enabled
policies.enabled
traffic_control.enabled
security_anomaly.enabled
```

All default OFF.

#### CP-0A.5 — Empty API / OpenAPI / Frontend Registries

Add only minimal registries/descriptors for future fork features.

No product feature pages yet.

### Merge gate

WP-0A may merge only when the available test environment can verify the package.

If local Windows cannot run required tests, use an approved alternative such as WSL/CI/container-based verification. Do not claim green tests when they were not executed.

---

## WP-0B — Release Identity & Updater Safety — DONE

Recommended model: `Sol High`

Branch:

```text
feature/wp-0b-release-safety
```

Includes:

```text
TASK-005 CatX-UI Identity & Updater Isolation
TASK-006 Transactional Updater Staging & Automatic Rollback
```

### Goal

Ensure CatX-UI cannot be overwritten by official upstream releases and failed updates restore a known-good installation.

### Checkpoints

1. CatX-UI release identity.
2. Fork/upstream/Xray version separation.
3. Retarget all updater/install paths.
4. Integrity validation.
5. Staged update.
6. Healthcheck.
7. Automatic rollback.
8. Update/rollback tests.

---

## WP-0C — Xray / Database Recovery Safety — CURRENT

Recommended model: `Sol High`

Branch:

```text
feature/wp-0c-runtime-recovery
```

Includes:

```text
TASK-008 Xray Candidate Validation & Known-Good Rollback
TASK-009 DB Import / Xray Start Recovery Gap
```

### Goal

Make runtime-affecting changes recoverable before Policy Engine or traffic-control features exist.

---

# PHASE 1 — Analytics

## WP-1A — Analytics Data Foundation

Recommended model: `Luna Medium`

Includes:

- analytics persistence schema;
- destination event model;
- Xray online/traffic adapter;
- historical aggregation skeleton;
- retention foundation.

Reuse upstream clients/groups/nodes/traffic counters.

Do not create parallel upstream models.

---

## WP-1B — access.log & Session Pipeline

Recommended model: `Sol High`

Includes:

- incremental access.log collector;
- rotation/truncate/restart handling;
- bounded batching;
- parser fuzz/race tests;
- session correlation.

---

## WP-1C — Activity API/UI

Recommended model: `Luna Medium`

Includes:

- client activity API;
- client activity UI;
- basic historical views;
- source/confidence display.

---

# PHASE 2 — DNS / Destination Intelligence

## WP-2A — DNS & Evidence Fusion

Recommended model: `Sol High`

Includes:

- DNS observer;
- destination evidence fusion;
- SNI/TLS/QUIC metadata where observable;
- source/confidence semantics.

No TLS MITM.

---

## WP-2B — Enrichment & Classification

Recommended model: `Luna Medium`

Includes:

- ASN/GeoIP;
- service recognition;
- category classifier;
- first-seen/new-domain intelligence;
- relationship graph.

---

## WP-2C — DNS Intelligence UI / Privacy / Retention

Recommended model: `Luna Medium`

Includes:

- DNS dashboard;
- privacy dashboard;
- retention controls;
- delete-history controls.

---

# PHASE 3 — Policy Engine v1

## WP-3A — Policy Data/API

Recommended model: `Luna Medium`

Includes:

- policy schema/repository;
- CRUD API;
- group assignment;
- client override.

Reuse upstream `ClientGroup` and normalized clients.

---

## WP-3B — Policy Decision Engine & Xray Compiler

Recommended model: `Sol High`

Includes:

- precedence engine;
- deterministic final-config decorator;
- route/order preservation;
- policy fixtures;
- Xray validation;
- rollback integration.

---

## WP-3C — Simulator / Explain / UI

Recommended model: `Luna Medium`

Includes:

- Policy Simulator;
- Explain Decision;
- Explain Route;
- policy UI.

Reuse upstream route-test capabilities where possible.

---

# PHASE 4 — Policy Engine v2

## WP-4A — Categories / Schedules / Temporary Overrides

Recommended model: `Sol Medium`

Includes:

- categories;
- schedules;
- timezone/DST tests;
- temporary overrides.

---

## WP-4B — Quarantine / Managed DNS / SafeSearch

Recommended model: `Sol High`

Includes:

- quarantine;
- managed DNS policy;
- SafeSearch where supported;
- documented DoH limitations.

---

# PHASE 5 — Traffic History & Quotas

## WP-5A — Historical Traffic

Recommended model: `Luna Medium`

Includes:

- custom date ranges;
- service/category breakdown;
- reuse upstream per-client/per-inbound/per-node accounting.

---

## WP-5B — Shared Group Quota / Accounting Extensions

Recommended model: `Sol Medium`

Includes:

- shared group quota;
- atomic accounting;
- optional fixed-point traffic multiplier.

---

# PHASE 6 — QoS / Traffic Control

## WP-6A — Shaping Core

Recommended model: `Sol High`

Includes:

- capability detection;
- shaping adapter;
- Linux implementation;
- reconciliation.

---

## WP-6B — Speed / Rolling Quota / Soft Throttle

Recommended model: `Sol High`

Includes:

- per-client upload/download speed;
- rolling/window quota;
- post-quota/post-expiry throttle;
- optional category cap when classification is reliable.

---

# PHASE 7 — Security / Anomaly

## WP-7A — Risk Intelligence

Recommended model: `Sol Medium`

Includes:

- extended IP/session history;
- sharing-risk engine;
- ASN/country alerts;
- DNS anomaly heuristics.

No automatic ban by default.

---

# PHASE 8 — Operations / Product Polish

## WP-8A — Audit / Webhooks / Metrics

Recommended model: `Luna Medium`

Includes:

- durable fork audit UI;
- webhooks;
- Prometheus export.

Reuse upstream event bus for notifications, not as durable audit storage.

---

## WP-8B — Self-service / Host Visibility / Fleet UI

Recommended model: `Luna Medium`, with security review for self-service.

Includes:

- self-service HWID/device portal;
- host visibility per group/client if upstream still lacks it;
- fleet dashboard extensions.

---

## WP-8C — Multi-node Update Orchestration

Recommended model: `Sol High`

Only after single-node updater/rollback is proven.

---

# Operating Rules

## Branching

One branch per work package, not per checkpoint.

Example:

```text
feature/wp-0a-foundation-guardrails
```

Inside that branch the agent may create multiple focused commits:

```text
test: establish upstream compatibility baseline
feat: add no-op fork hook contracts
ci: add verify-fork gate
feat: add fork feature flags
feat: add empty fork registries
```

## Git automation

The agent may:

- create/switch the work-package branch;
- commit checkpoints;
- push the branch.

The agent must NOT merge into `develop` unless the package merge gate is satisfied.

A merge is forbidden if required tests were not executed successfully.

## Human interaction

Normal workflow should require human review only at work-package boundaries, not after every microtask.

## Expensive model usage

Prioritize Sol for:

```text
WP-0B
WP-0C
WP-1B
WP-2A
WP-3B
WP-4B
WP-6A
WP-6B
WP-8C
upstream synchronization
security/recovery blockers
```

Use Luna Medium for the majority of implementation work.
