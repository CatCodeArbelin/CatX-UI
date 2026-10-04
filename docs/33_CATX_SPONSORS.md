# CatX Sponsors

## Status and baseline

This document describes the first post-`v0.2.0` CatX product package on
`feature/catx-sponsors`, based on synchronized `develop` at
`b4f26cf2827c74592b6b733f0a2e9f24745622d0`. It does not change the published
`v0.2.0` tag or release artifacts and is not a `v0.2.1` release plan.

## Upstream boundary review

Upstream `MHSanaei/3x-ui v3.9.0` has a Sponsors service, public `/sponsors`
data route, logo proxy, frontend page, dashboard/sidebar slots, and generated
OpenAPI types. Its production service uses a hard-coded Sanaei-hosted feed and
is therefore not a valid CatX data source. CatX keeps the upstream files
available for low-divergence compatibility, but the CatX UI path and
authenticated slots use a separate fork-owned module.

The integration points are intentionally narrow:

- `internal/forkext/settings.go` owns the authoritative feature flag;
- `internal/forkext/hooks.go` registers the module and passes runtime settings;
- `internal/forkext/sponsors/` owns configuration, fetching, filtering, cache,
  logo policy, and API routes;
- `frontend/src/forkext/registry.ts` owns the CatX route/navigation/API entry;
- existing sponsor presentation components are reused through the CatX query;
- the login page no longer requests sponsor data before authentication.

## Data and API contract

The operator configures an HTTPS JSON source URL and optional HTTPS contact URL
through authenticated panel settings. Configuration uses the existing setting
key/value mechanism, so no table, column, or runtime schema mutation is
introduced. The feature flag remains separately controlled by the authoritative
fork settings inventory.

CatX routes are under the authenticated `/panel/api` group:

- `GET /panel/api/fork/sponsors` returns a compatible sponsor list;
- `GET /panel/api/fork/sponsors/logo/:name` returns a validated active logo;
- `GET /panel/api/fork/sponsors/settings` returns sanitized configuration;
- `PUT /panel/api/fork/sponsors/settings` updates source/contact settings.

When disabled or unconfigured, the list is empty and no remote request occurs.
The contact value is operator configuration, not remote arbitrary markup.

## Remote-data safety

Only HTTPS URLs without userinfo are accepted. Redirects are revalidated. DNS
resolution and direct IP targets reject loopback, private, link-local,
unspecified, multicast, and other non-public addresses. Metadata and logo
responses have strict byte limits, timeouts, content validation, and bounded
cache/error retry behavior. Sponsor IDs, links, active windows, slot names,
localized text, and logo filenames are validated before presentation. Logo
requests are derived from the configured source host/path rather than allowing
an arbitrary remote URL from the feed.

The module stores no decrypted HTTP content, cookies, credentials,
authorization headers, or passwords. Remote metadata is held only in bounded
process memory cache; persisted state is limited to the two operator settings.

## Frontend behavior

The CatX route is `/catx/sponsors`, with existing authenticated dashboard and
sidebar slots using the same CatX API. The page preserves the upstream visual
language and RTL CSS through a thin CatX route adapter. It has explicit loading,
empty, and error states and uses existing localized sponsor strings for
English, Russian, and Persian (with locale parity retained by the existing
translation catalog). The login page does not issue a sponsor request.

## Testing and recovery

Tests cover disabled/unconfigured no-fetch behavior, URL/redirect/SSRF policy,
bounded and malformed responses, active-window/slot filtering, cache reuse and
failure fallback, logo validation/content limits, settings validation, and
authenticated route registration. Frontend tests cover the CatX route/registry,
loading, empty, and error states plus RTL-safe presentation behavior.

The recovery path is reversible without a migration: disable `sponsors.enabled`
or clear the source URL, reload fork settings, and the module returns an empty
set without contacting the remote source. A temporary source failure retains
the last known-good bounded cache for its retry window; failed updates never
replace known-good data.
