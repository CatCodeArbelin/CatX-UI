# CURRENT TASK

## Work Package

`WP-3C — Simulator / Explain / UI`

## Goal

Build the human-facing inspection and management layer for the policy engine
without changing WP-3A/WP-3B decision semantics. The simulator must reuse the
real decision engine and explain-route must reuse the real Xray compiler.

## Required

- side-effect-free simulation for client, group, domain, IP/CIDR, category,
  service, timestamp, and persisted policy state;
- simulator never mutates storage, feature flags, Xray config, or runtime;
- inspectable explain-decision output containing final action, winner,
  assignment/override source, precedence stage, priority, client/group source,
  temporary status/expiry, deterministic tie-break data, and reasons for
  winners and losers;
- explain-route derived from the actual compiler representation, including
  CatX rule identity, destination, user selector, outbound, ordering, source,
  and whether a rule would be emitted;
- protected simulator/explain APIs through the existing fork API boundary;
- explicit stable schemas and regenerated OpenAPI artifacts;
- policy management UI using existing 3x-ui containers, spacing, Ant Design
  tokens, tables, forms, drawers/modals, typography, states, and responsive
  conventions;
- UI for policy CRUD, client/group assignments, explicit and temporary
  overrides, simulator, decision explanation, and route explanation;
- frontend must consume existing protected APIs rather than duplicate CRUD
  semantics;
- tests for equivalence, zero side effects, explanations, route previews,
  validation, disabled/empty/loading/error states, frontend behavior, and
  concurrency where applicable;
- run full `make verify-fork` and Release CatX-UI through GitHub Actions.

## Hard constraints

- do not create a second decision engine or Xray compiler;
- do not apply simulation results or restart/reload Xray;
- do not change WP-3A/WP-3B precedence or compiler behavior except for a
  clearly required compatibility adapter;
- no TLS MITM, decrypted HTTPS, payload inspection, credentials, cookies,
  Authorization headers, or request bodies;
- no schedules UI, quarantine, managed DNS enforcement, SafeSearch, QoS,
  quotas, anomaly actions, or multi-node distribution in WP-3C;
- preserve feature-disabled upstream-compatible no-op behavior;
- do not modify or merge into `main`;
- do not merge WP-3C into `develop` automatically; stop on the feature branch
  after both required CI workflows are green.

## Safety

Simulation and explanation are read-only. Malformed or unsupported policy
state must be represented explicitly and conservatively; it must never cause
an Xray mutation. Keep API errors stable and testable.
