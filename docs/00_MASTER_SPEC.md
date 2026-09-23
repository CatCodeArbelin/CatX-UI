# MASTER SPEC — 3x-ui Maintained Fork

## 1. Product statement

Create a maintained downstream fork of 3x-ui that preserves upstream product behavior while adding four modular layers:

```text
3x-ui upstream-compatible fork
├── POLICY
├── TRAFFIC CONTROL
├── INSIGHT
└── OPERATIONS
```

The fork must not become a different panel with a different architecture.

## 2. Baseline

At requirement freeze:

- upstream: `MHSanaei/3x-ui`
- stable baseline: `v3.8.5`
- backend: Go / Gin / GORM
- DB: SQLite + PostgreSQL
- frontend: React + Ant Design + Vite + TypeScript
- Xray managed as runtime/child process.

Before implementation, agents must re-check current upstream because it evolves quickly.

## 3. Strategic objectives

### O1. Low divergence

Custom subsystems require clear boundaries and minimal upstream hooks.

### O2. Safe upstream updates

New stable upstream releases must integrate through a controlled PR with full verification.

### O3. Network visibility without content interception

Provide useful DNS/destination/session analytics without HTTPS decryption.

### O4. Human-friendly policy management

Expose reusable client/group policies that compile to existing Xray capabilities.

### O5. Production safety

Risky changes use snapshot → validate → apply → healthcheck → rollback.

## 4. Product modules

### 4.1 Fork Core

- fork identity/version;
- own release source;
- own updater;
- stable/dev channels;
- feature flags;
- extension registry;
- config snapshot;
- validation;
- automatic rollback;
- audit events;
- upstream compatibility tests.

### 4.2 Policy Engine

- policies;
- group/client assignment;
- inheritance;
- client overrides;
- allow/block domains;
- categories;
- schedules;
- temporary overrides;
- outbound selection;
- quarantine;
- managed DNS;
- optional SafeSearch;
- Policy Simulator;
- Explain Decision / Explain Route.

### 4.3 Insight

- incremental access.log collector;
- Xray API observer;
- DNS observer;
- SNI metadata when available;
- plaintext HTTP Host only;
- TLS/QUIC metadata;
- destination IP;
- ASN/GeoIP;
- normalized events;
- correlation/deduplication;
- service/category classification;
- traffic history;
- activity timeline;
- anomaly signals;
- retention/privacy controls.

### 4.4 Traffic Control

- shared group quota;
- rolling/window quota;
- per-client speed limit;
- optional post-quota/expiry throttle;
- per-inbound traffic;
- per-node traffic;
- optional traffic/accounting multiplier;
- isolated enforcement adapter.

### 4.5 Operations

- updater/rollback;
- backup/restore validation;
- node/fleet health;
- notifications/webhooks;
- limited self-service portal;
- audit log;
- Prometheus;
- upstream sync tooling.

## 5. Mandatory privacy boundary

Never implement:

```text
TLS MITM
Root CA issuance for inspection
HTTPS decryption
decrypted HTTP body capture
cookie capture
Authorization token capture
password capture
message/content capture
```

## 6. Compatibility contract

When fork features are disabled:

- upstream-generated config remains semantically equivalent;
- existing API clients keep working;
- existing databases migrate safely;
- existing subscriptions remain functional;
- existing inbounds/outbounds/routing remain usable.

## 7. Performance targets

Initial engineering targets:

- no full access.log rescan;
- bounded collector memory;
- batched analytics writes;
- aggregated dashboard queries;
- configurable retention;
- no network I/O under global locks;
- analytics failure must not stop Xray;
- enrichment failure must not block traffic unless explicitly required by policy.

## 8. Suggested default retention

```text
raw destination events   7 days
raw DNS events           7 days
sessions                30 days
daily aggregates       365 days
audit log              180 days
```

All configurable.

## 9. Data minimization

Prefer aggregates such as:

```text
service=YouTube
category=Streaming
bytes=...
sessions=...
```

instead of permanent raw event storage.

## 10. Release gates

A release cannot reach `main` unless:

- `make verify` passes;
- `make verify-fork` passes;
- migration tests pass;
- smoke installation passes;
- previous-custom upgrade passes;
- Xray starts/reloads;
- rollback is validated for relevant changes;
- no unresolved critical/high security issue;
- license/notices are preserved.

## 11. Success criteria

The fork is successful if:

1. upstream updates remain manageable;
2. custom modules can be disabled independently;
3. DNS/activity analytics work without HTTPS interception;
4. per-client/group policy is understandable without raw JSON;
5. bad config/update does not easily strand a remote server;
6. feature code remains isolated and testable;
7. community-driven additions improve 3x-ui instead of replacing it.
