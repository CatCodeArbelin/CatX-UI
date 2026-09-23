# Database and Migrations

## Principles

- SQLite remains first-class.
- PostgreSQL remains first-class.
- No fork feature may silently require an external DB.
- Every model change requires migration.
- Analytics retention must prevent unbounded growth.

## Suggested fork tables

Conceptual:

```text
fork_policies
fork_policy_assignments
fork_policy_overrides
fork_policy_schedules
fork_policy_temp_overrides

fork_destination_events
fork_dns_events
fork_sessions
fork_daily_aggregates
fork_service_catalog

fork_group_quotas
fork_quota_windows
fork_traffic_usage

fork_config_snapshots
fork_audit_events
fork_feature_flags
```

## Storage rules

Do not store sensitive HTTPS content.

Use compact event rows.

Index real query patterns:

- timestamp;
- client;
- domain/service;
- session;
- action;
- node/inbound.

## Batch writes

Collectors should batch writes.

Do not create one DB transaction per log line.

## Aggregation

Roll raw events into:

```text
hourly/daily traffic
session summaries
service/category totals
```

Then delete expired raw events in bounded batches.

## Migration rules

Every migration must document:

- idempotence assumptions;
- transaction behavior;
- large-table impact;
- defaults/backward compatibility;
- index strategy.

## Destructive changes

Prefer staged migration:

1. add new structure;
2. dual read/write if needed;
3. backfill;
4. switch;
5. remove old in a later release.

## SQLite

Avoid long write transactions.

## PostgreSQL

Use atomic updates/transactions for shared quotas/counters.

Never use unsafe read-modify-write for shared quota.
