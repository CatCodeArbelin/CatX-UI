# Baseline Compatibility Suite

## Baseline commit / upstream version

- Checked-out baseline commit: `4f8646729c204693a741082ddee7fe941079ea8a` (`docs: revise roadmap after repository audit`).
- Upstream base: `v3.8.5`.
- Fork state: upstream-compatible source plus documentation; no fork hooks or product features are enabled.

The checked-out repository is authoritative. This baseline does not change runtime behavior.

## Existing upstream coverage reused

The compatibility suite reuses the current upstream tests rather than creating a second test harness:

| Area | Existing coverage |
| --- | --- |
| Xray config generation and Xray-core acceptance | `internal/web/service/golden_fixtures_xray_test.go`, `xray_config_clients_test.go`, `xray_config_inject_test.go`, `xray_strip_rules_test.go`, `internal/xray/*_test.go` |
| Standard and raw subscriptions | `internal/sub/controller_test.go`, `service_test.go`, `links_test.go`, `export_all_links_test.go` |
| JSON subscriptions | `internal/sub/json_service_test.go`, `json_routing_test.go`, `json_routing_baked_test.go`, `json_dns_test.go`, `json_flow_gate_test.go` |
| Clash / Mihomo output | `internal/sub/clash_service_test.go`, `clash_yaml_test.go`, `clash_info_node_test.go`, `controller_test.go` |
| Happ output | `internal/sub/happ_test.go`, `remote_routing_test.go`, `internal/web/controller/client_happ_test.go` |
| Client CRUD and normalized source of truth | `internal/web/service/client_*.go` tests, `bulk_clients_test.go`, `bulk_traffic_test.go`, `client_identity_normalized_test.go` |
| Client groups | `client_group_bulk_test.go`, `client_group_reset_test.go`, `client_group_node_sync_test.go`, `api_scale_postgres_test.go` |
| Client-to-inbound attachment | `client_link_delta_test.go`, `client_inbound_apply_test.go`, `bulk_clients_test.go`, `client_create_fanout_test.go`, `client_flow_isolation_test.go` |
| Routing | `xray_strip_rules_test.go`, `xray_config_inject_test.go`, `internal/sub/json_routing*_test.go`, `internal/sub/remote_routing_test.go`, frontend routing tests |
| SQLite | `internal/database/*_test.go` SQLite cases, including `dump_sqlite_test.go`, `backup_test.go`, `prepare_sqlite_test.go`, and migration tests |
| PostgreSQL | Opt-in cases in `internal/database/*_test.go` and `internal/web/service/*postgres*_test.go`, including `api_scale_postgres_test.go` and `inbound_durable_postgres_test.go` |

## New compatibility coverage added

No new Go test or fixture was added. The audit found that the existing suite already covers each required baseline area, and adding duplicate characterization tests would increase maintenance and upstream divergence without increasing protection.

This document is the new baseline-suite inventory and merge-gate record. Any later fork change must add a focused regression test only where it introduces a behavior not covered by the reused tests.

## Xray config guarantees

- Existing Xray config generation remains the canonical upstream path.
- The golden fixture tests build inbound, stream, routing, balancer, and DNS fixtures through Xray-core builders.
- Client emission tests verify normalized enabled/disabled client behavior and empty-array compatibility.
- Routing tests preserve rule order and keep the required API rule; injection tests verify intentional prepend/append behavior.
- Compatibility comparisons must preserve semantic array order. Tests and future fixtures must not sort routing or outbound arrays.

## Subscription guarantees

- Standard subscription output continues to use the existing `internal/sub` engine.
- Raw/base64, JSON, Clash/Mihomo, and Happ paths are covered by their existing service/controller tests.
- Auto-detection tests cover recognized clients, precedence, configured regexes, disabled settings, and raw fallback.
- JSON tests cover legacy object and array output, routing profiles, DNS, and baked routing.
- No second subscription engine or fork-specific transformation is introduced.

## Client / group guarantees

- Normalized `ClientRecord` and client-inbound attachment rows remain the source of truth.
- Existing tests cover client create/update/delete, reattachment, duplicate handling, enable state, credentials, fan-out, and protocol-specific behavior.
- Group tests cover membership changes, reset behavior, synchronization, and group totals.
- No fork client, group, or attachment tables are introduced.

## Routing guarantees

- Existing upstream routing compilation and subscription routing paths are both covered.
- Disabled rules are removed without reordering surviving rules.
- The API routing rule is retained even when marked disabled in stored input.
- Existing outbound, balancer, and egress injection tests verify order-sensitive insertion behavior.

## SQLite coverage

SQLite is the default local test engine. Existing tests cover initialization, schema/index creation, migrations and repair paths, backup/restore, permissions, and data-copy preparation. No fork schema or migration is part of this baseline task.

## PostgreSQL coverage

PostgreSQL coverage is available through opt-in tests. Run it with a reachable PostgreSQL instance and `XUI_DB_TYPE=postgres` plus `XUI_DB_DSN`. Existing tests cover schema parity, API scale, durable commit-failure behavior, and migration paths. Without those environment variables, PostgreSQL tests skip and must not be reported as passed.

## Startup / smoke coverage

The repository’s existing startup/install smoke path is `.github/workflows/smoke.yml`, which runs `deploy/test/smoke-noninteractive.sh` for amd64 and arm64 install paths. This baseline does not create a replacement harness or alter install/startup behavior.

## Known gaps

- `make verify-fork` does not exist yet; it is intentionally not added by TASK-002.
- A real Xray process start/reload smoke requires deployment/runtime assets and is not equivalent to the unit-level Xray-core builder tests.
- PostgreSQL integration is environment-gated and cannot be claimed locally unless a configured PostgreSQL service is available.
- The existing Xray restart path does not provide automatic known-good rollback after a failed full restart; that operational gap belongs to later work and is not changed here.
- The missing `skills/test-release-guardian/SKILL.md` referenced by `CURRENT_TASK.md` was unavailable in the configured skill root; repository instructions and the mandated testing documents were followed instead.

Verification blockers observed in this environment:

- `go test ./internal/web/service ./internal/sub ./internal/database ./internal/xray/...` could not start because the repository requires Go `1.27.1`, while the installed toolchain is `go1.23.3`; automatic download of `go1.27.1` was denied by the restricted network.
- `make verify` could not start because `make` is not installed or available on `PATH`.

## Commands to run

Smallest relevant backend coverage:

```bash
go test ./internal/web/service ./internal/sub ./internal/database ./internal/xray/...
```

Full upstream verification:

```bash
make verify
```

PostgreSQL-gated coverage, when a test database is available:

```bash
XUI_DB_TYPE=postgres XUI_DB_DSN="<dsn>" go test ./internal/database ./internal/web/service
```

The repository currently has no `make verify-fork` target. Do not report that command as passed until a later task adds it. The commands above were attempted; their blockers are recorded in **Known gaps**.

## Merge gate

For this baseline task, merge requires the relevant tests to pass, `make verify` to pass when the local environment supports it, and all environment-blocked tests to be recorded with their exact blocker. The working diff must contain only this baseline documentation; no runtime, schema, updater, Xray, or subscription behavior may change.
