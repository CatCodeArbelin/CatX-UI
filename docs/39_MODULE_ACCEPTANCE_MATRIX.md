# CatX Module Acceptance Matrix

## Purpose

This document defines what must be proven for each CatX product module before
its DEVS marker may be removed.

The actual implementation remains authoritative for module boundaries.

Tests must prove the promises documented here rather than merely reproduce the
current implementation.

## Common evidence fields

Every module entry must define:

- Purpose
- User Promise
- Supported
- Unsupported
- Inputs
- Outputs
- Persistence
- Runtime Effect
- Feature-Off Effect
- Failure Semantics
- Unit Tests
- Synthetic Integration
- Database Tests
- Runtime Tests
- Frontend Tests
- en-US Review
- ru-RU Review
- Human Review Checklist

## M00 — Core Lifecycle / Feature Settings

Purpose:

Provide truthful enable/disable/runtime lifecycle for CatX modules.

Minimum proof:

- all features OFF preserves upstream behavior;
- desired enabled state is distinct from active state;
- feature activation prepares required state safely;
- ACTIVE is reported only after successful activation;
- activation failure produces truthful ERROR/pending state;
- disabling removes fork runtime effect;
- restartPanel state remains truthful;
- full restart reconstructs correct state;
- generic runtime reload does not perform uncontrolled migration;
- malformed persisted flags do not make the panel unusable;
- dependency violations are rejected;
- SQLite recovery works;
- PostgreSQL recovery works.

## M01 — Analytics / Activity

Purpose:

Provide truthful client activity and traffic analytics from supported evidence.

Deterministic synthetic scenario should include:

client:
`catx-test-user`

upload:
`100 MiB`

download:
`250 MiB`

Expected proof:

synthetic production-path observations
→ persistence/aggregation
→ API
→ exact expected counters

Healthy zero observations:

`EMPTY`

Backend/runtime failure:

`ERROR`

No error may be represented as healthy empty data.

## M02 — DNS Intelligence

Purpose:

Expose supported destination evidence without TLS MITM.

Synthetic evidence:

DNS:
`video.example.test`

SNI:
`video.example.test`

Verify:

- observation ingestion;
- correlation;
- source/confidence;
- persistence;
- API;
- UI;
- retention/delete semantics where applicable.

Unsupported encrypted/opaque evidence remains explicitly unsupported.

## M03 — Policy Engine

Purpose:

Convert deterministic policy decisions into effective CatX routing changes
without becoming a second Xray control plane.

Synthetic client:
`catx-test-user`

Synthetic destination:
`video.example.test`

Required proof:

policy CRUD
→ decision resolution
→ compiler
→ effective Xray configuration

Verify:

- create;
- read;
- update;
- delete;
- group assignment;
- client override;
- schedule where supported;
- simulator/explain consistency;
- runtime apply;
- disable removes CatX routing delta;
- re-enable restores expected behavior;
- feature OFF preserves upstream config semantics.

Database-only CRUD is not sufficient proof.

## M04 — Traffic History / Quotas

Purpose:

Provide correct historical accounting and quota state.

Synthetic traffic:

upload `100 MiB`

download `250 MiB`

Use deterministic additional deltas to cross quota threshold.

Verify:

- accounting;
- history;
- exact quota usage;
- exhaustion;
- reset/window;
- persistence;
- restart behavior;
- group quota behavior where supported.

## M05 — Traffic Control / QoS

Purpose:

Expose only traffic-control behavior that can actually be enforced.

Separate:

`SUPPORTED:`

quota/accounting behavior proven by runtime

from:

`UNSUPPORTED:`

generic speed shaping when production attribution is unavailable

Never allow UI to imply unsupported rate shaping is active.

If shaping is supported in a specific environment, acceptance must prove actual
runtime enforcement, not saved configuration.

## M06 — Security / Risk

Purpose:

Provide explainable risk/anomaly evidence without automatically treating a
single heuristic as proof.

Synthetic scenarios should exercise:

- IP history;
- ASN/country change where supported;
- session evidence;
- DNS anomaly evidence where supported.

Verify:

- deterministic scoring/flags;
- persistence;
- explanation;
- no default automatic punitive action from one heuristic.

## M07 — Audit / Webhooks / Metrics

Purpose:

Provide durable operational evidence and supported outbound notifications.

Synthetic administrative actions:

- login;
- settings update;
- policy action;
- client action.

Verify durable audit events.

Healthy zero events:

`EMPTY`

Backend failure:

`ERROR`

For webhooks verify:

- local fake endpoint;
- payload;
- retry/failure behavior;
- no sensitive secret leakage.

Metrics must reflect documented operational state only.

## M08 — Portal / Self-service

Purpose:

Provide isolated client self-service without sharing admin authentication.

Acceptance sequence:

issue credential
→ portal authentication
→ session/CSRF
→ allowed resource access
→ revoke credential
→ old credential rejected

Verify ownership and isolation boundaries.

## M09 — Fleet

Purpose:

Provide fleet-oriented presentation/context over upstream Nodes.

Fleet must not introduce a second node inventory.

Verify:

- upstream Nodes remain authority;
- node state presentation is truthful;
- empty state is understandable;
- navigation to upstream node management works;
- no duplicate node-control plane.

## M10 — Fleet Updates

Purpose:

Safely orchestrate supported node update operations.

Synthetic nodes:

- online node;
- second online node;
- unavailable/offline node.

Verify deterministic state machine:

plan
→ ready
→ dispatch
→ running
→ terminal state

Verify legal/illegal transitions for:

abort
retry
reconcile

Frontend actions must match backend legal transitions.

## M11 — Sponsors

Purpose:

Provide one canonical sponsor product surface and CatX sponsor authority.

Use a local fake provider.

Verify:

- local CRUD;
- remote provider validation;
- cache;
- fallback;
- placement;
- SSRF protections;
- canonical route;
- management entry;
- legacy route convergence;
- no second sponsor data authority.

## Implementation boundary notes

The current implementation inventory is the authority for these rows. The
current M00–M11 grouping is intentional:

- M01/M02 share analytics storage and collection code but have separate
  activity and DNS contracts.
- M04 spans analytics history and the fork quota/accounting packages.
- M07 groups the audit package's durable audit, webhook, and metrics surfaces;
  webhooks and metrics do not currently have separate managed runtime flags.
- M09 presents the upstream Nodes authority and does not own a second fleet
  inventory.

Any future implementation boundary change must update this matrix before the
affected module is qualified.

## Final rule

Passing unit tests alone is never sufficient for a module whose user promise
depends on database integration, runtime state, Xray, network behavior, or
frontend interaction.
