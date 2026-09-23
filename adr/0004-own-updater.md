# ADR-0004: Fork-owned Updater

Status: Accepted

## Decision

Fork releases and updates come from a fork-owned release source.

The official upstream updater must never replace fork binaries.

## Safety model

```text
backup
→ stage
→ install
→ migrate
→ healthcheck
→ commit
or rollback
```
