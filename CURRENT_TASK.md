# CURRENT TASK

## Work Package

`WP-4A — Categories / Schedules / Temporary Overrides`

## Goal

Extend the existing policy engine with conservative category decisions, durable
timezone-aware schedules, and the temporary override behavior required by the
roadmap without changing WP-3A/WP-3B semantics or creating parallel models.

## Required

- reuse the authoritative analytics/enrichment category vocabulary and
  provenance/uncertainty contracts;
- add durable schedules with explicit IANA timezone, weekdays, local start/end
  times, cross-midnight handling, restart-safe persistence, injectable time, and
  deterministic DST behavior;
- extend the existing WP-3A temporary override model only where required;
- extend the existing decision engine and Xray compiler, not a second evaluator;
- add minimum protected API/UI for category policies, schedule CRUD/assignment,
  temporary override management, and active/next-active state;
- regenerate OpenAPI and add comprehensive unit, integration, API, frontend,
  SQLite/PostgreSQL-appropriate, and race/concurrency coverage;
- run full fork verification and Release CatX-UI CI.

## Required behavior

- unknown or ambiguous classifications must not cause destructive decisions
  unless an explicit safe rule contract supports them;
- schedules support normal, cross-midnight, weekday, DST-forward, DST-backward,
  non-DST, and timezone-change cases, with explicit invalid-timezone rejection;
- schedules never mutate their underlying policy;
- disabled schedules and expired temporary overrides are immediately inactive;
- preserve deterministic precedence, idempotent compilation, exact disabled/no-
  policy no-op behavior, and upstream Xray config preservation.

## Hard constraints

- no duplicate client/group/node/category models;
- no TLS MITM, decrypted HTTPS, payload inspection, SafeSearch, quarantine,
  managed DNS, QoS, quotas, anomaly actions, or multi-node distribution;
- no upstream service rewrite or second recovery mechanism;
- do not implement WP-4B;
- keep WP-4A on its own feature branch and merge only after all required tests
  and CI are green; do not modify `main`.
