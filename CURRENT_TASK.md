# CURRENT TASK

## Program

`CatX v0.3 Module Qualification`

## Status

`ACTIVE`

## Technical baseline

`5b20b4af7fec6ad5a382e1f6a30022b8ed498d0f`

## Current program stage

`MODULE QUALIFICATION`

Step 0 DEVS marking is complete. M00 code review, contract freeze, automated
baseline, en-US / ru-RU localization review, and the hosted UI / CSS / Product
Visual Review are recorded on the dedicated module branch. The next governed
action is maintainer Human Review; no Human Review approval is included in this
package.

## Current module

`M00 — Core Lifecycle / Feature Settings`

## Current workflow stage

`WAITING_HUMAN_REVIEW`

## Current maturity

`DEVS`

## Qualification integration branch

`qualification/v0.3-module-review`

## Current maturity policy

Every unqualified CatX-added product module is:

`DEVS`

Only explicit maintainer Human Review PASS may change a module to:

`READY`

## Step 0 scope

Status: `COMPLETE`

Allowed:

- implement centralized module maturity registry;
- display `DEVS` on all unqualified CatX product surfaces;
- verify navigation/UI still builds and works;
- update qualification evidence.

Forbidden:

- functional changes to individual product modules;
- module bug fixing;
- M00 qualification tests or acceptance fixes;
- removing any DEVS marker;
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

`MAINTAINER HUMAN REVIEW OF M00`
