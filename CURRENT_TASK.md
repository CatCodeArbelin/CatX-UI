# CURRENT TASK

## Work Package

`WP-3A — Policy Data & API`

## Goal

Build the durable fork-owned policy data model and protected API foundation
that WP-3B can later compile into Xray behavior. WP-3A must not alter Xray
routing behavior yet.

## Required

- Policy, policy assignment, policy override, and temporary override models;
- client-target and group-target assignments reusing upstream identities;
- deterministic priority/precedence metadata and inspectable evaluation order;
- SQLite/PostgreSQL fork-owned migrations through the forkext migration seam;
- protected CRUD/read APIs and stable API shapes for policies, assignments,
  overrides, temporary overrides, and precedence inspection;
- validation for targets, references, duplicates, priorities, values, scopes,
  and temporary start/expiry semantics;
- `policies.enabled` feature gating with safe default OFF;
- OpenAPI registration and regenerated artifacts;
- focused unit, persistence, API, disabled-mode, concurrency, and migration
  coverage; `make verify` and `make verify-fork` before merge.

## Hard constraints

- no Xray routing-rule compilation or `DecorateXrayConfig` policy behavior;
- no category blocking, DNS blocking, SafeSearch, QoS, quota, or quarantine
  enforcement;
- no TLS MITM, decrypted HTTPS, payload inspection, cookies, credentials,
  Authorization headers, or request bodies;
- do not duplicate upstream client/group/node/inbound/traffic models;
- expired temporary overrides must never evaluate as active;
- migration failure must not prevent core panel/Xray operation under the
  existing fork safety contract;
- minimal upstream-touch points and no modification of `main`.

## Restrictions

- do not implement WP-3B or WP-3C behavior;
- do not merge into `main`;
- preserve the low-divergence downstream fork architecture;
- use explicit persisted migrations; never mutate schema at runtime.
