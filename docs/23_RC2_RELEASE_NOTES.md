# CatX-UI release candidate notes

## Scope

CatX-UI is a low-divergence downstream fork of `MHSanaei/3x-ui`. It keeps the
upstream panel and Xray runtime boundaries while adding independently gated
analytics, destination intelligence, policy, traffic-control, risk, audit,
webhook, portal, and fleet-update extensions.

## Installation and upgrade

Use the CatX-UI installer from the CatX release assets. The installer verifies
the same-release checksum before extraction and validates the candidate's
CatX-UI identity. Re-running the installer is an upgrade and uses the
transactional updater.

Back up the database and `/usr/local/x-ui` configuration before an operator
upgrade. SQLite and PostgreSQL installations require the supported migration
path; do not replace the database directory by hand.

## Recovery

Risky updates use snapshot → validate → stage → backup → install → migrate →
start → panel/Xray healthcheck → commit. A failure after activation restores
the previous binary, service files, configuration, and database, then reports
whether the restored service is healthy. A failed rollback is a distinct
critical state and must not be treated as an ordinary update failure.

## Feature defaults and safety

Fork features are disabled by default unless explicitly enabled in the
namespaced fork settings. Fleet mutation additionally requires the mutation
setting and explicit administrator confirmation. Dry-run never dispatches a
node update.

The portal exposes metadata-only self-service actions. Webhooks are optional,
signed, and use a local delivery queue. No decrypted HTTP bodies, cookies,
authorization headers, credentials, or passwords are collected by CatX-UI.

## Node and fleet requirements

Fleet rollout requires direct eligible nodes with fresh heartbeats, supported
update endpoints, and healthy Xray state. Canary, batch, max-parallel,
stop-on-failure, abort, restart, stale-status, and rollback evidence are
tracked durably. A timeout after POST is ambiguous until the node's correlated
`runId` evidence resolves it.

## Known limitations

- Xray remains an external runtime; optional APIs and enforcement capabilities
  are reported as unsupported instead of being simulated.
- Generic Xray per-user kernel attribution for traffic shaping is not claimed
  where the runtime cannot provide it.
- PostgreSQL and privileged Linux enforcement checks depend on the staging
  services and capabilities available to the canonical CI runner.
- No public RC should be announced until the exact final commit has passed the
  full verification and staging matrix.

## RC-3 prerelease semantics

The first public RC is v0.1.0-rc.1. It is an explicit RC-channel release,
not a stable release and not dev-latest. Stable clients continue to query only
the GitHub releases/latest stable pointer and reject prerelease metadata, so
they do not receive the RC automatically. RC clients resolve the exact
checked-in RC tag. The rolling dev channel remains opt-in and unchanged.

The tagged RC publication must set prerelease=true and latest=false. Release
metadata records the channel, exact release version/tag, publication flags,
build commit, and CatX repository identity. The normal checksum, candidate
identity, transactional snapshot/rollback, and post-restore health checks
apply unchanged to the RC path.
