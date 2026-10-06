# CURRENT TASK

## Program

`CatX v0.3 Module Qualification`

## Status

`ACTIVE`

## Technical baseline

`5b20b4af7fec6ad5a382e1f6a30022b8ed498d0f`

## Current program stage

`STEP 0 — DEVS MARKING`

Governance and qualification documentation must be completed before product
module qualification begins.

## Current module

None yet.

The first functional module will be:

`M00 — Core Lifecycle / Feature Settings`

but M00 MUST NOT begin until Step 0 is complete.

## Qualification integration branch

`qualification/v0.3-module-review`

## Current maturity policy

Every unqualified CatX-added product module is:

`DEVS`

Only explicit maintainer Human Review PASS may change a module to:

`READY`

## Step 0 scope

Allowed:

- implement centralized module maturity registry;
- display `DEVS` on all unqualified CatX product surfaces;
- verify navigation/UI still builds and works;
- update qualification evidence.

Forbidden:

- functional changes to individual product modules;
- module bug fixing;
- removing any DEVS marker;
- starting M00 before Step 0 completion;
- merge into develop;
- merge into main;
- release;
- tag.

## Module qualification order

The authoritative order is maintained in `ROADMAP.md` and the Module Acceptance
Matrix.

M00 must remain first.

## Mandatory workflow for every module

1. Code Review
2. Test Contract
3. Automated Baseline
4. Fix / Retest Loop
5. en-US / ru-RU Review
6. UI / CSS Review
7. Full Module Regression
8. Maintainer Human Review
9. PASS → READY
10. FAIL → REWORK

## Human approval

Automated qualification is necessary but not sufficient.

The agent MUST stop at:

`WAITING_HUMAN_REVIEW`

Only the maintainer may provide:

`HUMAN REVIEW PASS`

## Required reading

- `AGENTS.md`
- `CURRENT_TASK.md`
- `ROADMAP.md`
- `TASK_QUEUE.md`
- `AGENT_WORKFLOW_AUTOMATION.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/08_FRONTEND_UX.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/38_MODULE_QUALIFICATION_PROTOCOL.md`
- `docs/39_MODULE_ACCEPTANCE_MATRIX.md`

## Current next action

Complete documentation/governance bootstrap.

Then perform:

`STEP 0 — DEVS MARKING`

Do not begin M00 in the same work package unless the maintainer explicitly
requests it.
