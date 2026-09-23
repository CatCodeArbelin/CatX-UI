# 3x-ui Maintained Fork — Engineering Handoff Pack

Status: **final engineering specification / coding-agent handoff**  
Upstream baseline: **MHSanaei/3x-ui v3.8.5**  
Requirement freeze date: **2026-09-23**

This package defines a maintainable downstream fork of 3x-ui that:

- preserves the architecture, behavior, and product model of upstream 3x-ui;
- minimizes long-term divergence from upstream;
- adds modular policy, analytics, DNS intelligence, traffic-control, and operations features;
- can be synchronized with future upstream releases safely;
- never implements TLS interception / HTTPS MITM;
- does not transplant Remnawave architecture;
- owns its updater and release channel;
- requires tests and rollback for risky changes;
- remains GPLv3-compatible.

## Agent read order

Always start with:

1. `AGENTS.md`
2. `CURRENT_TASK.md`

Then read only the files explicitly listed inside `CURRENT_TASK.md`.

Do not read every Markdown document for every task.

## Project principle

**This is not a replacement panel. It is a maintained 3x-ui fork.**

Upstream behavior has priority. Custom functionality must live in isolated modules and connect through minimal integration points.

## Recommended Git model

```text
upstream/main            # upstream remote-tracking branch
develop                  # fork integration branch
main                     # tested fork releases only
feature/*                 # one feature / one PR
sync/upstream-vX.Y.Z      # upstream synchronization PR
release/vX.Y.Z-custom.N   # release preparation
```

Suggested tags:

```text
v3.8.5-custom.1
v3.8.5-custom.2
v3.9.0-custom.1
```

## Quality commands

Upstream verification:

```bash
make verify
```

Fork verification:

```bash
make verify-fork
```

`make verify-fork` must include upstream verification plus fork-specific tests.

## Hard non-goals

- TLS MITM / HTTPS interception.
- Root CA distribution for traffic inspection.
- Cookie/password/token/body collection.
- Remnawave architectural concepts as the fork foundation.
- Rewriting 3x-ui from scratch.
- Scattering custom logic throughout upstream files.
- Blind production auto-update without test and rollback gates.
