# CURRENT TASK

## Work Package

`WP-8B — Self-service / Host Visibility / Fleet UI`

## Status

WP-8A is DONE: merged into `develop` at `0a6dab0f`, pushed, and its feature
branch was deleted after Fork verification `36347039861` and Release CatX-UI
`36347039863` reported green. Do not modify `main` or preview/demo data.

WP-8B is active on this feature branch. This phase is implementation-readiness
only. Do not implement production code until the readiness decision is accepted.

## Readiness decisions

- Add a fork-owned self-service authentication/session boundary that resolves to
  existing normalized client identity; do not create a second client, device,
  host, group, node, or runtime model.
- Keep self-service routes and DTOs separate from admin routes and responses.
  Enforce client ownership and group visibility server-side on every request;
  UI filtering is never authorization.
- Reuse `ClientRecord`, `ClientHwid`, `ClientGroup`, `ClientInbound`, `Host`,
  `Node`, node traffic/IP observations, and local/remote runtime status through
  narrow read/write services and adapters.
- Default visibility is the authenticated client's own profile, assigned
  devices/HWIDs, assigned inbounds/hosts, safe traffic counters, and
  capability-filtered connection status. Group/fleet visibility is explicit,
  administrator-granted, and metadata-minimized.
- Self-service mutations are limited to safe profile/device actions approved by
  the product contract. Credential/subscription reset, device reset, and other
  sensitive actions require confirmation, cooldown/rate limits, and WP-8A audit;
  no node control, inbound policy changes, host CRUD, group membership changes,
  runtime commands, or admin settings.
- Host visibility is an allowlisted projection from existing host/inbound/node
  associations. Never expose admin-only host fields, node credentials, private
  addresses, raw runtime configuration, or unrelated client data.
- Fleet UI extends the existing admin dashboard with capability-aware local and
  remote node cards, sync freshness, health, traffic, and client/group rollups.
  It uses existing runtime synchronization and must show stale/unknown state
  rather than inventing online truth.
- Tokens/sessions are opaque, revocable, short-lived where practical, scoped,
  hashed at rest, protected against fixation/replay, and never placed in URLs or
  logged. Apply CSRF/origin protections appropriate to the chosen transport,
  bounded login/device-reset attempts, and per-account/IP rate limits.
- Persisted schema changes are not presumed. Prefer existing models and a
  fork-owned namespaced settings/token/session store only if an actual gap is
  proven; any migration must cover SQLite, PostgreSQL, indexes, retention, and
  rollback.
- Use existing OpenAPI generation, frontend generated types, upstream i18n,
  and the WP-8A audit registration hook. Feature-off behavior remains upstream
  compatible.

## Required workflow

- Keep the working tree clean and use focused commits.
- Before implementation, document exact integration points, upstream-touch
  budget, threat model, API/UI contracts, and migration necessity.
- Implement only after readiness is accepted, then run relevant tests,
  `make verify`, `make verify-fork`, and the full security/concurrency matrix.
- Do not merge into `main`.

## Authorized documents

- `docs/01_ARCHITECTURE.md`
- `docs/05_TESTING_ROLLBACK_RELEASE.md`
- `docs/07_SECURITY_PRIVACY.md`
- `docs/08_FRONTEND_UX.md`
- `docs/15_DEFINITION_OF_DONE.md`
- `docs/19_REPOSITORY_MAP.md`

## Stop condition

Stop after the WP-8B readiness review and report whether Luna Medium can safely
implement it. Do not implement production code in this readiness phase.
