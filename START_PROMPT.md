# Startup Prompt

Copy the text below into the first coding-agent session.

---

Read `AGENTS.md` and `CURRENT_TASK.md` first.

Then read only the documents and skill files explicitly listed in `CURRENT_TASK.md`.

Execute the current task strictly within its defined scope.

Important rules:

- This repository is a maintained downstream fork of 3x-ui.
- Upstream architecture and behavior have priority.
- Do not introduce broad upstream refactors.
- Do not begin any task other than the one defined in `CURRENT_TASK.md`.
- If the current task forbids production-code changes, do not modify production code.
- Verify every repository fact from the actual current source code rather than assuming the specification is perfectly up to date.
- Follow all completion criteria in `CURRENT_TASK.md`.
- Create every required output file.
- When the current task is complete, stop. Do not automatically move to the next item in `TASK_QUEUE.md`.

For TASK-001 specifically, this is an analysis-only repository-mapping pass. The required output is `docs/19_REPOSITORY_MAP.md`. Do not modify production code.
