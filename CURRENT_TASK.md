# CURRENT TASK

## Work Package

`WP-0A — Foundation Guardrails`

## Recommended model

**Luna Medium**

## Branch

```text
feature/wp-0a-foundation-guardrails
```

The agent is allowed to manage this work-package branch, create checkpoint commits, and push it.

Do not merge into `develop` until the merge gate is satisfied.

---

## Mandatory reads

1. `AGENTS.md`
2. `CURRENT_TASK.md`
3. `TASK_QUEUE.md`
4. `docs/19_REPOSITORY_MAP.md`
5. `docs/05_TESTING_ROLLBACK_RELEASE.md`
6. `docs/15_DEFINITION_OF_DONE.md`
7. `skills/test-release-guardian/SKILL.md`
8. `skills/feature-implementer/SKILL.md`

Do not read every Markdown file.

---

## Important current state

The previous TASK-002 attempt created `docs/20_BASELINE_COMPATIBILITY.md`, but local verification was blocked because the environment reported:

```text
required Go: 1.27.1
installed Go: 1.23.3
make: unavailable
```

Therefore baseline verification is NOT considered complete yet.

Do not claim green verification until the required checks actually execute in an approved environment.

---

# WP-0A Scope

## CP-0A.1 — Baseline Compatibility

First:

1. inspect the existing `docs/20_BASELINE_COMPATIBILITY.md`;
2. inspect existing upstream tests;
3. preserve/reuse existing coverage;
4. add only missing baseline tests/fixtures;
5. determine the best available verification environment.

Allowed verification environments, in preference order:

```text
existing working local toolchain
WSL toolchain
existing repository-supported container/CI workflow
GitHub CI
```

Do not silently install arbitrary system software.

If no verification environment is available, stop with one concise blocker report instead of pretending the checkpoint is complete.

## CP-0A.2 — Fixed No-op Fork Hooks

After baseline coverage is established:

- implement only the smallest fixed first-party hook contracts identified in `docs/19_REPOSITORY_MAP.md`;
- no generic plugin framework;
- no product behavior;
- exact no-op defaults;
- preserve baseline behavior.

Create a focused checkpoint commit.

## CP-0A.3 — `make verify-fork` + CI

Add an additive fork verification target.

Do not alter the semantics of upstream `make verify`.

Add CI invocation using existing repository patterns.

Create a focused checkpoint commit.

## CP-0A.4 — Fork Settings / Feature Flags

Add a fork-owned settings facade with these initial flags:

```text
analytics.enabled
dns_intelligence.enabled
policies.enabled
traffic_control.enabled
security_anomaly.enabled
```

Defaults:

```text
OFF
```

Requirements:

- SQLite + PostgreSQL compatible;
- no unrelated upstream setting-field sprawl;
- all OFF preserves baseline behavior.

Create a focused checkpoint commit.

## CP-0A.5 — Empty API / OpenAPI / Frontend Registries

Add minimal no-op registries/descriptors for future fork features.

Verified integration areas are documented in `docs/19_REPOSITORY_MAP.md`.

Do not add product feature pages.

Create a focused checkpoint commit.

---

# Git Automation Rules

At the beginning:

1. inspect current branch and working tree;
2. if currently on the old `feature/task-002-baseline-compatibility` branch, preserve its uncommitted/committed baseline-document work;
3. create or switch to:

```text
feature/wp-0a-foundation-guardrails
```

from the current `develop` baseline;
4. carry forward only legitimate WP-0A work.

During the package:

- create one focused commit per checkpoint;
- push after each checkpoint;
- do not rewrite upstream history;
- do not merge into `develop`.

---

# Merge Gate

WP-0A is complete only if:

- baseline compatibility coverage exists;
- no-op hooks preserve baseline behavior;
- `make verify-fork` exists;
- CI gate exists;
- feature flags default OFF;
- empty registries exist;
- required relevant tests have actually passed in an approved environment;
- no production behavior change occurs with all fork features OFF;
- no unrelated refactor exists.

If verification is blocked:

**STOP. DO NOT MERGE.**

Report the exact blocker.

---

# Completion Behavior

When all WP-0A checkpoints and verification pass:

1. ensure working tree is clean;
2. ensure checkpoint commits are pushed;
3. provide a concise summary of commits/tests;
4. stop.

Do not start WP-0B automatically.
