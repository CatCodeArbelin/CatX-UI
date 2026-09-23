# Roadmap

## Phase 0 — Fork Safety/Foundation

Deliver:

- clean fork from upstream stable;
- upstream remote;
- `develop` / `main`;
- fork version metadata;
- own release source;
- own updater source;
- `make verify-fork`;
- feature flags;
- extension registration boundary;
- snapshot/rollback foundation.

Exit criteria:

- features OFF behave like upstream;
- updater cannot overwrite fork with official upstream binary;
- CI green.

## Phase 1 — Analytics Foundation

Implement:

1. incremental `access.log` tailer;
2. normalized event model;
3. batch persistence;
4. rotation/truncate recovery;
5. Xray API observer;
6. session history;
7. retention;
8. basic Activity UI.

## Phase 2 — DNS / Destination Intelligence

Implement:

- DNS observer;
- domain/destination fusion;
- SNI metadata;
- ASN/GeoIP;
- service/category classifier;
- session deduplication;
- DNS Dashboard;
- traffic by service/category.

## Phase 3 — Policy Engine v1

Implement:

- policy CRUD;
- group assignment;
- client override;
- domain allow/block;
- outbound selection;
- compiler;
- Policy Simulator;
- Explain Decision;
- audit.

## Phase 4 — Policy Engine v2

Implement:

- categories;
- schedules;
- temporary overrides;
- quarantine;
- managed DNS;
- optional SafeSearch.

## Phase 5 — Traffic Accounting

Implement:

- persistent historical traffic;
- per-inbound;
- per-node;
- shared group quota.

## Phase 6 — Traffic Control / QoS

Experimental first:

- shaping adapter;
- one Linux backend;
- per-client speed limit;
- rolling quota;
- soft throttle.

Do not ship stable until reconnect/restart/multi-node behavior is proven.

## Phase 7 — Security / Anomaly

- improved session/IP history;
- risk engine;
- ASN/country alerts;
- DNS anomaly heuristics;
- alerts.

No automatic ban by default.

## Phase 8 — Operations

- self-service portal;
- webhooks;
- Prometheus;
- fleet extensions;
- restore validation;
- update orchestration.

## Phase 9 — Community Backlog

Only accept features that:

- have real demand;
- fit 3x-ui architecture;
- do not duplicate upstream work;
- have reasonable maintenance cost.
