# AGENT WORKFLOW AUTOMATION

## Goal

Minimize manual Git operations and avoid one branch per tiny task.

## Branch granularity

Use one branch per work package.

Example:

```text
feature/wp-0a-foundation-guardrails
```

Do not create a separate branch for every checkpoint.

## Agent responsibility

The coding agent may manage:

- branch creation/switching;
- checkpoint commits;
- pushes.

The agent must not merge a work package when required tests are blocked or failing.

## Optional PowerShell helper

Use:

```powershell
.\scripts\catx-task.ps1 -Action status
```

Start a package:

```powershell
.\scripts\catx-task.ps1 -Action start -Branch feature/wp-0a-foundation-guardrails
```

Checkpoint:

```powershell
.\scripts\catx-task.ps1 -Action checkpoint -Message "test: establish upstream compatibility baseline"
```

Finish only after tests actually passed:

```powershell
.\scripts\catx-task.ps1 -Action finish -TestsPassed
```

The `finish` command refuses to merge without the explicit `-TestsPassed` switch.

## Normal human workflow

The human should usually only:

1. choose the model/work package;
2. review the package result;
3. approve finishing the package.

No manual sequence of `switch`, `merge`, `push`, and branch deletion should be required for every checkpoint.
