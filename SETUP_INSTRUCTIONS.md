# Setup Instructions

## Purpose

Copy this handoff pack into the root of your forked `MHSanaei/3x-ui` repository without replacing upstream project files.

## Important

Do **not** replace the upstream `README.md`.

This package intentionally uses `FORK_HANDBOOK.md` instead.

## Root files to copy

Place these files directly in the repository root:

```text
AGENTS.md
CURRENT_TASK.md
TASK_QUEUE.md
FINAL_TZ.md
FORK_HANDBOOK.md
SETUP_INSTRUCTIONS.md
START_PROMPT.md
```

## Directories to copy

Copy these directories into the repository root:

```text
docs/
skills/
adr/
templates/
```

## Expected repository layout

```text
3x-ui/
├── AGENTS.md
├── CURRENT_TASK.md
├── TASK_QUEUE.md
├── FINAL_TZ.md
├── FORK_HANDBOOK.md
├── SETUP_INSTRUCTIONS.md
├── START_PROMPT.md
│
├── docs/
│   ├── 00_MASTER_SPEC.md
│   ├── 01_ARCHITECTURE.md
│   ├── 02_FEATURE_CATALOG.md
│   ├── 03_ROADMAP.md
│   ├── 04_UPSTREAM_SYNC.md
│   ├── 05_TESTING_ROLLBACK_RELEASE.md
│   ├── 06_DATABASE_MIGRATIONS.md
│   ├── 07_SECURITY_PRIVACY.md
│   ├── 08_FRONTEND_UX.md
│   ├── 09_ANALYTICS_DNS.md
│   ├── 10_POLICY_ENGINE.md
│   ├── 11_QOS_TRAFFIC_CONTROL.md
│   ├── 12_UPDATER_IDENTITY.md
│   ├── 13_COMMUNITY_BACKLOG.md
│   ├── 14_NON_GOALS.md
│   ├── 15_DEFINITION_OF_DONE.md
│   ├── 16_LEGAL_GPL_NOTES.md
│   ├── 17_AI_WORKFLOW.md
│   └── 18_SOURCE_AUDIT.md
│
├── skills/
│   ├── repository-cartographer/
│   │   └── SKILL.md
│   ├── feature-implementer/
│   │   └── SKILL.md
│   ├── upstream-sync/
│   │   └── SKILL.md
│   ├── test-release-guardian/
│   │   └── SKILL.md
│   ├── policy-xray-compiler/
│   │   └── SKILL.md
│   ├── analytics-dns/
│   │   └── SKILL.md
│   ├── database-migration/
│   │   └── SKILL.md
│   └── security-privacy-review/
│       └── SKILL.md
│
├── adr/
│   ├── 0001-maintained-downstream-fork.md
│   ├── 0002-no-tls-mitm.md
│   ├── 0003-realtime-xray-api-history-accesslog.md
│   └── 0004-own-updater.md
│
└── templates/
    ├── TASK.md
    ├── ADR.md
    ├── PR.md
    └── RELEASE.md
```

## First agent run

Use Sol High if available.

Give the agent only the prompt stored in `START_PROMPT.md`.

The prompt instructs the agent to:

1. read `AGENTS.md`;
2. read `CURRENT_TASK.md`;
3. read only the files listed by the current task;
4. perform TASK-001;
5. create `docs/19_REPOSITORY_MAP.md`;
6. make no production-code changes;
7. stop.

## After TASK-001

Do not let the agent automatically continue.

Review:

```text
docs/19_REPOSITORY_MAP.md
```

Then update `CURRENT_TASK.md` to TASK-002.

`TASK_QUEUE.md` is the execution plan, but it is not permission to execute all tasks.

## Daily operating rule

Every agent session starts with:

```text
AGENTS.md
CURRENT_TASK.md
```

The current task decides what other files must be read.

Do not tell the agent to read every `.md` file for every task.

## Branch rule

Use one branch per task:

```text
feature/task-002-fork-identity
feature/task-003-extension-registry
...
```

Upstream synchronization uses:

```text
sync/upstream-vX.Y.Z
```

## Merge rule

Before merge:

```bash
make verify
make verify-fork
```

Plus task-specific tests.

If required tests fail, do not merge.

## Rollback rule

Risky configuration/update changes require:

```text
snapshot
→ validate
→ apply
→ healthcheck
→ commit
```

Failure requires restore of the previous known-good state.
