# CatX v0.3 architecture corrections

This document records the corrective boundary implemented on
`fix/v0.3-architecture-corrections`. It is intentionally limited to the
review findings from the v0.3 architecture audit.

## Runtime lifecycle

Fork-owned schema preparation and in-process activation are now separate:

```text
persist desired flags
  -> PrepareRuntimeFromSettings (explicit, idempotent schema preparation)
  -> restart/apply boundary
  -> ReloadRuntimeFromSettings (configuration only)
```

`RegisterMigrations` owns the initial startup preparation pass. A generic
panel restart never invokes a migration. The runtime snapshot tracks the
desired configuration that has been prepared separately from the
configuration currently active in memory.

If preparation fails, the desired setting remains persisted, the active
services remain on their last known-good configuration, and the error is
exposed as `error`/`restart_required` state. Retrying the explicit preparation
boundary is safe. The preparation sequence is retry-safe rather than claiming
cross-database atomic DDL: SQLite or PostgreSQL may retain an idempotent table
created before a later module fails, and the next retry validates and resumes
from that state.

The database bootstrap logs an optional CatX preparation/activation failure
and continues the upstream panel startup. A healthy upstream panel therefore
remains available to inspect and retry the fork feature state.

## Panel restart semantics

`POST /panel/api/setting/restartPanel` acknowledges that a restart is
scheduled, not that it has already completed. The existing delayed panel
lifecycle remains authoritative. On startup, a CatX activation error keeps the
last known-good in-process runtime and does not turn an optional fork failure
into a fatal panel start error.

## Policy enforcement boundary

Policy repository mutations mark the existing Xray restart/apply path as
pending. The policy package does not start or own another Xray controller.
The existing candidate validation, health check, and rollback path clears the
pending state only after a successful apply. A failed apply records an error
while preserving the known-good Xray process/configuration.

The compiler removes only CatX-owned rule tags before emitting deterministic
replacement rules. It also removes stale CatX rules when the effective policy
set becomes empty or the policy feature is disabled; upstream routing rules
are preserved. Generic users still receive no policy effect without an
identity-bearing inbound client and a persisted effective policy assignment.

## Traffic control boundary

Quota accounting continues to use upstream cumulative client counters and the
existing fixed-window checkpoint. Rate shaping is not presented as active for
ordinary Xray users: without a production attribution provider that proves
stable client identity in the datapath, non-zero rate fields are rejected and
the UI explains that shaping is unsupported. Quota persistence remains
available.

## Frontend and localization

Analytics, audit, and policy pages distinguish successful empty collections
from failed requests. Failed requests clear stale rows and render an explicit
error state, so a backend failure cannot appear as an empty dataset.

The historical `/sponsors` URL redirects to the canonical CatX sponsors route;
the upstream duplicate page is no longer registered. Sponsor placement labels
use a dedicated localization object. Duplicate JSON object keys are rejected
by the frontend localization contract test; the known Russian duplicate and
the English sponsor key collision are removed without bulk-copying malformed
objects into every catalog.

## Fleet scope

Fleet remains a presentation and operational-context layer over the upstream
Nodes inventory. It summarizes node health and counters and links operators to
the upstream node-management surface; it is not an independent node inventory
or control plane.

## Verification expectations

The branch adds SQLite lifecycle/recovery tests, an opt-in PostgreSQL lifecycle
test using `XUI_TEST_PG_DSN`, effective Xray policy decoration tests, stale-rule
removal coverage, unsupported-rate persistence coverage, frontend error/empty
state tests, canonical-route coverage, and duplicate-key localization checks.

The final review report records which repository, database, race/fuzz, security,
OpenAPI, and real-panel/Xray gates were executed versus unavailable in the
execution environment. This branch does not merge, release, tag, or modify
`main`/`develop`.
