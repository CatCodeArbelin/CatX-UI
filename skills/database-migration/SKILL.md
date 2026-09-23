# SKILL: Database Migration Engineer

## Rules

- SQLite + PostgreSQL;
- explicit migration;
- populated upgrade fixture;
- no destructive one-step migration;
- atomic shared quota updates;
- indexes based on query patterns;
- retention plan for event tables.

## Required tests

```text
empty → current
upstream baseline → current
previous custom → current
SQLite
PostgreSQL
```

## Output

```text
tables/columns
indexes
data backfill
locking/downtime risk
rollback/forward recovery
```
