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

Provide a truthful, recoverable lifecycle for CatX feature configuration while
leaving upstream lifecycle ownership and feature-disabled behavior intact.

User promise:

An operator can distinguish what is persisted as desired from what is actually
running. Enabling, disabling, restart, startup reconstruction, preparation
failure, retry, and malformed optional settings are represented truthfully in
the Feature Settings API. A failed optional CatX transition does not make the
unrelated panel unusable.

Supported:

- namespaced persisted feature flags with default-OFF behavior;
- dependency validation without implicit dependency enablement;
- explicit schema preparation followed by runtime activation;
- runtime deactivation that removes CatX effects;
- startup reconstruction from persisted desired state;
- the existing `restartPanel`/SIGHUP lifecycle boundary;
- database import followed by startup/restart reconstruction;
- deterministic error and retry semantics, including partial preparation;
- SQLite and PostgreSQL semantic parity;
- typed Feature Settings API state;
- idempotent activation and deactivation.

Unsupported / outside M00:

- Analytics, DNS, Policy, QoS, Risk, Audit, Portal, Fleet, Fleet Updates, or
  Sponsors business behavior;
- proving that a policy blocks a destination, that analytics records a real
  event, or that a traffic shaper enforces a rate;
- frontend visual, localization, or CSS qualification;
- changing the upstream Xray or panel lifecycle architecture.

Inputs:

- namespaced desired flags in the existing settings key/value table;
- declared feature dependencies;
- the database dialect and current schema;
- explicit preparation, startup, restart, or import lifecycle triggers.

Outputs:

- persisted desired state;
- active in-process runtime state;
- `feature_off`, `initializing`, `restart_required`, `active`, or `error`
  lifecycle state;
- preparation/activation errors suitable for operator retry;
- a readable Feature Settings API response even when one optional flag is
  malformed.

State contract:

1. Desired persisted state and active runtime state are separate concepts.
2. Persisting `desired=true` never by itself means `ACTIVE`.
3. `ACTIVE` is allowed only after all required preparation and activation work
   for that desired configuration succeeds.
4. An activation attempt from an inactive runtime that fails reports
   `active=false` and `state=error`. If a previously known-good runtime is
   intentionally preserved during a later failed transition, `active=true`
   may describe that prior runtime only when `state=error` makes clear that the
   new desired configuration is not active.
5. Disabling a feature removes its CatX runtime effect after the approved apply
   boundary; the pending interval is `restart_required`, not `feature_off`.
6. Feature-OFF behavior has no CatX runtime effect and remains upstream
   compatible.
7. Generic runtime reload/restart applies already-prepared state only. It must
   not silently become a schema migration boundary.
8. Schema preparation occurs through the approved explicit activation or
   startup/import lifecycle, is idempotent, and exposes partial-progress
   recovery rather than pretending an atomic rollback that is not guaranteed.
9. A failed transition has a deterministic retry path and does not tear down a
   previously healthy optional runtime.
10. Malformed persisted flags make the affected item `enabled=false`,
    `active=false`, `state=error` (or an explicitly equivalent typed error)
    without turning the whole Feature Settings response into a 500.
11. Invalid dependency combinations are rejected atomically and never silently
    activated.
12. Startup reconstructs persisted desired state truthfully.
13. `restartPanel` acknowledges a pending request only; it does not claim final
    activation before the restart/apply lifecycle completes.
14. SQLite and PostgreSQL expose equivalent product semantics.
15. Optional CatX failure is never represented as healthy empty data or a
    healthy `active` state for the failed desired configuration.
16. Repeating successful activation or disable is an idempotent no-op in
    observable state and persistence.
17. Maturity `DEVS` is independent of runtime state and remains unchanged
    throughout M00 qualification until explicit Human Review PASS.

Persistence:

Desired flags use the existing upstream settings table. M00 introduces no new
persisted lifecycle model. Feature-owned schemas are prepared through the
existing fork migration boundaries; partial DDL is recorded by tests rather
than assumed to roll back.

Runtime effect:

The lifecycle adapter configures or disables fork consumers and publishes a
snapshot. It does not become a second Xray controller and does not own later
module business decisions.

Feature-off effect:

All managed CatX consumers are inactive, settings remain readable, upstream
routes and behavior remain available, and no feature-specific effect is
reported as active.

Failure and recovery:

Preparation errors preserve the prior known-good runtime where available,
publish a typed error, retain the desired setting for retry, and leave the
panel usable. A retry reruns explicit preparation, then the normal reload/apply
boundary. Partial preparation is recorded exactly and may be completed by a
subsequent idempotent retry.

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
