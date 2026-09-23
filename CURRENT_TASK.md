# CURRENT TASK

## Task ID

`PHASE-0 / TASK-002`

## Title

Baseline Compatibility Suite

## Recommended model

**Luna Medium**

Sol is not required unless a genuine architectural ambiguity or difficult upstream regression is discovered.

---

## Mandatory first reads

Read these files completely before changing code:

1. `AGENTS.md`
2. `CURRENT_TASK.md`
3. `docs/19_REPOSITORY_MAP.md`
4. `docs/05_TESTING_ROLLBACK_RELEASE.md`
5. `docs/15_DEFINITION_OF_DONE.md`
6. `skills/test-release-guardian/SKILL.md`

Read additional source files only as required by the task.

Do not read every project Markdown file.

---

## Goal

Create a baseline compatibility suite that captures the current upstream-compatible behavior of CatX-UI before any fork hooks or product features are added.

The purpose is to create a regression safety net for later Phase 0 work.

No intentional runtime behavior change is allowed.

---

## Baseline

Current fork baseline:

```text
Upstream base: v3.8.5
Fork state: upstream v3.8.5 + documentation only
Fork product features: none
```

Use the actual checked-out repository state as authoritative.

---

## Required analysis before writing tests

Inspect the existing test infrastructure and determine which behaviors are already sufficiently covered.

Do not duplicate upstream tests unnecessarily.

Create new compatibility fixtures/tests only where needed to establish the fork baseline.

At minimum evaluate coverage for:

- Xray config generation;
- standard subscription output;
- JSON subscription output;
- Clash/Happ output where supported by existing tests;
- client CRUD;
- client groups;
- client-to-inbound attachment behavior;
- routing behavior;
- startup smoke;
- SQLite;
- PostgreSQL.

---

## Required compatibility guarantees

The baseline suite must make it practical to prove later that:

```text
CatX-UI with fork hooks/features disabled
=
current upstream-compatible behavior
```

Especially for:

### Xray configuration

Preserve semantic ordering where order matters.

Do not normalize/sort routing or outbound arrays if that could change semantics.

### Subscriptions

Capture relevant current behavior without introducing a second subscription engine.

### Client / group behavior

Reuse normalized upstream clients/groups as the source of truth.

### Database

Do not introduce fork schema in this task.

Use existing SQLite/PostgreSQL test infrastructure.

### Startup

Use existing smoke/install/startup mechanisms where practical.

Do not invent an unrelated test harness if upstream already has one.

---

## Required deliverables

### 1. Baseline tests / fixtures

Add only the tests or fixtures required to close important compatibility gaps.

### 2. Documentation

Create:

`docs/20_BASELINE_COMPATIBILITY.md`

It must document:

```text
# Baseline Compatibility Suite

## Baseline commit / upstream version
## Existing upstream coverage reused
## New compatibility coverage added
## Xray config guarantees
## Subscription guarantees
## Client / group guarantees
## Routing guarantees
## SQLite coverage
## PostgreSQL coverage
## Startup / smoke coverage
## Known gaps
## Commands to run
## Merge gate
```

### 3. Test commands

Run the smallest relevant test set during implementation, then run the required final verification available at this stage.

`make verify-fork` does not exist yet, so do not invent success output for it.

Run `make verify` if the current environment supports it.

If an environment dependency prevents part of the suite from running, document the exact blocker and do not claim that test passed.

---

## Forbidden in this task

Do not:

- add fork hooks;
- add `forkext`;
- add feature flags;
- change updater behavior;
- add fork DB tables;
- change Xray runtime behavior;
- add Policy Engine;
- add Analytics;
- add DNS Observer;
- add QoS;
- refactor unrelated upstream code;
- change product behavior just to make tests easier.

---

## Git scope

Expected changes should be limited to:

```text
tests / existing test fixtures where appropriate
docs/20_BASELINE_COMPATIBILITY.md
minimal test-helper changes only if strictly necessary
```

If production source code appears to require modification, stop and explain why before changing it.

---

## Completion criteria

TASK-002 is complete only when:

- existing upstream test coverage has been mapped and reused;
- important baseline gaps have compatibility coverage;
- `docs/20_BASELINE_COMPATIBILITY.md` exists;
- no intentional runtime behavior changed;
- relevant tests pass;
- any unrun test is explicitly documented with its environment blocker;
- the working diff contains no unrelated refactor;
- the task result is ready to become the baseline used by TASK-003.

After completion:

**STOP.**

Do not start TASK-003 automatically.
