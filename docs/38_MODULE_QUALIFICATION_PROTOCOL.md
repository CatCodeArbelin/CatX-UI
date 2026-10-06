# CatX Module Qualification Protocol

## Goal

Prove each CatX-added product module independently before release.

Whole-product CI remains valuable but does not replace per-module proof.

## Maturity states

### DEVS

The module exists but has not completed the current release qualification
protocol.

### READY

The module passed all applicable automated gates and explicit maintainer Human
Review.

Only the maintainer can authorize `DEVS → READY`.

## Qualification state machine

```text
DEVS
  ↓
CODE_REVIEW
  ↓
CONTRACT
  ↓
TEST_SPEC
  ↓
AUTOMATED_BASELINE
  ↓
RED ───────────────┐
  ↓                │
FIX                │
  ↓                │
RETEST ────────────┘
  ↓
AUTOMATED_GREEN
  ↓
LOCALIZATION
  ↓
UI_REVIEW
  ↓
FULL_MODULE_REGRESSION
  ↓
WAITING_HUMAN_REVIEW
  ├── FAIL → HUMAN_REVIEW_FAILED → REWORK → CODE/TEST REVIEW → FIX LOOP
  └── PASS → READY
```

### Code review phase

Before product changes, trace the real implementation.

Identify:

- module purpose;
- user promise;
- data sources;
- persistence;
- feature flags;
- API;
- runtime integration;
- Xray integration;
- upstream integration;
- frontend;
- lifecycle;
- failure behavior;
- recovery;
- unsupported capabilities.

### Contract phase

Write the expected observable behavior independently of the current
implementation.

Unsupported functionality must be stated explicitly.

Do not write a contract merely describing existing code.

### Test specification phase

Define deterministic proof for every important product promise.

Use production boundaries.

Tests should fail if the real feature is broken.

### Automated baseline

Run the tests before fixing discovered behavior where practical.

Classify:

`PASS`

`FAIL`

`BLOCKED`

### Fix loop

For every confirmed product failure:

```text
reproduce
→ root cause
→ minimal architecture-correct fix
→ focused test
→ complete module regression
```

Repeat until green.

### Test-change policy

A valid failing test must not be weakened.

Changing a frozen test requires:

`TEST CHANGE JUSTIFICATION`

containing:

- old expectation;
- reason it was wrong;
- correct contract;
- evidence that the new test does not hide a defect.

### Localization gate

Required languages:

`en-US`

`ru-RU`

Check all visible module strings and state labels.

### UI gate

Backend/runtime functionality must be green first.

Then review:

- usability;
- layout;
- responsive behavior;
- terminology;
- empty/error/loading states;
- unsupported states;
- dangerous actions;
- upstream visual consistency.

After frontend changes, rerun the complete module suite.

### Human Review

Human Review is an external approval gate.

The candidate is frozen at an exact SHA.

The maintainer receives:

- exact SHA;
- isolated launch instructions;
- checklist;
- limitations.

Automation stops.

### Human Review PASS

On explicit PASS:

- record evidence;
- set module READY;
- remove DEVS;
- update roadmap/current task;
- merge to qualification integration branch;
- move to next module.

### Human Review FAIL

On FAIL:

- keep DEVS;
- assign HR finding IDs;
- set HUMAN_REVIEW_FAILED / REWORK;
- inspect why tests missed the issue;
- add regression coverage where appropriate;
- repair;
- rerun complete qualification;
- repeat Human Review.

There is no maximum retry count.

### Cross-module issue

A module must not absorb unrelated module work.

Record:

`BLOCKED_BY <module/shared foundation>`

Fix the dependency separately, then rerun the blocked module from the required
qualification stage.

### Branch model

Integration:

`qualification/v0.3-module-review`

Modules:

`qualify/v0.3-mNN-<name>`

Only Human-approved READY modules merge into the integration branch.

No direct module merge into `develop/main`.

### Evidence

Every READY module must record:

- source SHA;
- test SHA;
- hosted run IDs;
- automated results;
- runtime results;
- localization result;
- Human Review result;
- known limitations.

### Final program closure

After all modules are READY, run whole-product integration qualification before
any `develop`/release promotion.
