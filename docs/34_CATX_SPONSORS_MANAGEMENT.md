# CatX Sponsors Management & Product UX

## Status and branch

This is the second Sponsors work package, based on merged develop
`d324840c034c8e9dd2a02f14e1ab4ca0cecbd42b`, on
`feature/catx-sponsors-management`. It must remain a draft PR and must not
change `main`, release tags, or published artifacts.

## Source of truth

CatX uses an explicit provider mode:

- `local` (default): the CatX database is the sole authoritative sponsor
  source and CRUD is enabled;
- `remote`: the PR #9 bounded HTTPS provider is the sole authoritative source
  for rendering and is read-only from the management UI;
- no implicit merge is performed.

Remote import is deferred from this package. If it is added later, it must be a
preview/validate/confirm transaction and must never destructively overwrite
local records without explicit operator action.

## Schema and migration

The fork-owned `fork_sponsors` table stores a stable ID, enabled state, name,
priority, JSON-encoded slots, optional start/end timestamps, HTTPS destination,
validated logo URL, and JSON-encoded localized title/text maps. JSON text keeps
future locales from requiring schema changes and works on SQLite and
PostgreSQL. The model is registered through the existing fork migration-model
boundary; no upstream table is changed and no Redis or second database is
introduced.

## API and UI

Authenticated CRUD lives under `/panel/api/fork/sponsors` with a separate
management listing and status/provider endpoints so the existing active-list
contract remains stable. The CatX management page supports list/create/edit,
enable/disable, delete confirmation, scheduling, priority, supported slots,
localized content, destination/logo URLs, production-style preview, and
provider status. Only dashboard, sidebar, and page are offered; login is not a
CatX slot.

## Security and audit

All validation is server-side: bounded IDs/text/locales, HTTPS URLs without
userinfo, supported slots, priority limits, valid time windows, and PNG/JPEG/
WebP logo constraints. Remote logo fetching uses the existing SSRF-safe bounded
client and stores no image blob in the database. State-changing management
actions emit metadata-only CatX audit events when the independent audit feature
is enabled; sponsor content and URLs are not copied into audit metadata.

## Feature-off and lifecycle

Disabling Sponsors preserves local records but prevents rendering, remote
fetching/import, logo fetches, and background sync. A panel restart/reload
reconfigures the provider without reinitializing upstream database state.
Migration tests cover clean and existing SQLite/PostgreSQL databases; lifecycle
tests cover enable, restart/reload, disable, persistence, and retained data.

## Deferred scope

Remote import, local binary logo upload/storage, payments, billing, tracking,
and commercial workflows are intentionally deferred.
