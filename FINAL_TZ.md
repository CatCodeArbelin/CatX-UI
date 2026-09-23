# FINAL SPEC — Maintained 3x-ui Fork

This is the concise top-level specification. Detailed requirements live in `docs/*`.

## 1. Product

Build a long-lived downstream fork of 3x-ui that preserves upstream behavior and adds four isolated layers:

```text
POLICY
TRAFFIC CONTROL
INSIGHT
OPERATIONS
```

Baseline at requirement freeze: upstream `v3.8.5`.

## 2. Non-negotiable principle

Upstream 3x-ui remains the architectural center.

Implement new functionality inside fork-specific modules and connect through minimal integration points.

If a feature requires widespread upstream-core edits, redesign or postpone it.

## 3. Foundation

Required first:

- fork identity/version;
- own updater/release source;
- separate fork/upstream/Xray versioning;
- stable/dev channels;
- feature flags;
- extension registration boundary;
- config snapshots;
- validate/apply/healthcheck;
- rollback;
- audit log;
- `make verify-fork`;
- upstream-sync workflow.

## 4. Insight

### access.log

Incremental tailer with:

- offset persistence;
- rotation;
- truncate handling;
- batching;
- bounded queues;
- graceful shutdown;
- fuzz/race tests.

No full rescans.

### Unified Destination Observer

Fuse:

```text
Xray API
access.log
DNS
requested destination
SNI metadata where visible
TLS/QUIC metadata
plaintext HTTP Host only
destination IP
```

Normalize → correlate → session → aggregate.

### Analytics features

- ASN/GeoIP;
- service/category recognition;
- source/confidence;
- Client Activity;
- DNS Dashboard;
- traffic history;
- service relationship graph;
- first-seen/new-domain intelligence;
- privacy dashboard;
- custom date ranges;
- per-client/inbound/node/service/category breakdowns.

## 5. Policy Engine

Entities:

- Policy;
- Assignment;
- Override;
- Temporary Override;
- Schedule.

Features:

- group policy;
- client overrides;
- allow/block domains;
- categories;
- outbound selection;
- schedules;
- temporary access;
- quarantine;
- managed DNS;
- optional SafeSearch;
- Policy Simulator;
- Explain Decision / Explain Route.

Use existing Xray routing/client semantics.

## 6. Traffic Control

Core:

- shared group quota;
- per-inbound accounting;
- per-node accounting.

Later/high-risk:

- per-client speed limit;
- rolling quota;
- soft throttle after quota/expiry;
- optional category caps;
- optional traffic/accounting multiplier.

Use an isolated shaping adapter and capability detection.

## 7. Security intelligence

Metadata-only:

- IP/session history;
- ASN/country changes;
- concurrent sessions;
- sharing-risk score;
- alerts;
- optional DNS anomaly heuristics.

No automatic ban by default.

## 8. Subscription/client usability

Architecture-compatible only:

- group subscription link if reusable from existing engine;
- Host visibility per group/client;
- self-service HWID/device management;
- safe subscription-token actions;
- small routing/subscription improvements that expose existing Xray capabilities.

Do not build a Remnawave-style replacement control plane.

## 9. Routing/observatory

Extend upstream:

- outbound/node health notifications;
- quality dashboard;
- Explain Route;
- explicit smart failover later if upstream mechanisms support it.

## 10. Operations

- backup validation;
- config diff;
- audit;
- webhooks;
- Prometheus;
- fleet dashboard extensions;
- remote Xray version control later if cleanly supported;
- logging level/retention controls;
- safe updater;
- release/rollback tooling.

## 11. Privacy boundary

Permanently excluded:

```text
TLS MITM
Root CA
HTTPS decryption
cookies
Authorization headers
passwords
decrypted messages
decrypted HTTP bodies
```

## 12. Remnawave boundary

Do not transplant:

- Squads architecture;
- External Squads;
- mandatory Config Profiles;
- separate subscription control plane;
- Remnawave node architecture;
- Remnawave source code.

## 13. Testing

Before merge:

```bash
make verify
make verify-fork
```

Plus applicable:

- SQLite;
- PostgreSQL;
- migration upgrade;
- race;
- fuzz;
- Xray validation;
- startup smoke;
- subscription smoke;
- feature-off compatibility;
- rollback smoke.

Red test = no merge.

## 14. Rollback

Risky change flow:

```text
snapshot
→ candidate
→ validate
→ apply
→ restart/reload
→ healthcheck
→ commit
```

Failure:

```text
restore known-good
→ restart
→ healthcheck
→ audit/report
```

## 15. Upstream synchronization

```text
fetch/tag
→ sync branch
→ semantic review
→ merge
→ verify
→ fork verify
→ migration/integration
→ smoke
→ PR
→ RC if needed
→ main
```

No blind auto-merge.

## 16. Database

- explicit migrations;
- SQLite/PostgreSQL;
- no destructive one-step migrations;
- batched analytics writes;
- configurable retention;
- atomic group quota counters;
- no decrypted HTTPS content.

## 17. AI development workflow

- Local Qwen: repository mapping, broad reading, mechanical work.
- Luna Medium: daily scoped implementation.
- Sol High: architecture, concurrency, Xray, updater, upstream merge, security, difficult review.
- One task per session.
- Git/docs/tests are project memory, not chat history.

## 18. Source of truth priority

If documents conflict:

```text
AGENTS.md
→ accepted ADR
→ module-specific docs
→ FINAL_TZ.md
→ old chat discussions
```
