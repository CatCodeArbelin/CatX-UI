# CatX-UI v0.1.0

CatX-UI `0.1.0` is the first qualified stable release of the maintained
downstream fork. It is based on **MHSanaei/3x-ui v3.8.5** and bundles
**Xray 26.9.9**. The stable release is published from
[CatCodeArbelin/CatX-UI](https://github.com/CatCodeArbelin/CatX-UI) under GPLv3.

## Shipped product

- **Policy**: reusable policies, group assignment, client overrides,
  categories, schedules, temporary decisions, quarantine/managed-DNS paths,
  simulation, and explainable policy/route decisions.
- **Insight**: Activity, metadata-only destination evidence, DNS Intelligence,
  service/category classification, traffic history, retention/privacy controls,
  and heuristic risk/anomaly signals.
- **Traffic history and quota**: persistent history, per-inbound/per-node
  views, shared group accounting, fixed-window quota lifecycle, and supported
  traffic-history breakdowns.
- **Traffic Control**: capability detection, supported/degraded/unsupported
  status, Linux enforcement where the required capabilities and attribution are
  available, quota/speed/window workflows, reconciliation, and safe cleanup.
- **Risk and anomaly**: IP/session history, sharing-risk signals, ASN/country
  alerts, and DNS heuristics. These are evidence and alerts, not automatic
  bans from one signal.
- **Operations**: durable audit, signed webhooks, Prometheus metrics,
  self-service metadata/device flows, fleet visibility/update orchestration,
  backup/restore validation, and transactional updater recovery.
- **Upstream compatibility**: the 3x-ui panel, Xray integrations, protocols,
  subscriptions, clients, inbounds, routing, nodes, SQLite/PostgreSQL,
  localization, and RTL support remain part of the product boundary.

## Capability and privacy limits

Generic Xray-user attribution for kernel shaping is **unsupported**. Per-user
kernel Traffic Control must not be assumed to be active; the panel reports
capability state and refuses unsupported mutations. Optional Xray APIs, Linux
privileges, node capabilities, resolver visibility, and network topology can
produce degraded results.

CatX observes allowed metadata such as identities, timestamps, DNS/domains,
visible SNI, destination IP/port, protocol/network, inbound/outbound, node,
traffic counters, and policy decisions. It does not use TLS MITM, decrypt
HTTPS, or collect HTTP bodies, cookies, Authorization headers, credentials,
passwords, decrypted messages, or billing data. CatX is not a billing platform
or a generic plugin ecosystem.

## Upgrade and rollback

Use the CatX-owned installer or the `x-ui update` management command. Stable
installations follow the CatX stable release source; `dev-latest` is opt-in and
RC tags are for advanced testing. The update transaction performs:

```text
snapshot → checksum/identity validation → stage → backup → activate
→ migrate → start → healthcheck → commit
```

On an activation, migration, start, healthcheck, or interruption failure, the
updater attempts to restore the previous known-good binary, service files,
configuration, and database, then healthchecks the restored service. A failed
rollback remains a critical recovery state: preserve the recovery snapshot and
follow the reported recovery path rather than deleting the installation.
Back up SQLite or PostgreSQL before planned upgrades.

## Distribution

Stable release archives cover Linux `amd64`, `arm64`, `armv7`, `armv6`, `armv5`,
`s390x`, and `386`, plus Windows `amd64`. SQLite is the default database;
PostgreSQL is supported through the installer and `XUI_DB_*` settings.

The stable Docker image is
`ghcr.io/catcodearbelin/catx-ui` with aliases `v0.1.0`, `0.1.0`, and `latest`.
The published manifest contains `linux/amd64`, `linux/arm64`, `linux/arm/v7`,
`linux/arm/v6`, and `linux/386`.

## Attribution

CatX-UI is a distinct downstream fork based on
[MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). Upstream copyright,
license, and derivative-source obligations are preserved. The project does
not imply endorsement by upstream maintainers. See [LICENSE](../LICENSE) and
the [CatX README](../README.md).
