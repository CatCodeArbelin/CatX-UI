# CURRENT TASK

## Work Package

`WP-0C — Xray / Database Recovery Safety`

## Recommended model

**GPT-5.6 Sol High**

## Goal

Harden runtime configuration application and database restore/import paths so a bad candidate configuration or restored database cannot leave CatX-UI/Xray unusable.

### CP-0C.1 — Xray known-good snapshot

Preserve the minimum state required to restore the currently working Xray runtime/configuration before applying a candidate.

Reuse existing Xray lifecycle/config-generation mechanisms.

Do not introduce a second Xray config generator.

### CP-0C.2 — Candidate validation

Required flow:

`build candidate through existing GetXrayConfig()`

`→ validate using installed Xray`

`→ only then allow activation`

Invalid candidates must not replace the known-good runtime.

### CP-0C.3 — Safe activation

Preserve existing hot-diff/hot-apply behavior where supported.

Required conceptual flow:

`snapshot`

`→ generate candidate`

`→ validate`

`→ apply through existing Xray lifecycle`

`→ healthcheck`

`→ commit known-good state`

### CP-0C.4 — Automatic rollback

On failed apply/start/healthcheck:

`restore previous known-good state`

`→ restart/reload`

`→ healthcheck`

`→ report/audit failure`

Rollback failure must be distinguishable from candidate failure.

### CP-0C.5 — Database import/recovery gap

Harden the path where a DB import/restore succeeds structurally but produces an Xray configuration that cannot start.

Cover:

- SQLite;
- PostgreSQL where applicable;
- previous DB/config restoration;
- Xray validation;
- healthcheck;
- failure-path tests.

### Critical constraints

- Preserve upstream `GetXrayConfig()` as the authoritative config builder.
- Preserve existing Xray runtime/process management.
- Preserve existing hot-apply behavior.
- Minimize permanent upstream-touch points.
- Reuse the WP-0B recovery primitives where appropriate instead of creating another unrelated rollback system.
- Do not implement Policy Engine.
- Do not implement Analytics.
- Do not implement DNS Intelligence.
- Do not implement QoS.
- Do not modify `main`.
- Do not start the next work package.
- `make verify-fork` must remain green.

Required verification must include candidate-validation failures, failed Xray startup, failed healthcheck, successful rollback, failed rollback reporting, and DB-import recovery paths.

## Administrative boundary

This task activates WP-0C only. Do not begin implementation until the work-package branch is prepared and reviewed.
