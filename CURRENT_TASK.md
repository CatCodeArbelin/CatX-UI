# CURRENT TASK

## Work Package

`WP-8C — Multi-node Update Orchestration — Stage B: DONE`

## Authorization and guardrails

Stage A is complete at `3fec93b4903c569f57379687b3acbd37802bd06b`. Its
canonical Fork verification is `36367011305` and Release CatX-UI workflow is
`36367011335`. Do not merge Stage A or this branch, modify `main`, or touch
preview/demo data.

Stage B was authorized to implement production rollout execution on its feature
branch. It is complete and was accepted on the final feature tip after the
canonical Fork verification and Release CatX-UI workflows passed. Keep the
change modular and low-divergence from upstream. Preserve the
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

The final WP-8C tip was `cb760398`.

Canonical verification passed:

- Fork verification: `36390706017`
- Release CatX-UI: `36390706149`

WP-8C is complete. RC-1 is the next authorized work package and must be
performed on its own feature branch. Do not modify `main`, create a release
tag, or update production as part of this completion record.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`
