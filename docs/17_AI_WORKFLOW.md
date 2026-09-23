# AI / Vibe-Coding Workflow

## Purpose

Keep a large fork feasible on a limited cloud-model budget and avoid context bloat.

## Suggested model roles

```text
Local Qwen3-Coder-Next
→ broad repository reading
→ mechanical edits
→ boilerplate/tests/docs
→ low-cost first implementation

Luna Medium
→ daily implementation
→ scoped backend/frontend tasks
→ CRUD/migrations/UI
→ regular refactors

Sol High
→ architecture
→ Xray semantics
→ concurrency
→ security-sensitive changes
→ upstream sync
→ difficult regressions
→ high-risk PR review

Astra
→ only exceptional blockers/security/very difficult merges
```

## One task = one session

Bad:

```text
Policy → DNS → updater → QoS → upstream merge
```

Good:

```text
policy-db
policy-api
policy-compiler
dns-collector
activity-ui
upstream-sync-3.9.0
```

## Repository is memory

Canonical project memory:

```text
Git
AGENTS.md
CURRENT_TASK.md
docs/
ADR/
tests
```

Do not depend on old chat context.

## Context packet

For most tasks provide only:

1. `AGENTS.md`
2. `CURRENT_TASK.md`
3. relevant module spec
4. relevant code
5. relevant tests
6. current diff

## Task decomposition

Never prompt:

```text
Build Policy Engine.
```

Instead:

```text
policy schema
repository
CRUD
assignment
override
compiler
simulator
schedule
temporary override
frontend
tests
```

## Spend Sol where mistakes are expensive

Do not spend Sol on:

- renames;
- translations;
- simple forms;
- boilerplate DTOs;
- routine fixtures.

Use Sol for:

- architecture;
- update conflicts;
- data races;
- policy precedence;
- updater;
- rollback;
- security;
- complex Xray behavior.
