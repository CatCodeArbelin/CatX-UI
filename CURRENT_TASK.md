# CURRENT TASK

## Work Package

`WP-1A — Analytics Data Foundation`

## Recommended model

**GPT-5.6 Luna Medium**

## Goal

Establish the fork-owned analytics data layer without duplicating existing upstream clients, client groups, nodes, online-state tracking, or traffic counters.

### CP-1A.1 — Analytics domain model

Introduce the minimum models needed for future metadata analytics:

- destination observations/events;
- DNS observations where future ingestion requires persistence;
- correlated network sessions;
- service/category aggregates;
- source/provenance;
- confidence.

Do not store decrypted HTTPS content, cookies, passwords, authorization tokens, request bodies, or message contents.

### CP-1A.2 — Persistence schema

Add fork-owned analytics persistence for SQLite and PostgreSQL.

Requirements:

- reference existing upstream stable client/group/node identifiers where appropriate;
- do not create duplicate client/group/node models;
- do not duplicate upstream traffic counters;
- migrations must be additive and reversible/recoverable according to existing fork safety rules;
- indexes must support client/time/domain/session lookups without premature over-indexing.

### CP-1A.3 — Destination event abstraction

Establish a normalized metadata event suitable for later inputs from:

- `access.log`;
- Xray destination metadata;
- DNS observer;
- SNI/TLS metadata where visible;
- destination IP/port/protocol.

Every inferred value must preserve source/provenance and confidence.

### CP-1A.4 — Xray online/traffic adapter

Reuse existing upstream Xray API/runtime data as read-only analytics inputs.

Do not build:

- another online-user subsystem;
- another per-client traffic counter;
- another per-inbound traffic counter;
- another per-node traffic counter.

Optional analytics failure must never break Xray or the panel.

### CP-1A.5 — Historical aggregation skeleton

Add the minimal aggregation/repository interfaces required for later:

- hourly/daily activity;
- per-client service/category statistics;
- session counts;
- first/last seen.

Do not implement the full UI in WP-1A.

### CP-1A.6 — Retention foundation

Support configurable retention semantics suitable for:

- raw metadata events: short retention;
- sessions: medium retention;
- daily aggregates: long retention.

Do not hard-delete unrelated upstream traffic history.

### Critical constraints

- Preserve the privacy boundary: metadata only, no TLS MITM.
- Analytics is optional and must default OFF through the existing fork feature flag.
- With analytics OFF, behavior must be an exact no-op.
- Analytics errors must not interrupt core panel/Xray operation.
- Reuse WP-0A settings/hooks.
- Reuse upstream normalized clients/groups/nodes/counters.
- Do not implement Policy Engine.
- Do not implement DNS collection yet beyond schema/interfaces required for future ingestion.
- Do not implement the access.log tailer yet; that belongs to WP-1B.
- Do not implement the Client Activity UI yet; that belongs to WP-1C.
- Keep permanent upstream-touch points minimal.
- `make verify-fork` must remain green.

Required verification:

- SQLite migrations;
- PostgreSQL migrations;
- persistence CRUD/query tests;
- retention behavior;
- analytics-disabled no-op behavior;
- concurrency/race tests where relevant;
- `make verify-fork`;
- existing release/build gates.
