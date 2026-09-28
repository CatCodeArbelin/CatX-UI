# CURRENT TASK

## Work Package

`WP-8C — Multi-node Update Orchestration — Stage B: production rollout execution`

## Authorization and guardrails

Stage A is complete at `3fec93b4903c569f57379687b3acbd37802bd06b`. Its
canonical Fork verification is `36367011305` and Release CatX-UI workflow is
`36367011335`. Do not merge Stage A or this branch, modify `main`, or touch
preview/demo data.

Stage B is authorized to implement production rollout execution on this feature
branch. Keep the change modular and low-divergence from upstream. Preserve the
existing `UpdatePanel` compatibility method and the node-local `update.sh`
transactional updater/rollback boundary.

Production fleet mutation must remain disabled by default and may execute only
when both the explicit fleet-update mutation setting and explicit admin
confirmation are present. Dry-run never mutates. Only direct, eligible nodes
may execute. A timeout after POST is ambiguous: reconcile node status before
any retry and require positive evidence before redispatch. Never redispatch
blindly or accept status belonging to another `runId`.

## Required Stage B behavior

- Add a real `runtime.Remote` executor with typed start/update status methods.
- Persist the exact node `runId` for each target and correlate status strictly
  by that value; reject stale or ambiguous evidence.
- Gate canary and subsequent batches on complete convergence and soak.
- Stop-on-failure prevents later dispatches.
- Success requires matching successful update evidence plus fresh heartbeat,
  target `PanelVersion`, and healthy node/Xray state.
- Treat `rolledBack=true` as rollback-attempt evidence only; claim healthy
  rollback only when `rollbackHealthy=true`. Never infer rollback from version
  regression or timeout.
- Preserve unknown/ambiguous state instead of blindly redispatching.
- Extend Fleet UI with run ID, dispatch/result, rollback evidence, version
  convergence, soak progress, failure reason, and safe retry state.
- Emit audit events for authorization, dispatch, completion, rollback evidence,
  abort, and retry.
- Add end-to-end coverage for executor boundary, run ID correlation, stale
  status rejection, POST-timeout ambiguity, canary/batch gating, restart
  recovery, rollback evidence, health timeout, stop-on-failure, and duplicate
  dispatch prevention.

## Verification and completion

Run relevant unit, integration, race, frontend, and migration checks. Run
`make verify`, `make verify-fork`, and canonical Fork verification plus Release
CatX-UI CI until both workflows are green. Perform the final WP-8C production-
safety audit before stopping. Do not merge. Report the final SHA, production
gate, dispatch/evidence protocol, rollback handling, tests, CI IDs, and
remaining limitations.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`
